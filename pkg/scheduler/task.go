// Package scheduler 调度器 - 异步任务调度和管理
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"   // 等待执行
	StatusRunning   TaskStatus = "running"   // 执行中
	StatusCompleted TaskStatus = "completed" // 已完成
	StatusFailed    TaskStatus = "failed"    // 失败
	StatusCancelled TaskStatus = "cancelled" // 已取消
	StatusPaused    TaskStatus = "paused"    // 已暂停
)

// TaskPriority 任务优先级
type TaskPriority int

const (
	PriorityLow    TaskPriority = 1
	PriorityNormal TaskPriority = 5
	PriorityHigh   TaskPriority = 10
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeWorldBuild     TaskType = "world_build"      // 世界构建
	TaskTypeNarrativePlan  TaskType = "narrative_plan"   // 叙事规划
	TaskTypeChapterGen     TaskType = "chapter_gen"      // 章节生成
	TaskTypeSceneGen       TaskType = "scene_gen"        // 场景生成
	TaskTypeExport         TaskType = "export"           // 导出
)

// Task 任务
type Task struct {
	ID            string        `json:"id"`
	Type          TaskType      `json:"type"`
	Priority      TaskPriority  `json:"priority"`
	Status        TaskStatus    `json:"status"`
	Progress      float64       `json:"progress"`
	CreatedAt     time.Time     `json:"created_at"`
	StartedAt     *time.Time    `json:"started_at,omitempty"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
	Error         string        `json:"error,omitempty"`

	// 任务参数
	ProjectID     string        `json:"project_id"`
	Params        interface{}   `json:"params"`

	// 执行函数
	Executor      TaskExecutor  `json:"-"`

	// 上下文控制
	ctx           context.Context `json:"-"`
	cancel        context.CancelFunc `json:"-"`

	// 结果
	Result        interface{}   `json:"result,omitempty"`

	mu            sync.RWMutex `json:"-"`
}

// TaskExecutor 任务执行器
type TaskExecutor func(ctx context.Context, task *Task) error

// NewTask 创建新任务
func NewTask(taskType TaskType, projectID string, params interface{}, executor TaskExecutor) *Task {
	ctx, cancel := context.WithCancel(context.Background())
	return &Task{
		ID:        uuid.New().String(),
		Type:      taskType,
		Priority:  PriorityNormal,
		Status:    StatusPending,
		Progress:  0,
		CreatedAt: time.Now(),
		ProjectID: projectID,
		Params:    params,
		Executor:  executor,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// SetPriority 设置优先级
func (t *Task) SetPriority(priority TaskPriority) *Task {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Priority = priority
	return t
}

// GetStatus 获取状态
func (t *Task) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status
}

// SetStatus 设置状态
func (t *Task) SetStatus(status TaskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	t.Status = status

	if status == StatusRunning && t.StartedAt == nil {
		t.StartedAt = &now
	}
	if status == StatusCompleted || status == StatusFailed || status == StatusCancelled {
		t.CompletedAt = &now
	}
}

// GetProgress 获取进度
func (t *Task) GetProgress() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Progress
}

// SetProgress 设置进度
func (t *Task) SetProgress(progress float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Progress = progress
}

// IncrementProgress 增加进度
func (t *Task) IncrementProgress(delta float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Progress = min(100, t.Progress+delta)
}

// Cancel 取消任务
func (t *Task) Cancel() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancel != nil {
		t.cancel()
	}

	if t.Status == StatusPending || t.Status == StatusRunning {
		t.Status = StatusCancelled
		now := time.Now()
		t.CompletedAt = &now
	}
}

// IsCancelled 检查是否已取消
func (t *Task) IsCancelled() bool {
	select {
	case <-t.ctx.Done():
		return true
	default:
		return false
	}
}

// Context 获取任务上下文
func (t *Task) Context() context.Context {
	return t.ctx
}

// SetError 设置错误
func (t *Task) SetError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err != nil {
		t.Error = err.Error()
	}
}

// GetResult 获取结果
func (t *Task) GetResult() interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Result
}

// SetResult 设置结果
func (t *Task) SetResult(result interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Result = result
}

// String 返回任务字符串表示
func (t *Task) String() string {
	return fmt.Sprintf("Task[%s:%s] status=%s progress=%.1f%%",
		t.Type, t.ID[:8], t.Status, t.Progress)
}

// min 返回最小值
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
