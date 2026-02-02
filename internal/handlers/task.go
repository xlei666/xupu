// Package handlers HTTP处理器 - 任务状态
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/pkg/orchestrator"
	"github.com/xlei/xupu/pkg/scheduler"
)

// TaskHandler 任务处理器
type TaskHandler struct{}

// NewTaskHandler 创建任务处理器
func NewTaskHandler() *TaskHandler {
	return &TaskHandler{}
}

// TaskStatusResponse 任务状态响应
type TaskStatusResponse struct {
	TaskID        string        `json:"task_id"`
	Type          string        `json:"type"`
	Status        string        `json:"status"`
	Progress      float64       `json:"progress"`
	Priority      int           `json:"priority"`
	CreatedAt     string        `json:"created_at"`
	StartedAt     *string       `json:"started_at,omitempty"`
	CompletedAt   *string       `json:"completed_at,omitempty"`
	Error         string        `json:"error,omitempty"`
	ProjectID     string        `json:"project_id,omitempty"`
}

// SchedulerStatsResponse 调度器统计响应
type SchedulerStatsResponse struct {
	TotalTasks     int32  `json:"total_tasks"`
	CompletedTasks int32  `json:"completed_tasks"`
	FailedTasks    int32  `json:"failed_tasks"`
	CancelledTasks int32  `json:"cancelled_tasks"`
	PendingTasks   int    `json:"pending_tasks"`
	ActiveWorkers  int    `json:"active_workers"`
	QueuedTasks    int    `json:"queued_tasks"`
}

// CreateAsyncProject 创建异步项目
// @Summary 创建异步项目
// @Description 异步创建AI小说创作项目，立即返回任务ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body CreateProjectRequest true "项目信息"
// @Success 200 {object} APIResponse
// @Router /api/v1/tasks/project [post]
func (h *TaskHandler) CreateAsyncProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 构建参数
	params := orchestrator.CreationParams{
		ProjectName: req.Name,
		Description: req.Description,
		WorldName:   req.Params.WorldName,
		WorldType:   req.Params.WorldType,
		WorldTheme:  req.Params.WorldTheme,
		WorldScale:  req.Params.WorldScale,
		WorldStyle:  req.Params.WorldStyle,
		StoryType:   req.Params.StoryType,
		StoryTheme:  req.Params.Theme,
		Protagonist: req.Params.Protagonist,
		StoryLength: req.Params.Length,
		ChapterCount: req.Params.ChapterCount,
		Structure:   req.Params.Structure,
		Options: orchestrator.GenerationOptions{
			SkipWorldBuild:      req.Params.Options.SkipWorldBuild,
			ExistingWorldID:     req.Params.Options.ExistingWorldID,
			SkipNarrative:       req.Params.Options.SkipNarrative,
			ExistingBlueprintID: req.Params.Options.ExistingBlueprintID,
			GenerateContent:     req.Params.Options.GenerateContent,
			StartChapter:        req.Params.Options.StartChapter,
			EndChapter:          req.Params.Options.EndChapter,
			Style:               req.Params.Options.Style,
		},
	}

	// 获取编排器用于创建任务
	orc, err := orchestrator.New()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INIT_FAILED", "初始化失败", err.Error()))
		return
	}

	// 创建异步任务
	task, err := orchestrator.CreateProjectAsync(params, orc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("CREATE_FAILED", "创建任务失败", err.Error()))
		return
	}

	c.JSON(http.StatusAccepted, successResponse(gin.H{
		"task_id": task.ID,
		"status":  "pending",
		"message": "项目创建任务已提交",
	}))
}

// GetTaskStatus 获取任务状态
// @Summary 获取任务状态
// @Description 获取指定任务的执行状态
// @Tags tasks
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/tasks/{id} [get]
func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	task, err := orchestrator.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "任务不存在", ""))
		return
	}

	c.JSON(http.StatusOK, successResponse(toTaskStatusResponse(task)))
}

// CancelTask 取消任务
// @Summary 取消任务
// @Description 取消指定任务的执行
// @Tags tasks
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/tasks/{id}/cancel [post]
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")

	if err := orchestrator.CancelTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("CANCEL_FAILED", "取消任务失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"task_id": taskID,
		"status":  "cancelled",
	}))
}

