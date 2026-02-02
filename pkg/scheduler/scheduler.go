// Package scheduler 调度器 - 异步任务调度和管理
package scheduler

import (
	"container/heap"
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Scheduler 调度器
type Scheduler struct {
	// 任务队列
	taskQueue     *PriorityQueue
	tasks         map[string]*Task
	taskMutex     sync.RWMutex

	// 工作池
	workers       []*Worker
	workerCount   int
	activeWorkers int32

	// 控制
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup

	// 统计
	stats         Stats
	statsMutex    sync.RWMutex

	// 回调
	onTaskComplete func(*Task)
	onTaskFailed  func(*Task, error)
}

// Stats 统计信息
type Stats struct {
	TotalTasks     int32 `json:"total_tasks"`
	CompletedTasks int32 `json:"completed_tasks"`
	FailedTasks    int32 `json:"failed_tasks"`
	CancelledTasks int32 `json:"cancelled_tasks"`
}

// Config 调度器配置
type Config struct {
	WorkerCount    int           // 工作协程数量
	QueueSize      int           // 队列大小
	CheckInterval  time.Duration // 检查间隔
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		WorkerCount:   3,
		QueueSize:     1000,
		CheckInterval: 100 * time.Millisecond,
	}
}

// New 创建调度器
func New(cfg *Config) *Scheduler {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		taskQueue:     NewPriorityQueue(),
		tasks:         make(map[string]*Task),
		workers:       make([]*Worker, 0, cfg.WorkerCount),
		workerCount:   cfg.WorkerCount,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	log.Printf("[Scheduler] Starting with %d workers", s.workerCount)

	// 启动工作协程
	for i := 0; i < s.workerCount; i++ {
		worker := NewWorker(i, s)
		s.workers = append(s.workers, worker)
		s.wg.Add(1)
		go worker.Run()
	}

	// 启动调度循环
	s.wg.Add(1)
	go s.scheduleLoop()

	log.Println("[Scheduler] Started successfully")
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	log.Println("[Scheduler] Stopping...")

	s.cancel()

	// 等待所有工作协程完成
	s.wg.Wait()

	log.Println("[Scheduler] Stopped")
}

// Submit 提交任务
func (s *Scheduler) Submit(task *Task) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	if s.ctx.Err() != nil {
		return fmt.Errorf("scheduler is stopped")
	}

	// 检查任务是否已存在
	if _, exists := s.tasks[task.ID]; exists {
		return fmt.Errorf("task %s already exists", task.ID)
	}

	// 添加到队列
	heap.Push(s.taskQueue, task)
	s.tasks[task.ID] = task

	// 更新统计
	atomic.AddInt32(&s.stats.TotalTasks, 1)

	log.Printf("[Scheduler] Task submitted: %s", task)
	return nil
}

// GetTask 获取任务
func (s *Scheduler) GetTask(id string) (*Task, bool) {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	task, exists := s.tasks[id]
	return task, exists
}

// CancelTask 取消任务
func (s *Scheduler) CancelTask(id string) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("task %s not found", id)
	}

	task.Cancel()

	// 更新统计
	if task.Status == StatusCancelled {
		atomic.AddInt32(&s.stats.CancelledTasks, 1)
	}

	log.Printf("[Scheduler] Task cancelled: %s", task)
	return nil
}

// PauseTask 暂停任务（实际上是取消，稍后可重新提交）
func (s *Scheduler) PauseTask(id string) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("task %s not found", id)
	}

	task.SetStatus(StatusPaused)
	task.Cancel()

	log.Printf("[Scheduler] Task paused: %s", task)
	return nil
}

// GetProjectTasks 获取项目的所有任务
func (s *Scheduler) GetProjectTasks(projectID string) []*Task {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	var tasks []*Task
	for _, task := range s.tasks {
		if task.ProjectID == projectID {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// GetStats 获取统计信息
func (s *Scheduler) GetStats() Stats {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	return Stats{
		TotalTasks:     atomic.LoadInt32(&s.stats.TotalTasks),
		CompletedTasks: atomic.LoadInt32(&s.stats.CompletedTasks),
		FailedTasks:    atomic.LoadInt32(&s.stats.FailedTasks),
		CancelledTasks: atomic.LoadInt32(&s.stats.CancelledTasks),
	}
}

// SetTaskCompleteCallback 设置任务完成回调
func (s *Scheduler) SetTaskCompleteCallback(fn func(*Task)) {
	s.onTaskComplete = fn
}

// SetTaskFailedCallback 设置任务失败回调
func (s *Scheduler) SetTaskFailedCallback(fn func(*Task, error)) {
	s.onTaskFailed = fn
}

// scheduleLoop 调度循环
func (s *Scheduler) scheduleLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return

		case <-ticker.C:
			s.dispatchTasks()
		}
	}
}

// dispatchTasks 分发任务
func (s *Scheduler) dispatchTasks() {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	// 检查是否有空闲工作协程
	active := atomic.LoadInt32(&s.activeWorkers)
	if int(active) >= s.workerCount {
		return
	}

	// 从队列中取出任务
	for s.taskQueue.Len() > 0 && int(active) < s.workerCount {
		task := heap.Pop(s.taskQueue).(*Task)

		// 检查任务状态
		if task.Status != StatusPending {
			continue
		}

		// 检查任务是否被取消
		if task.IsCancelled() {
			task.SetStatus(StatusCancelled)
			atomic.AddInt32(&s.stats.CancelledTasks, 1)
			continue
		}

		// 将任务放入工作通道
		for _, worker := range s.workers {
			if worker.TrySubmit(task) {
				atomic.AddInt32(&s.activeWorkers, 1)
				active++
				break
			}
		}
	}
}

// markTaskComplete 标记任务完成
func (s *Scheduler) markTaskComplete(task *Task) {
	atomic.AddInt32(&s.activeWorkers, -1)
	atomic.AddInt32(&s.stats.CompletedTasks, 1)

	task.SetStatus(StatusCompleted)

	if s.onTaskComplete != nil {
		s.onTaskComplete(task)
	}

	log.Printf("[Scheduler] Task completed: %s", task)
}

// markTaskFailed 标记任务失败
func (s *Scheduler) markTaskFailed(task *Task, err error) {
	atomic.AddInt32(&s.activeWorkers, -1)
	atomic.AddInt32(&s.stats.FailedTasks, 1)

	task.SetStatus(StatusFailed)
	task.SetError(err)

	if s.onTaskFailed != nil {
		s.onTaskFailed(task, err)
	}

	log.Printf("[Scheduler] Task failed: %s, error: %v", task, err)
}

// GetQueueSize 获取队列大小
func (s *Scheduler) GetQueueSize() int {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()
	return s.taskQueue.Len()
}

// GetActiveWorkers 获取活跃工作协程数
func (s *Scheduler) GetActiveWorkers() int {
	return int(atomic.LoadInt32(&s.activeWorkers))
}

// CleanCompletedTasks 清理已完成的任务
func (s *Scheduler) CleanCompletedTasks(olderThan time.Duration) int {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	cutoff := time.Now().Add(-olderThan)
	count := 0

	for id, task := range s.tasks {
		if task.CompletedAt != nil && task.CompletedAt.Before(cutoff) {
			delete(s.tasks, id)
			count++
		}
	}

	return count
}
