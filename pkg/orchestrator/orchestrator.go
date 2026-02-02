// Package orchestrator 编排器 - 系统的核心协调者
// 负责串联各个模块，完成从世界设定到小说文本的端到端生成
package orchestrator

import (
	"fmt"
	"log"
	"time"

	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/config"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/narrative"
	"github.com/xlei/xupu/pkg/writer"
	"github.com/xlei/xupu/pkg/worldbuilder"
)

// CreationParams 创作参数
type CreationParams struct {
	// 项目信息
	ProjectName string `json:"project_name"`
	Description string `json:"description"`
	UserID      string `json:"user_id,omitempty"`

	// 世界设定参数
	WorldName   string `json:"world_name"`
	WorldType   string `json:"world_type"`
	WorldTheme  string `json:"world_theme"`
	WorldScale  string `json:"world_scale"`
	WorldStyle  string `json:"world_style"`

	// 故事参数
	StoryType    string `json:"story_type"`
	StoryTheme   string `json:"story_theme"`
	Protagonist  string `json:"protagonist"`
	StoryLength  string `json:"story_length"`
	ChapterCount int  `json:"chapter_count,omitempty"`
	Structure    string `json:"structure,omitempty"`

	// 生成选项
	Options GenerationOptions `json:"options"`
}

// GenerationOptions 生成选项
type GenerationOptions struct {
	SkipWorldBuild   bool `json:"skip_world_build"`    // 跳过世界构建（使用已有）
	ExistingWorldID  string `json:"existing_world_id"`  // 已有世界ID
	SkipNarrative    bool `json:"skip_narrative"`       // 跳过叙事规划
	ExistingBlueprintID string `json:"existing_blueprint_id"` // 已有蓝图ID
	GenerateContent  bool `json:"generate_content"`      // 是否生成文本内容
	StartChapter     int  `json:"start_chapter"`         // 起始章节
	EndChapter       int  `json:"end_chapter"`           // 结束章节
	Style            string `json:"style"`                // 写作风格
}

// Orchestrator 编排器
type Orchestrator struct {
	db              db.Database
	cfg             *config.Config
	worldBuilder    *worldbuilder.WorldBuilder
	narrativeEngine *narrative.NarrativeEngine
	writer          *writer.Writer
}

// New 创建编排器
func New() (*Orchestrator, error) {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 初始化各模块
	worldBuilder, err := worldbuilder.New()
	if err != nil {
		return nil, fmt.Errorf("初始化世界设定器失败: %w", err)
	}

	narrativeEngine, err := narrative.New()
	if err != nil {
		return nil, fmt.Errorf("初始化叙事器失败: %w", err)
	}

	writer, err := writer.New()
	if err != nil {
		return nil, fmt.Errorf("初始化写作器失败: %w", err)
	}

	return &Orchestrator{
		db:              db.Get(),
		cfg:             cfg,
		worldBuilder:    worldBuilder,
		narrativeEngine: narrativeEngine,
		writer:          writer,
	}, nil
}

