// Package orchestrator 编排器 - 异步任务支持
package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/scheduler"
	"github.com/xlei/xupu/pkg/writer"
)

// SchedulerHolder 调度器持有者接口
type SchedulerHolder interface {
	GetScheduler() *scheduler.Scheduler
	SetScheduler(*scheduler.Scheduler)
}

// initScheduler 初始化全局调度器
var (
	globalScheduler *scheduler.Scheduler
	schedulerOnce   schedulerInitializer
)

type schedulerInitializer struct {
	initialized bool
}

// InitScheduler 初始化全局调度器
func InitScheduler() error {
	if schedulerOnce.initialized {
		return nil
	}

	cfg := &scheduler.Config{
		WorkerCount:   3,
		QueueSize:     1000,
		CheckInterval: 100 * time.Millisecond,
	}

	globalScheduler = scheduler.New(cfg)

	// 设置回调
	globalScheduler.SetTaskCompleteCallback(func(task *scheduler.Task) {
		log.Printf("[Orchestrator] Task completed: %s", task.ID)
		onTaskComplete(task)
	})

	globalScheduler.SetTaskFailedCallback(func(task *scheduler.Task, err error) {
		log.Printf("[Orchestrator] Task failed: %s, error: %v", task.ID, err)
		onTaskFailed(task, err)
	})

	if err := globalScheduler.Start(); err != nil {
		return fmt.Errorf("启动调度器失败: %w", err)
	}

	schedulerOnce.initialized = true
	log.Println("[Orchestrator] Scheduler initialized")
	return nil
}

// GetScheduler 获取全局调度器
func GetScheduler() *scheduler.Scheduler {
	return globalScheduler
}

// StopScheduler 停止全局调度器
func StopScheduler() {
	if globalScheduler != nil {
		globalScheduler.Stop()
		schedulerOnce.initialized = false
		log.Println("[Orchestrator] Scheduler stopped")
	}
}

// onTaskComplete 任务完成处理
func onTaskComplete(task *scheduler.Task) {
	// 更新项目状态
	if task.ProjectID != "" {
		// 这里可以添加额外的完成处理逻辑
	}
}

// onTaskFailed 任务失败处理
func onTaskFailed(task *scheduler.Task, err error) {
	// 更新项目状态为失败
	if task.ProjectID != "" {
		// 这里可以添加额外的失败处理逻辑
	}
}

// ============================================
// 异步项目创建
// ============================================

// CreateProjectAsync 异步创建项目
func CreateProjectAsync(params CreationParams, orc *Orchestrator) (*scheduler.Task, error) {
	if globalScheduler == nil {
		return nil, fmt.Errorf("调度器未初始化")
	}

	// 创建任务
	task := scheduler.NewJob(
		scheduler.TaskTypeWorldBuild, // 实际上这是一个完整的创作流程
		"", // projectID will be set after creation
		params,
		func(ctx context.Context, t *scheduler.Task) error {
			return executeProjectCreation(ctx, t, orc)
		},
	).SetPriority(scheduler.PriorityNormal).
	 SetScheduler(globalScheduler).
	 Build()

	// 提交任务
	if err := globalScheduler.Submit(task); err != nil {
		return nil, err
	}

	return task, nil
}

// executeProjectCreation 执行项目创建
func executeProjectCreation(ctx context.Context, task *scheduler.Task, orc *Orchestrator) error {
	params := task.Params.(CreationParams)

	// 创建项目对象
	project := &models.Project{
		ID:        task.ProjectID,
		Name:      params.ProjectName,
		Description: params.Description,
		UserID:    params.UserID,
		Mode:      models.ModePlanning,
		Status:    models.StatusBuilding,
		Progress:  0,
	}

	// 保存项目
	if err := orc.db.SaveProject(project); err != nil {
		return fmt.Errorf("保存项目失败: %w", err)
	}

	// 更新任务的项目ID
	task.ProjectID = project.ID

	// 检查取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 执行创作流程
	result, err := orc.executeCreationFlowAsync(project, params, ctx)
	if err != nil {
		orc.db.UpdateProjectStatus(project.ID, models.StatusFailed, project.Progress)
		return fmt.Errorf("执行创作流程失败: %w", err)
	}

	// 更新项目
	project.WorldID = result.WorldID
	project.NarrativeID = result.NarrativeID
	project.Progress = 100
	project.Status = models.StatusCompleted
	orc.db.SaveProject(project)

	// 设置任务结果
	task.SetResult(result)

	return nil
}

