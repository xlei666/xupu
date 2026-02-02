// Package scheduler 调度器 - 优先级队列和工作协程
package scheduler

import (
	"container/heap"
	"log"
	"sync/atomic"
	"time"
)

// PriorityQueue 优先级队列（基于堆实现）
type PriorityQueue struct {
	tasks []*Task
}

// NewPriorityQueue 创建优先级队列
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		tasks: make([]*Task, 0),
	}
	heap.Init(pq)
	return pq
}

// Len 实现heap.Interface
func (pq *PriorityQueue) Len() int {
	return len(pq.tasks)
}

// Less 实现heap.Interface - 优先级高的在前面
func (pq *PriorityQueue) Less(i, j int) bool {
	// 首先比较优先级
	if pq.tasks[i].Priority != pq.tasks[j].Priority {
		return pq.tasks[i].Priority > pq.tasks[j].Priority
	}
	// 优先级相同时，先创建的先执行
	return pq.tasks[i].CreatedAt.Before(pq.tasks[j].CreatedAt)
}

// Swap 实现heap.Interface
func (pq *PriorityQueue) Swap(i, j int) {
	pq.tasks[i], pq.tasks[j] = pq.tasks[j], pq.tasks[i]
}

// Push 实现heap.Interface
func (pq *PriorityQueue) Push(x interface{}) {
	pq.tasks = append(pq.tasks, x.(*Task))
}

// Pop 实现heap.Interface
func (pq *PriorityQueue) Pop() interface{} {
	old := pq.tasks
	n := len(old)
	task := old[n-1]
	pq.tasks = old[0 : n-1]
	return task
}

// ============================================
// Worker 工作协程
// ============================================

// Worker 工作协程
type Worker struct {
	id         int
	scheduler  *Scheduler
	taskChan   chan *Task
	running    atomic.Bool
}

// NewWorker 创建工作协程
func NewWorker(id int, scheduler *Scheduler) *Worker {
	return &Worker{
		id:        id,
		scheduler: scheduler,
		taskChan:  make(chan *Task, 1),
	}
}

// Run 运行工作协程
func (w *Worker) Run() {
	defer w.scheduler.wg.Done()
	w.running.Store(true)

	log.Printf("[Worker-%d] Started", w.id)

	for {
		select {
		case <-w.scheduler.ctx.Done():
			w.running.Store(false)
			log.Printf("[Worker-%d] Stopped", w.id)
			return

		case task := <-w.taskChan:
			w.executeTask(task)
		}
	}
}

// TrySubmit 尝试提交任务
func (w *Worker) TrySubmit(task *Task) bool {
	select {
	case w.taskChan <- task:
		return true
	default:
		return false
	}
}

// executeTask 执行任务
func (w *Worker) executeTask(task *Task) {
	log.Printf("[Worker-%d] Executing: %s", w.id, task)

	// 设置任务状态
	task.SetStatus(StatusRunning)

	// 执行任务
	err := task.Executor(task.Context(), task)

	// 处理结果
	if err != nil {
		if task.IsCancelled() {
			task.SetStatus(StatusCancelled)
		} else {
			w.scheduler.markTaskFailed(task, err)
		}
	} else {
		w.scheduler.markTaskComplete(task)
	}
}

// IsRunning 检查是否运行中
func (w *Worker) IsRunning() bool {
	return w.running.Load()
}

// ============================================
// Job 任务构建器
// ============================================

// Job 任务构建器
type Job struct {
	task        *Task
	scheduler   *Scheduler
}

// NewJob 创建任务构建器
func NewJob(taskType TaskType, projectID string, params interface{}, executor TaskExecutor) *Job {
	task := NewTask(taskType, projectID, params, executor)
	return &Job{task: task}
}

// SetPriority 设置优先级
func (j *Job) SetPriority(priority TaskPriority) *Job {
	j.task.SetPriority(priority)
	return j
}

// SetScheduler 设置调度器
func (j *Job) SetScheduler(scheduler *Scheduler) *Job {
	j.scheduler = scheduler
	return j
}

// Build 构建任务
func (j *Job) Build() *Task {
	return j.task
}

// Submit 提交到调度器
func (j *Job) Submit() error {
	if j.scheduler == nil {
		return ErrNoScheduler
	}
	return j.scheduler.Submit(j.task)
}

// ============================================
// 错误定义
// ============================================

var (
	ErrNoScheduler = &SchedulerError{Code: "NO_SCHEDULER", Message: "no scheduler assigned"}
	ErrTaskNotFound = &SchedulerError{Code: "TASK_NOT_FOUND", Message: "task not found"}
	ErrSchedulerStopped = &SchedulerError{Code: "SCHEDULER_STOPPED", Message: "scheduler is stopped"}
)

// SchedulerError 调度器错误
type SchedulerError struct {
	Code    string
	Message string
}

func (e *SchedulerError) Error() string {
	return e.Code + ": " + e.Message
}

// ============================================
// 任务进度跟踪器
// ============================================

// ProgressTracker 进度跟踪器
type ProgressTracker struct {
	task       *Task
	total      int
	current    int
	lastUpdate time.Time
}

// NewProgressTracker 创建进度跟踪器
func NewProgressTracker(task *Task, total int) *ProgressTracker {
	return &ProgressTracker{
		task:       task,
		total:      total,
		current:    0,
		lastUpdate: time.Now(),
	}
}

// Increment 增加进度
func (pt *ProgressTracker) Increment() {
	pt.current++
	pt.update()
}

// Add 增加指定数量
func (pt *ProgressTracker) Add(n int) {
	pt.current += n
	pt.update()
}

// Set 设置当前进度
func (pt *ProgressTracker) Set(n int) {
	pt.current = n
	pt.update()
}

// update 更新任务进度
func (pt *ProgressTracker) update() {
	if pt.total > 0 {
		progress := float64(pt.current) / float64(pt.total) * 100
		pt.task.SetProgress(progress)
	}
}

// Complete 标记完成
func (pt *ProgressTracker) Complete() {
	pt.current = pt.total
	pt.task.SetProgress(100)
}