// CreateProject 创建新项目并执行完整的创作流程
func (o *Orchestrator) CreateProject(params CreationParams) (*models.Project, error) {
	// 1. 创建项目对象
	project := &models.Project{
		ID:          db.GenerateID("project"),
		Name:        params.ProjectName,
		Description: params.Description,
		UserID:      params.UserID,
		Mode:        models.ModePlanning,
		Status:      models.StatusBuilding,
		Progress:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存项目
	if err := o.db.SaveProject(project); err != nil {
		return nil, fmt.Errorf("保存项目失败: %w", err)
	}

	// 2. 执行创作流程
	result, err := o.executeCreationFlow(project, params)
	if err != nil {
		// 更新项目状态为失败
		o.db.UpdateProjectStatus(project.ID, models.StatusFailed, project.Progress)
		return nil, fmt.Errorf("执行创作流程失败: %w", err)
	}

	// 3. 更新项目关联
	project.WorldID = result.WorldID
	project.NarrativeID = result.NarrativeID
	project.Progress = 100
	project.Status = models.StatusCompleted
	project.UpdatedAt = time.Now()

	if err := o.db.SaveProject(project); err != nil {
		return nil, fmt.Errorf("更新项目失败: %w", err)
	}

	return project, nil
}

// CreationResult 创作结果
type CreationResult struct {
	WorldID     string `json:"world_id"`
	NarrativeID string `json:"narrative_id"`
	SceneCount  int    `json:"scene_count"`
	WordCount   int    `json:"word_count"`
	Duration    time.Duration `json:"duration"`
}

// executeCreationFlow 执行创作流程
func (o *Orchestrator) executeCreationFlow(project *models.Project, params CreationParams) (*CreationResult, error) {
	startTime := time.Now()
	result := &CreationResult{}

	log.Printf("[编排器] 开始执行创作流程，项目ID: %s", project.ID)

	// 阶段1: 世界设定
	worldID, err := o.stage1_WorldBuilding(params, result)
	if err != nil {
		return nil, fmt.Errorf("世界设定阶段失败: %w", err)
	}
	result.WorldID = worldID
	project.WorldID = worldID
	o.db.SaveProject(project) // 更新进度
	log.Printf("[编排器] 世界设定完成，ID: %s", worldID)

	// 阶段2: 叙事蓝图
	narrativeID, err := o.stage2_NarrativePlanning(worldID, params, result)
	if err != nil {
		return nil, fmt.Errorf("叙事规划阶段失败: %w", err)
	}
	result.NarrativeID = narrativeID
	project.NarrativeID = narrativeID
	o.db.SaveProject(project) // 更新进度
	log.Printf("[编排器] 叙事蓝图完成，ID: %s", narrativeID)

	// 阶段3: 内容生成（如果需要）
	if params.Options.GenerateContent {
		sceneCount, wordCount, err := o.stage3_ContentGeneration(narrativeID, params, result)
		if err != nil {
			return nil, fmt.Errorf("内容生成阶段失败: %w", err)
		}
		result.SceneCount = sceneCount
		result.WordCount = wordCount
		log.Printf("[编排器] 内容生成完成，场景数: %d, 字数: %d", sceneCount, wordCount)
	}

	result.Duration = time.Since(startTime)
	log.Printf("[编排器] 创作流程完成，耗时: %v", result.Duration)

	return result, nil
}

// stage1_WorldBuilding 阶段1: 世界设定
func (o *Orchestrator) stage1_WorldBuilding(params CreationParams, result *CreationResult) (string, error) {
	// 如果指定了已有世界，直接使用
	if params.Options.ExistingWorldID != "" {
		world, err := o.db.GetWorld(params.Options.ExistingWorldID)
		if err != nil {
			return "", fmt.Errorf("获取指定世界失败: %w", err)
		}
		log.Printf("[编排器] 使用已有世界: %s", world.Name)
		return world.ID, nil
	}

	// 构建新世界
	world, err := o.worldBuilder.Build(worldbuilder.BuildParams{
		Name:      params.WorldName,
		Type:      parseWorldType(params.WorldType),
		Scale:     parseWorldScale(params.WorldScale),
		Theme:     params.WorldTheme,
		Style:     params.WorldStyle,
	})

	if err != nil {
		return "", err
	}

	return world.ID, nil
}

// stage2_NarrativePlanning 阶段2: 叙事规划
func (o *Orchestrator) stage2_NarrativePlanning(worldID string, params CreationParams, result *CreationResult) (string, error) {
	// 如果指定了已有蓝图，直接使用
	if params.Options.ExistingBlueprintID != "" {
		blueprint, err := o.db.GetNarrativeBlueprint(params.Options.ExistingBlueprintID)
		if err != nil {
			return "", fmt.Errorf("获取指定蓝图失败: %w", err)
		}
		log.Printf("[编排器] 使用已有蓝图: %s", blueprint.ID)
		return blueprint.ID, nil
	}

	// 创建新蓝图
	narrativeParams := narrative.CreateParams{
		WorldID:      worldID,
		StoryType:    params.StoryType,
		Theme:        params.StoryTheme,
		Protagonist:  params.Protagonist,
		Length:       params.StoryLength,
		ChapterCount: params.ChapterCount,
		Structure:    parseNarrativeStructure(params.Structure),
	}

	blueprint, err := o.narrativeEngine.CreateBlueprint(narrativeParams)
	if err != nil {
		return "", err
	}

	return blueprint.ID, nil
}

// stage3_ContentGeneration 阶段3: 内容生成
func (o *Orchestrator) stage3_ContentGeneration(narrativeID string, params CreationParams, result *CreationResult) (int, int, error) {
	// 获取叙事蓝图
	blueprint, err := o.db.GetNarrativeBlueprint(narrativeID)
	if err != nil {
		return 0, 0, fmt.Errorf("获取叙事蓝图失败: %w", err)
	}

	// 获取世界设定
	world, err := o.db.GetWorld(blueprint.WorldID)
	if err != nil {
		return 0, 0, fmt.Errorf("获取世界设定失败: %w", err)
	}

	// 确定生成范围
	startChapter := params.Options.StartChapter
	endChapter := params.Options.EndChapter
	if endChapter == 0 || endChapter > len(blueprint.ChapterPlans) {
		endChapter = len(blueprint.ChapterPlans)
	}

	sceneCount := 0
	totalWordCount := 0

	// 获取风格配置
	style := writer.DefaultStyle()
	if params.Options.Style != "" {
		if styleProfile, ok := writer.GetStyle(params.Options.Style); ok {
			params.Options.Style = styleProfile.Name
		}
	}

	// 逐章生成
	for i := startChapter - 1; i < endChapter; i++ {
		chapter := blueprint.ChapterPlans[i]
		log.Printf("[编排器] 生成第%d章: %s", chapter.Chapter, chapter.Title)

		// 获取该章的场景指令
		chapterScenes := getScenesForChapter(blueprint.Scenes, chapter.Chapter)

		for _, sceneInstr := range chapterScenes {
			// 生成场景
			sceneResult, err := o.writer.GenerateScene(writer.GenerateParams{
				BlueprintID:    blueprint.ID,
				Chapter:        sceneInstr.Chapter,
				Scene:          sceneInstr.Scene,
				Instruction:    &sceneInstr,
				PreviousSummary: buildPreviousSummary(blueprint.ChapterPlans[:i]),
				CharacterStates: buildCharacterStates(blueprint, world),
				WorldContext:   world,
				Style:          style,
			})

			if err != nil {
				log.Printf("[编排器] 警告: 场景%d-%d生成失败: %v", sceneInstr.Chapter, sceneInstr.Scene, err)
				continue
			}

			sceneCount++
			totalWordCount += sceneResult.WordCount
			log.Printf("[编排器] 场景%d-%d生成完成，字数: %d", sceneInstr.Chapter, sceneInstr.Scene, sceneResult.WordCount)
		}
	}

	return sceneCount, totalWordCount, nil
}

// ResumeGeneration 恢复生成（从中断处继续）
func (o *Orchestrator) ResumeGeneration(projectID string) error {
	project, err := o.db.GetProject(projectID)
	if err != nil {
		return fmt.Errorf("获取项目失败: %w", err)
	}

	if project.Status != models.StatusPaused {
		return fmt.Errorf("项目状态不是暂停，无法恢复")
	}

	project.Status = models.StatusGenerating
	o.db.SaveProject(project)

	// 获取蓝图和已生成的场景
	blueprint, err := o.db.GetNarrativeBlueprint(project.NarrativeID)
	if err != nil {
		return fmt.Errorf("获取蓝图失败: %w", err)
	}

	// 找到下一个未生成的场景
	world, _ := o.db.GetWorld(blueprint.WorldID)
	style := writer.DefaultStyle()

	for _, chapter := range blueprint.ChapterPlans {
		chapterScenes := getScenesForChapter(blueprint.Scenes, chapter.Chapter)

		for _, sceneInstr := range chapterScenes {
			// 检查是否已生成
			existing, _ := o.db.GetSceneByBlueprintAndChapter(blueprint.ID, sceneInstr.Chapter, sceneInstr.Scene)
			if existing != nil {
				continue // 已生成，跳过
			}

			// 生成场景
			_, err := o.writer.GenerateScene(writer.GenerateParams{
				BlueprintID:    blueprint.ID,
				Chapter:        sceneInstr.Chapter,
				Scene:          sceneInstr.Scene,
				Instruction:    &sceneInstr,
				PreviousSummary: buildPreviousSummary(blueprint.ChapterPlans[:chapter.Chapter-1]),
				CharacterStates: buildCharacterStates(blueprint, world),
				WorldContext:   world,
				Style:          style,
			})

			if err != nil {
				log.Printf("场景生成失败: %v", err)
				continue
			}
		}
	}

	// 更新项目状态
	project.Status = models.StatusCompleted
	o.db.SaveProject(project)

	return nil
}

// PauseGeneration 暂停生成
func (o *Orchestrator) PauseGeneration(projectID string) error {
	project, err := o.db.GetProject(projectID)
	if err != nil {
		return fmt.Errorf("获取项目失败: %w", err)
	}

	if project.Status != models.StatusGenerating {
		return fmt.Errorf("项目状态不是生成中，无法暂停")
	}

	project.Status = models.StatusPaused
	return o.db.SaveProject(project)
}

// GetProjectProgress 获取项目进度
func (o *Orchestrator) GetProjectProgress(projectID string) (*ProjectProgress, error) {
	project, err := o.db.GetProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("获取项目失败: %w", err)
	}

	progress := &ProjectProgress{
		ProjectID:   project.ID,
		ProjectName: project.Name,
		Status:      string(project.Status),
		Progress:    project.Progress,
		CurrentStage: determineStage(project.Status),
	}

	// 获取详细进度
	if project.WorldID != "" {
		progress.WorldCompleted = true
	}
	if project.NarrativeID != "" {
		progress.NarrativeCompleted = true
		if blueprint, err := o.db.GetNarrativeBlueprint(project.NarrativeID); err == nil {
			progress.TotalChapters = len(blueprint.ChapterPlans)
			progress.TotalScenes = len(blueprint.Scenes)
		}
	}

	// 统计已生成的场景
	if project.NarrativeID != "" {
		scenes := o.db.ListScenesByBlueprint(project.NarrativeID)
		progress.GeneratedScenes = len(scenes)
		for _, scene := range scenes {
			progress.WordCount += scene.WordCount
		}
		progress.CompletionPercent = float64(progress.GeneratedScenes*100) / float64(progress.TotalScenes)
	}

	return progress, nil
}