// executeCreationFlowAsync 异步执行创作流程
func (o *Orchestrator) executeCreationFlowAsync(project *models.Project, params CreationParams, ctx context.Context) (*CreationResult, error) {
	result := &CreationResult{}
	progressStep := 100.0 / 3 // 三个阶段

	// 阶段1: 世界设定
	log.Printf("[编排器] 开始世界设定，项目ID: %s", project.ID)
	project.Progress = progressStep
	o.db.SaveProject(project)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	worldID, err := o.stage1_WorldBuilding(params, result)
	if err != nil {
		return nil, fmt.Errorf("世界设定阶段失败: %w", err)
	}
	result.WorldID = worldID
	project.WorldID = worldID

	// 阶段2: 叙事蓝图
	log.Printf("[编排器] 开始叙事规划，项目ID: %s", project.ID)
	project.Progress = progressStep * 2
	o.db.SaveProject(project)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	narrativeID, err := o.stage2_NarrativePlanning(worldID, params, result)
	if err != nil {
		return nil, fmt.Errorf("叙事规划阶段失败: %w", err)
	}
	result.NarrativeID = narrativeID
	project.NarrativeID = narrativeID

	// 阶段3: 内容生成（如果需要）
	if params.Options.GenerateContent {
		project.Status = models.StatusGenerating
		o.db.SaveProject(project)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		sceneCount, wordCount, err := o.stage3_ContentGenerationAsync(narrativeID, params, result, ctx)
		if err != nil {
			return nil, fmt.Errorf("内容生成阶段失败: %w", err)
		}
		result.SceneCount = sceneCount
		result.WordCount = wordCount
	}

	return result, nil
}

// stage3_ContentGenerationAsync 异步内容生成
func (o *Orchestrator) stage3_ContentGenerationAsync(narrativeID string, params CreationParams, result *CreationResult, ctx context.Context) (int, int, error) {
	blueprint, err := o.db.GetNarrativeBlueprint(narrativeID)
	if err != nil {
		return 0, 0, fmt.Errorf("获取叙事蓝图失败: %w", err)
	}

	world, err := o.db.GetWorld(blueprint.WorldID)
	if err != nil {
		return 0, 0, fmt.Errorf("获取世界设定失败: %w", err)
	}

	startChapter := params.Options.StartChapter
	endChapter := params.Options.EndChapter
	if endChapter == 0 || endChapter > len(blueprint.ChapterPlans) {
		endChapter = len(blueprint.ChapterPlans)
	}

	sceneCount := 0
	totalWordCount := 0

	for i := startChapter - 1; i < endChapter; i++ {
		select {
		case <-ctx.Done():
			return sceneCount, totalWordCount, ctx.Err()
		default:
		}

		chapter := blueprint.ChapterPlans[i]
		log.Printf("[编排器] 生成第%d章: %s", chapter.Chapter, chapter.Title)

		chapterScenes := getScenesForChapter(blueprint.Scenes, chapter.Chapter)

		for _, sceneInstr := range chapterScenes {
			sceneResult, err := o.writer.GenerateScene(writer.GenerateParams{
				BlueprintID:      blueprint.ID,
				Chapter:          sceneInstr.Chapter,
				Scene:            sceneInstr.Scene,
				Instruction:      &sceneInstr,
				PreviousSummary:  buildPreviousSummary(blueprint.ChapterPlans[:i]),
				CharacterStates:  buildCharacterStates(blueprint, world),
				WorldContext:     world,
				Style:            writer.DefaultStyle(),
			})

			if err != nil {
				log.Printf("[编排器] 警告: 场景%d-%d生成失败: %v", sceneInstr.Chapter, sceneInstr.Scene, err)
				continue
			}

			sceneCount++
			totalWordCount += sceneResult.WordCount
		}
	}

	return sceneCount, totalWordCount, nil
}

// ============================================
// 任务管理API
// ============================================

// GetTask 获取任务状态
func GetTask(taskID string) (*scheduler.Task, error) {
	if globalScheduler == nil {
		return nil, fmt.Errorf("调度器未初始化")
	}

	task, exists := globalScheduler.GetTask(taskID)
	if !exists {
		return nil, fmt.Errorf("任务不存在")
	}

	return task, nil
}

// CancelTask 取消任务
func CancelTask(taskID string) error {
	if globalScheduler == nil {
		return fmt.Errorf("调度器未初始化")
	}

	return globalScheduler.CancelTask(taskID)
}

// PauseTask 暂停任务
func PauseTask(taskID string) error {
	if globalScheduler == nil {
		return fmt.Errorf("调度器未初始化")
	}

	return globalScheduler.PauseTask(taskID)
}

// GetProjectTasks 获取项目的所有任务
func GetProjectTasks(projectID string) []*scheduler.Task {
	if globalScheduler == nil {
		return nil
	}

	return globalScheduler.GetProjectTasks(projectID)
}

// GetSchedulerStats 获取调度器统计信息
func GetSchedulerStats() scheduler.Stats {
	if globalScheduler == nil {
		return scheduler.Stats{}
	}

	return globalScheduler.GetStats()
}