// PauseTask 暂停任务
// @Summary 暂停任务
// @Description 暂停指定任务的执行
// @Tags tasks
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/tasks/{id}/pause [post]
func (h *TaskHandler) PauseTask(c *gin.Context) {
	taskID := c.Param("id")

	if err := orchestrator.PauseTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("PAUSE_FAILED", "暂停任务失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"task_id": taskID,
		"status":  "paused",
	}))
}

// GetSchedulerStats 获取调度器统计
// @Summary 获取调度器统计
// @Description 获取调度器的统计信息
// @Tags tasks
// @Produce json
// @Success 200 {object} APIResponse
// @Router /api/v1/tasks/stats [get]
func (h *TaskHandler) GetSchedulerStats(c *gin.Context) {
	stats := orchestrator.GetSchedulerStats()
	sched := orchestrator.GetScheduler()

	response := SchedulerStatsResponse{
		TotalTasks:     stats.TotalTasks,
		CompletedTasks: stats.CompletedTasks,
		FailedTasks:    stats.FailedTasks,
		CancelledTasks: stats.CancelledTasks,
		ActiveWorkers:  sched.GetActiveWorkers(),
		QueuedTasks:    sched.GetQueueSize(),
	}

	c.JSON(http.StatusOK, successResponse(response))
}

// ListProjectTasks 列出项目的所有任务
// @Summary 列出项目的所有任务
// @Description 获取指定项目的所有任务
// @Tags tasks
// @Produce json
// @Param id path string true "项目ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/tasks/project/{id} [get]
func (h *TaskHandler) ListProjectTasks(c *gin.Context) {
	projectID := c.Param("id")

	tasks := orchestrator.GetProjectTasks(projectID)
	response := make([]TaskStatusResponse, 0, len(tasks))

	for _, task := range tasks {
		response = append(response, toTaskStatusResponse(task))
	}

	c.JSON(http.StatusOK, successResponse(response))
}

// WaitForTask 等待任务完成（SSE）
// @Summary 等待任务完成
// @Description 使用Server-Sent Events等待任务完成
// @Tags tasks
// @Produce text/event-stream
// @Param id path string true "任务ID"
// @Param timeout query int false "超时时间(秒)" default(300)
// @Router /api/v1/tasks/{id}/wait [get]
func (h *TaskHandler) WaitForTask(c *gin.Context) {
	taskID := c.Param("id")
	timeoutStr := c.DefaultQuery("timeout", "300")
	timeout, _ := strconv.Atoi(timeoutStr)

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 获取任务
	task, err := orchestrator.GetTask(taskID)
	if err != nil {
		c.SSEvent("error", gin.H{"message": "任务不存在"})
		return
	}

	// 发送初始状态
	h.sendTaskUpdate(c, task)

	// 等待任务完成或超时
	done := make(chan struct{})
	go func() {
		for {
			time.Sleep(1 * time.Second)
			task, _ := orchestrator.GetTask(taskID)
			h.sendTaskUpdate(c, task)

			status := task.GetStatus()
			if status == scheduler.StatusCompleted ||
				status == scheduler.StatusFailed ||
				status == scheduler.StatusCancelled {
				close(done)
				return
			}
		}
	}()

	select {
	case <-done:
		c.SSEvent("done", gin.H{"message": "任务完成"})
	case <-time.After(time.Duration(timeout) * time.Second):
		c.SSEvent("timeout", gin.H{"message": "等待超时"})
	case <-c.Request.Context().Done():
		return
	}
}

// sendTaskUpdate 发送任务更新
func (h *TaskHandler) sendTaskUpdate(c *gin.Context, task *scheduler.Task) {
	c.SSEvent("update", gin.H{
		"task_id":   task.ID,
		"status":    string(task.GetStatus()),
		"progress":  task.GetProgress(),
		"error":     task.Error,
	})
}

// toTaskStatusResponse 转换任务状态响应
func toTaskStatusResponse(task *scheduler.Task) TaskStatusResponse {
	response := TaskStatusResponse{
		TaskID:    task.ID,
		Type:      string(task.Type),
		Status:    string(task.GetStatus()),
		Progress:  task.GetProgress(),
		Priority:  int(task.Priority),
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
		Error:     task.Error,
		ProjectID: task.ProjectID,
	}

	if task.StartedAt != nil {
		s := task.StartedAt.Format(time.RFC3339)
		response.StartedAt = &s
	}

	if task.CompletedAt != nil {
		s := task.CompletedAt.Format(time.RFC3339)
		response.CompletedAt = &s
	}

	return response
}