// ProjectProgress 项目进度
type ProjectProgress struct {
	ProjectID         string  `json:"project_id"`
	ProjectName       string  `json:"project_name"`
	Status            string  `json:"status"`
	Progress          float64 `json:"progress"`
	CurrentStage      string  `json:"current_stage"`
	WorldCompleted    bool    `json:"world_completed"`
	NarrativeCompleted bool    `json:"narrative_completed"`
	TotalChapters     int     `json:"total_chapters"`
	TotalScenes       int     `json:"total_scenes"`
	GeneratedScenes   int     `json:"generated_scenes"`
	WordCount         int     `json:"word_count"`
	CompletionPercent float64 `json:"completion_percent"`
}

// 辅助函数
func parseWorldType(t string) models.WorldType {
	switch t {
	case "fantasy", "奇幻":
		return models.WorldFantasy
	case "scifi", "科幻":
		return models.WorldScifi
	case "urban", "都市":
		return models.WorldUrban
	case "historical", "历史":
		return models.WorldHistorical
	case "wuxia", "武侠":
		return models.WorldWuxia
	case "xianxia", "仙侠":
		return models.WorldXianxia
	default:
		return models.WorldFantasy
	}
}

func parseWorldScale(s string) models.WorldScale {
	switch s {
	case "village", "村庄":
		return models.ScaleVillage
	case "city", "城市":
		return models.ScaleCity
	case "nation", "国家":
		return models.ScaleNation
	case "continent", "大陆":
		return models.ScaleContinent
	case "planet", "星球":
		return models.ScalePlanet
	default:
		return models.ScaleContinent
	}
}

func parseNarrativeStructure(s string) narrative.NarrativeStructure {
	switch s {
	case "three_act", "三幕":
		return narrative.StructureThreeAct
	case "heros_journey", "英雄之旅":
		return narrative.StructureHerosJourney
	case "save_the_cat", "救猫咪":
		return narrative.StructureSaveTheCat
	default:
		return narrative.StructureThreeAct
	}
}

func getScenesForChapter(scenes []models.SceneInstruction, chapter int) []models.SceneInstruction {
	result := make([]models.SceneInstruction, 0)
	for _, scene := range scenes {
		if scene.Chapter == chapter {
			result = append(result, scene)
		}
	}
	return result
}

func buildPreviousSummary(chapters []models.ChapterPlan) string {
	if len(chapters) == 0 {
		return ""
	}
	summary := "前情提要："
	for _, ch := range chapters {
		summary += fmt.Sprintf("第%d章%s；", ch.Chapter, ch.Title)
	}
	return summary
}

func buildCharacterStates(blueprint *models.NarrativeBlueprint, world *models.WorldSetting) map[string]*writer.CharacterContext {
	// 从世界的种族创建基础角色状态
	states := make(map[string]*writer.CharacterContext)
	for _, race := range world.Civilization.Races {
		states[race.Name] = &writer.CharacterContext{
			ID:            race.ID,
			Name:          race.Name,
			CurrentEmotion: "平静",
			Location:      "",
			Knowledge:     []string{},
			Relationships: make(map[string]string),
		}
	}
	return states
}

func determineStage(status models.ProjectStatus) string {
	switch status {
	case models.StatusDraft:
		return "草稿阶段"
	case models.StatusBuilding:
		return "构建中"
	case models.StatusGenerating:
		return "生成中"
	case models.StatusCompleted:
		return "已完成"
	case models.StatusPaused:
		return "已暂停"
	case models.StatusFailed:
		return "失败"
	default:
		return "未知"
	}
}
