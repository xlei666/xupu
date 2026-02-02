// Package narrative 叙事器 - 系统的大脑
// 负责通过多轮动态演化创造故事性
package narrative

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/config"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/llm"
)

// ============================================
// 动态演化系统
// ============================================

// EvolutionState 演化状态
type EvolutionState struct {
	CurrentRound    int                         `json:"current_round"`     // 当前演化轮次
	MaxRounds       int                         `json:"max_rounds"`        // 最大轮次
	WorldContext    *models.WorldSetting       `json:"world_context"`    // 世界设定上下文
	Characters      map[string]*CharacterState `json:"characters"`       // 角色状态
	Conflicts       []*ConflictThread          `json:"conflicts"`        // 冲突线程
	Foreshadowing   []*Foreshadow              `json:"foreshadowing"`    // 伏笔
	PlotThreads     []*PlotThread               `json:"plot_threads"`     // 情节线程
	ThemeEvolution  *ThemeEvolutionState        `json:"theme_evolution"` // 主题演化
	NarrativeDepth  int                         `json:"narrative_depth"`   // 叙事深度（0-10）
	StoryHook       string                      `json:"story_hook"`       // 故事钩子
	EvolutionLog    []EvolutionLogEntry         `json:"evolution_log"`    // 演化日志

	// 新增：关系网络
	RelationshipNetwork *RelationshipNetwork    `json:"relationship_network"` // 关系网络

	// 新增：伏笔计划（在生成大纲之前规划）
	ForeshadowPlan   []*ForeshadowPlan         `json:"foreshadow_plan"`   // 伏笔计划

	// 新增：故事架构（在阶段1确定）
	StoryArchitecture *StoryArchitecture        `json:"story_architecture"` // 故事架构

	// 新增：全局大纲（关键事件序列）
	GlobalOutline    *GlobalOutline            `json:"global_outline"`    // 全局大纲

	// 新增：章节规划（在阶段6确定）
	ChapterPlan      *ChapterPlan              `json:"chapter_plan"`      // 章节规划

	// 新增：角色演化追踪
	CharacterEvolution map[string]*CharacterEvolutionTracker `json:"character_evolution"` // 角色演化追踪
}

// EvolutionLogEntry 演化日志条目
type EvolutionLogEntry struct {
	Round    int       `json:"round"`
	Timestamp time.Time `json:"timestamp"`
	Action   string    `json:"action"`   // evolve_conflict, deepen_character, plant_foreshadow等
	Details  string    `json:"details"`
	Changes  []string  `json:"changes"` // 产生的变化
}

// EvolutionRound 演化轮次类型
type EvolutionRound string

const (
	RoundCharacterCreation  EvolutionRound = "character_creation"   // 角色创建
	RoundConflictDesign     EvolutionRound = "conflict_design"      // 冲突设计
	RoundConflictEvolution  EvolutionRound = "conflict_evolution"   // 冲突演化
	RoundCharacterDeepen    EvolutionRound = "character_deepen"     // 角色深化
	RoundForeshadowPlant    EvolutionRound = "foreshadow_plant"     // 种下伏笔
	RoundForeshadowWeave    EvolutionRound = "foreshadow_weave"     // 编织伏笔
	RoundThemeDeepen        EvolutionRound = "theme_deepen"         // 主题深化
	RoundPlotTwist          EvolutionRound = "plot_twist"           // 情节转折
	RoundClimaxBuild        EvolutionRound = "climax_build"         // 高潮构建
	RoundResolutionPlan     EvolutionRound = "resolution_plan"      // 结局规划
)

// ============================================
// 角色情感系统
// ============================================

// CharacterState 角色状态（用于叙事演化）
type CharacterState struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Role            string              `json:"role"`            // 主角/反派/配角
	EmotionalState  EmotionalSystem     `json:"emotional_state"` // 情感系统
	Desires         DesireSystem        `json:"desires"`         // 欲望系统
	Relationships   map[string]*RelationshipState `json:"relationships"` // 关系网络
	ArcProgress     float64             `json:"arc_progress"`     // 弧光进度 0-1
	InternalConflicts []string          `json:"internal_conflicts"` // 内在冲突
	Secrets         []string            `json:"secrets"`          // 秘密
}

// EmotionalSystem 情感系统
type EmotionalSystem struct {
	CurrentEmotion   string   `json:"current_emotion"`   // 当前主导情绪
	EmotionalIntensity int    `json:"emotional_intensity"` // 情绪强度 0-100
	EmotionalStack   []string `json:"emotional_stack"`   // 情绪栈（深层情绪）
	Triggers         []string `json:"triggers"`          // 情绪触发器
	EmpathyLevel     int      `json:"empathy_level"`     // 共情能力 0-100
	EQ               int      `json:"eq"`                // 情商 0-100
}

// DesireSystem 欲望系统
type DesireSystem struct {
	ConsciousWant    string   `json:"conscious_want"`    // 表层欲望（他想要什么）
	UnconsciousNeed  string   `json:"unconscious_need"`  // 深层需求（他真正需要什么）
	Fear             string   `json:"fear"`              // 最大恐惧
	MaskingBehavior  []string `json:"masking_behavior"`  // 掩饰行为
	WantVsNeedGap    string   `json:"want_vs_need_gap"`  // 欲望与需求的差距
}

// RelationshipState 关系状态
type RelationshipState struct {
	TargetCharacterID string   `json:"target_character_id"`
	VisibleEmotion    int      `json:"visible_emotion"`    // 表面情感 -100到100
	HiddenEmotion     int      `json:"hidden_emotion"`      // 隐藏情感
	PowerDynamic      string   `json:"power_dynamic"`      // 权力动态
	SharedHistory     []string `json:"shared_history"`     // 共同经历
	UnspokenTension   []string `json:"unspoken_tension"`  // 未言明的紧张
	SecretsFrom       []string `json:"secrets_from"`       // 向对方隐瞒的秘密
}

// ============================================
// 冲突系统（动态发展）
// ============================================

// ConflictThread 冲突线程（贯穿故事始终）
type ConflictThread struct {
	ID               string           `json:"id"`
	Type             string           `json:"type"`             // 内在冲突/人际冲突/社会冲突/存在冲突
	CoreQuestion     string           `json:"core_question"`     // 核心问题
	Participants     []string         `json:"participants"`      // 参与者ID
	CurrentIntensity int              `json:"current_intensity"` // 当前强度 0-100
	EvolutionPath    []ConflictStage  `json:"evolution_path"`    // 演化路径
	Stakes           []string         `json:"stakes"`            // 赌注（输了会怎样）
	ThematicRelevance string          `json:"thematic_relevance"` // 与主题的关联
	IsResolved       bool             `json:"is_resolved"`       // 是否已解决
	Resolution       string           `json:"resolution"`        // 解决方式
}

// ConflictStage 冲突阶段
type ConflictStage struct {
	Stage       string   `json:"stage"`       // 阶段名称
	Description string   `json:"description"` // 描述
	Intensity   int      `json:"intensity"`   // 强度 0-100
	Events      []string `json:"events"`      // 关键事件
	EmotionalImpact map[string]string `json:"emotional_impact"` // 对各角色的情感影响
}

// ============================================
// 伏笔系统
// ============================================

// Foreshadow 伏笔
type Foreshadow struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`        // 象征式/对话式/情节式/角色式
	Content     string    `json:"content"`     // 伏笔内容
	PlantRound  int       `json:"plant_round"` // 种下的轮次
	PlantScene  string    `json:"plant_scene"` // 种下的场景
	PayoffRound int       `json:"payoff_round"` // 回收的轮次
	PayoffScene string    `json:"payoff_scene"`// 回收的场景
	Subtlety    int       `json:"subtlety"`    // 隐蔽程度 0-100
	IsPlanted   bool      `json:"is_planted"`  // 是否已种下
	IsPaidOff   bool      `json:"is_paid_off"`  // 是否已回收
	RelatedThemes []string `json:"related_themes"` // 关联的主题
}

// ============================================
// 情节线程
// ============================================

// PlotThread 情节线程
type PlotThread struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        string           `json:"type"`        // 主线/副线/背景线
	Characters  []string         `json:"characters"`  // 涉及的角色
	KeyEvents   []PlotEvent      `json:"key_events"`  // 关键事件
	Tension     int              `json:"tension"`      // 当前张力 0-100
	Status      string           `json:"status"`      // dormant/active/paused/resolved
	Dependencies []string       `json:"dependencies"` // 依赖的其他线程
}

// PlotEvent 情节事件
type PlotEvent struct {
	Sequence    int       `json:"sequence"`
	Chapter     int       `json:"chapter"`
	Description string    `json:"description"`
	Reversal    bool      `json:"reversal"` // 是否是逆转点
}

// ============================================
// 主题演化
// ============================================

// ThemeEvolutionState 主题演化状态
type ThemeEvolutionState struct {
	CoreTheme     string              `json:"core_theme"`     // 核心主题
	ThematicLayers []ThematicLayer    `json:"thematic_layers"` // 主题层次
	SymbolTracker  map[string]*Symbol `json:"symbol_tracker"` // 象征追踪
	MotifProgress  map[string]int     `json:"motif_progress"`  // 母题进展（出现次数）
}

// ThematicLayer 主题层次
type ThematicLayer struct {
	Layer     string   `json:"layer"`     // surface/middle/deep/philosophical
	Expression string  `json:"expression"` // 表达方式
	Chapter    int      `json:"chapter"`   // 出现章节
	Deepened   bool     `json:"deepened"`  // 是否已深化
}

// Symbol 象征
type Symbol struct {
	Name        string   `json:"name"`
	Meaning     string   `json:"meaning"`
	Appearances []int   `json:"appearances"` // 出现的章节
	Evolution   []string `json:"evolution"`   // 含义的演化
}

// ============================================
// 演化引擎
// ============================================

// EvolutionEngine 演化引擎
type EvolutionEngine struct {
	db     db.Database
	cfg    *config.Config
	client *llm.Client
	mapping *config.ModuleMapping
}

// NewEvolutionEngine 创建演化引擎
func NewEvolutionEngine() (*EvolutionEngine, error) {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	client, mapping, err := llm.NewClientForModule("narrative_engine")
	if err != nil {
		return nil, fmt.Errorf("创建LLM客户端失败: %w", err)
	}

	return &EvolutionEngine{
		db:      db.Get(),
		cfg:     cfg,
		client:  client,
		mapping: mapping,
	}, nil
}

// CreateEvolutionState 创建初始演化状态
func (ee *EvolutionEngine) CreateEvolutionState(worldID string) (*EvolutionState, error) {
	world, err := ee.db.GetWorld(worldID)
	if err != nil {
		return nil, fmt.Errorf("获取世界设定失败: %w", err)
	}

	state := &EvolutionState{
		CurrentRound:   0,
		MaxRounds:      10, // 默认10轮演化
		WorldContext:   world,
		Characters:     make(map[string]*CharacterState),
		Conflicts:      make([]*ConflictThread, 0),
		Foreshadowing:  make([]*Foreshadow, 0),
		PlotThreads:    make([]*PlotThread, 0),
		ThemeEvolution: &ThemeEvolutionState{
			CoreTheme:     world.Philosophy.CoreQuestion,
			ThematicLayers: make([]ThematicLayer, 0),
			SymbolTracker:  make(map[string]*Symbol),
			MotifProgress:  make(map[string]int),
		},
		NarrativeDepth: 0,
		StoryHook:      ee.generateStoryHook(world),
		EvolutionLog:   make([]EvolutionLogEntry, 0),
	}

	// 记录初始状态
	state.logAction(0, "initialize", "创建演化状态", []string{"初始化完成"})

	return state, nil
}

// Evolve 执行一轮演化
func (ee *EvolutionEngine) Evolve(state *EvolutionState, roundType EvolutionRound) (*EvolutionResult, error) {
	state.CurrentRound++
	var result *EvolutionResult
	var err error

	switch roundType {
	case RoundCharacterCreation:
		result, err = ee.evolveCharacterCreation(state)
	case RoundConflictDesign:
		result, err = ee.evolveConflictDesign(state)
	case RoundConflictEvolution:
		result, err = ee.evolveConflictEvolution(state)
	case RoundCharacterDeepen:
		result, err = ee.evolveCharacterDeepen(state)
	case RoundForeshadowPlant:
		result, err = ee.evolveForeshadowPlant(state)
	case RoundThemeDeepen:
		result, err = ee.evolveThemeDeepen(state)
	case RoundPlotTwist:
		result, err = ee.evolvePlotTwist(state)
	default:
		result, err = ee.evolveGeneral(state)
	}

	if err != nil {
		return nil, err
	}

	// 更新叙事深度
	if state.CurrentRound%3 == 0 && state.NarrativeDepth < 10 {
		state.NarrativeDepth++
	}

	// 记录日志
	state.logAction(state.CurrentRound, string(roundType), result.Summary, result.Changes)

	return result, nil
}

// EvolutionResult 演化结果
type EvolutionResult struct {
	Round      int                `json:"round"`
	Type       string             `json:"type"`
	Summary    string             `json:"summary"`
	Changes    []string           `json:"changes"`
	NewContent *EvolutionNewContent `json:"new_content"`
	QualityScore int              `json:"quality_score"` // 故事性质量评分 0-100
}

// EvolutionNewContent 演化产生的新内容
type EvolutionNewContent struct {
	Characters     []*CharacterState  `json:"characters,omitempty"`
	Conflicts      []*ConflictThread   `json:"conflicts,omitempty"`
	Foreshadows    []*Foreshadow       `json:"foreshadows,omitempty"`
	PlotEvents     []*PlotEvent        `json:"plot_events,omitempty"`
	ThematicLayers []ThematicLayer     `json:"thematic_layers,omitempty"`
}

// 演化轮次实现

// evolveCharacterCreation 角色创建演化
func (ee *EvolutionEngine) evolveCharacterCreation(state *EvolutionState) (*EvolutionResult, error) {
	// 从世界设定中提取角色模板
	characterTemplates := ee.extractCharacterTemplates(state.WorldContext)

	// 为每个角色创建完整的情感系统
	for _, template := range characterTemplates {
		charState, err := ee.createCharacterState(template, state)
		if err != nil {
			continue
		}
		state.Characters[charState.ID] = charState
	}

	return &EvolutionResult{
		Round:   state.CurrentRound,
		Type:    "character_creation",
		Summary: fmt.Sprintf("创建了%d个角色", len(characterTemplates)),
		Changes: []string{fmt.Sprintf("角色数量: %d", len(state.Characters))},
		NewContent: &EvolutionNewContent{
			Characters: ee.characterStateSlice(state.Characters),
		},
		QualityScore: ee.evaluateCharacterQuality(state),
	}, nil
}

// evolveConflictDesign 冲突设计演化
func (ee *EvolutionEngine) evolveConflictDesign(state *EvolutionState) (*EvolutionResult, error) {
	// 基于角色欲望系统和世界设定中的矛盾，设计冲突
	conflicts, err := ee.designConflicts(state)
	if err != nil {
		return nil, err
	}

	state.Conflicts = append(state.Conflicts, conflicts...)

	return &EvolutionResult{
		Round:   state.CurrentRound,
		Type:    "conflict_design",
		Summary: fmt.Sprintf("设计了%d个冲突线程", len(conflicts)),
		Changes: []string{fmt.Sprintf("冲突线程: %d", len(state.Conflicts))},
		NewContent: &EvolutionNewContent{
			Conflicts: conflicts,
		},
		QualityScore: ee.evaluateConflictQuality(state),
	}, nil
}

// evolveConflictEvolution 冲突演化（升级）
func (ee *EvolutionEngine) evolveConflictEvolution(state *EvolutionState) (*EvolutionResult, error) {
	var changes []string

	// 对每个冲突进行演化（升级强度、添加阶段）
	for _, conflict := range state.Conflicts {
		if conflict.IsResolved {
			continue
		}

		// 增加冲突强度
		if conflict.CurrentIntensity < 90 {
			oldIntensity := conflict.CurrentIntensity
			conflict.CurrentIntensity += 10
			changes = append(changes, fmt.Sprintf("冲突%s强度: %d→%d", conflict.ID, oldIntensity, conflict.CurrentIntensity))
		}

		// 添加新的冲突阶段
		stageNum := len(conflict.EvolutionPath) + 1
		newStage := ConflictStage{
			Stage:       fmt.Sprintf("阶段%d", stageNum),
			Description: ee.generateNextConflictStage(conflict, state),
			Intensity:   conflict.CurrentIntensity,
			Events:      ee.generateEventsForStage(conflict, stageNum, state),
			EmotionalImpact: ee.generateEmotionalImpact(conflict, state),
		}
		conflict.EvolutionPath = append(conflict.EvolutionPath, newStage)

		// 更新主题关联
		if conflict.ThematicRelevance == "" {
			conflict.ThematicRelevance = ee.generateThematicRelevance(conflict, state)
		}

		// 更新角色情感状态
		for _, charID := range conflict.Participants {
			if char, ok := state.Characters[charID]; ok {
				ee.updateCharacterEmotionFromConflict(char, conflict, &newStage)
			}
		}
	}

	return &EvolutionResult{
		Round:        state.CurrentRound,
		Type:         "conflict_evolution",
		Summary:      "冲突演化升级",
		Changes:      changes,
		QualityScore: ee.evaluateConflictQuality(state),
	}, nil
}

// evolveCharacterDeepen 角色深化演化
func (ee *EvolutionEngine) evolveCharacterDeepen(state *EvolutionState) (*EvolutionResult, error) {
	var changes []string

	// 对每个角色进行深化
	for _, char := range state.Characters {
		// 深化内在冲突
		newConflicts := ee.deepenInternalConflicts(char, state)
		if len(newConflicts) > 0 {
			char.InternalConflicts = append(char.InternalConflicts, newConflicts...)
			changes = append(changes, fmt.Sprintf("%s内在冲突深化", char.Name))
		}

		// 添加秘密
		newSecret := ee.generateSecret(char, state)
		if newSecret != "" {
			char.Secrets = append(char.Secrets, newSecret)
			changes = append(changes, fmt.Sprintf("%s获得新秘密", char.Name))
		}

		// 演化欲望系统
		ee.evolveDesireSystem(char, state)
	}

	return &EvolutionResult{
		Round:        state.CurrentRound,
		Type:         "character_deepen",
		Summary:      "角色深度增加",
		Changes:      changes,
		QualityScore: ee.evaluateCharacterQuality(state),
	}, nil
}

// evolveForeshadowPlant 种下伏笔
func (ee *EvolutionEngine) evolveForeshadowPlant(state *EvolutionState) (*EvolutionResult, error) {
	foreshadows := ee.generateForeshadows(state)

	state.Foreshadowing = append(state.Foreshadowing, foreshadows...)

	return &EvolutionResult{
		Round:   state.CurrentRound,
		Type:    "foreshadow_plant",
		Summary: fmt.Sprintf("种下了%d个伏笔", len(foreshadows)),
		Changes: []string{fmt.Sprintf("伏笔总数: %d", len(state.Foreshadowing))},
		NewContent: &EvolutionNewContent{
			Foreshadows: foreshadows,
		},
		QualityScore: ee.evaluateForeshadowQuality(state),
	}, nil
}

// evolveThemeDeepen 主题深化
func (ee *EvolutionEngine) evolveThemeDeepen(state *EvolutionState) (*EvolutionResult, error) {
	newLayers := ee.deepenTheme(state)

	state.ThemeEvolution.ThematicLayers = append(state.ThemeEvolution.ThematicLayers, newLayers...)

	return &EvolutionResult{
		Round:   state.CurrentRound,
		Type:    "theme_deepen",
		Summary: "主题层次深化",
		Changes: []string{fmt.Sprintf("主题层次: %d", len(state.ThemeEvolution.ThematicLayers))},
		NewContent: &EvolutionNewContent{
			ThematicLayers: newLayers,
		},
		QualityScore: ee.evaluateThemeQuality(state),
	}, nil
}

// evolvePlotTwist 情节转折
func (ee *EvolutionEngine) evolvePlotTwist(state *EvolutionState) (*EvolutionResult, error) {
	twist := ee.generatePlotTwist(state)

	return &EvolutionResult{
		Round:   state.CurrentRound,
		Type:    "plot_twist",
		Summary: "添加情节转折",
		Changes: []string{twist},
		QualityScore: ee.evaluatePlotQuality(state),
	}, nil
}

// evolveGeneral 通用演化
func (ee *EvolutionEngine) evolveGeneral(state *EvolutionState) (*EvolutionResult, error) {
	// LLM调用来评估当前状态并提出改进
	prompt := ee.buildEvolutionPrompt(state)
	systemPrompt := ee.cfg.GetNarrativeEngineSystem()

	result, err := ee.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var evolutionAdvice struct {
		Suggestions []string `json:"suggestions"`
		Priority     string   `json:"priority"`
		NextSteps    []string `json:"next_steps"`
	}

	if err := json.Unmarshal([]byte(result), &evolutionAdvice); err != nil {
		extracted := extractJSON(result)
		json.Unmarshal([]byte(extracted), &evolutionAdvice)
	}

	return &EvolutionResult{
		Round:   state.CurrentRound,
		Type:    "general_evolution",
		Summary: "通用演化评估",
		Changes: evolutionAdvice.Suggestions,
		QualityScore: ee.calculateOverallQuality(state),
	}, nil
}

// ============================================
// 辅助方法
// ============================================

func (s *EvolutionState) logAction(round int, action, details string, changes []string) {
	s.EvolutionLog = append(s.EvolutionLog, EvolutionLogEntry{
		Round:    round,
		Timestamp: time.Now(),
		Action:   action,
		Details:  details,
		Changes:  changes,
	})
}

// 以下方法需要调用LLM实现（简化版本）
// extractCharacterTemplates 从世界设定中提取角色模板
// 如果世界设定中没有种族信息，则通过LLM生成角色概念
func (ee *EvolutionEngine) extractCharacterTemplates(world *models.WorldSetting) []models.Race {
	// 如果已有种族信息，直接返回
	if world.Civilization.Races != nil && len(world.Civilization.Races) > 0 {
		return world.Civilization.Races
	}

	// 没有种族信息时，通过LLM生成角色概念
	return ee.generateCharactersByLLM(world)
}

// generateCharactersByLLM 通过LLM生成角色概念
func (ee *EvolutionEngine) generateCharactersByLLM(world *models.WorldSetting) []models.Race {
	// 构建提示词
	prompt := ee.buildCharacterGenerationPrompt(world)
	systemPrompt := `你是一位专业的故事策划师，擅长创造深刻、复杂的角色。
请基于世界设定生成3-5个主要角色的概念。`

	// 调用LLM
	result, err := ee.callWithRetry(prompt, systemPrompt)
	if err != nil {
		// LLM失败时返回默认角色
		return ee.getDefaultCharacters(world)
	}

	// 解析LLM输出
	var characterData struct {
		Characters []struct {
			Name        string   `json:"name"`
			Role        string   `json:"role"`        // 主角/反派/配角
			Description string   `json:"description"`
			Traits      []string `json:"traits"`
			Abilities   []string `json:"abilities"`
		} `json:"characters"`
	}

	if err := json.Unmarshal([]byte(result), &characterData); err != nil {
		extracted := extractJSON(result)
		json.Unmarshal([]byte(extracted), &characterData)
	}

	// 如果解析失败，返回默认角色
	if len(characterData.Characters) == 0 {
		return ee.getDefaultCharacters(world)
	}

	// 转换为Race格式（复用Race结构表示角色概念）
	races := make([]models.Race, 0, len(characterData.Characters))
	for _, char := range characterData.Characters {
		races = append(races, models.Race{
			Name:        char.Name,
			Description: char.Description,
			Traits:      char.Traits,
			Abilities:   char.Abilities,
		})
	}

	return races
}

// buildCharacterGenerationPrompt 构建角色生成提示词
func (ee *EvolutionEngine) buildCharacterGenerationPrompt(world *models.WorldSetting) string {
	var prompt strings.Builder

	prompt.WriteString("# 角色生成任务\n\n")
	prompt.WriteString(fmt.Sprintf("## 世界设定\n"))
	prompt.WriteString(fmt.Sprintf("- 世界名称: %s\n", world.Name))
	prompt.WriteString(fmt.Sprintf("- 世界类型: %s\n", world.Type))
	prompt.WriteString(fmt.Sprintf("- 核心问题: %s\n", world.Philosophy.CoreQuestion))

	if world.Philosophy.ValueSystem.HighestGood != "" {
		prompt.WriteString(fmt.Sprintf("- 最高善: %s\n", world.Philosophy.ValueSystem.HighestGood))
	}
	if world.Philosophy.ValueSystem.UltimateEvil != "" {
		prompt.WriteString(fmt.Sprintf("- 最大恶: %s\n", world.Philosophy.ValueSystem.UltimateEvil))
	}

	// 故事土壤
	if len(world.StorySoil.SocialConflicts) > 0 {
		prompt.WriteString("\n## 社会冲突\n")
		for i, conflict := range world.StorySoil.SocialConflicts {
			prompt.WriteString(fmt.Sprintf("%d. %s\n", i+1, conflict.Description))
		}
	}

	if len(world.StorySoil.PotentialPlotHooks) > 0 {
		prompt.WriteString("\n## 故事钩子\n")
		for i, hook := range world.StorySoil.PotentialPlotHooks {
			prompt.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, hook.Description, hook.StoryPotential))
		}
	}

	prompt.WriteString("\n# 任务\n")
	prompt.WriteString("基于以上世界设定，生成3-5个主要角色的概念。\n")
	prompt.WriteString("每个角色应包含：\n")
	prompt.WriteString("- name: 角色名称\n")
	prompt.WriteString("- role: 角色（主角/反派/重要配角）\n")
	prompt.WriteString("- description: 角色描述（100字左右）\n")
	prompt.WriteString("- traits: 性格特质（3-5个）\n")
	prompt.WriteString("- abilities: 能力或技能（3-5个）\n\n")

	prompt.WriteString("请以JSON格式返回：\n")
	prompt.WriteString(`{
  "characters": [
    {
      "name": "角色名",
      "role": "主角/反派/配角",
      "description": "角色描述...",
      "traits": ["特质1", "特质2", "特质3"],
      "abilities": ["能力1", "能力2", "能力3"]
    }
  ]
}`)

	return prompt.String()
}

// getDefaultCharacters 获取默认角色（LLM失败时的降级方案）
func (ee *EvolutionEngine) getDefaultCharacters(world *models.WorldSetting) []models.Race {
	// 基于世界类型生成默认角色
	protagonistName := "主角"
	antagonistName := "反派"

	switch world.Type {
	case "xianxia", "wuxia":
		protagonistName = "求道者"
		antagonistName = "魔修"
	case "fantasy":
		protagonistName = "冒险者"
		antagonistName = "黑暗领主"
	case "scifi":
		protagonistName = "探索者"
		antagonistName = "控制者"
	case "historical":
		protagonistName = "志士"
		antagonistName = "权臣"
	}

	return []models.Race{
		{
			Name:        protagonistName,
			Description: fmt.Sprintf("一个在%s世界中寻求答案的%s", world.Name, protagonistName),
			Traits:      []string{"勇敢", "好奇", "执着"},
			Abilities:   []string{"适应力强", "学习能力"},
		},
		{
			Name:        antagonistName,
			Description: fmt.Sprintf("与%s对立的%s", protagonistName, antagonistName),
			Traits:      []string{"狡猾", "强大", "冷酷"},
			Abilities:   []string{"操控人心", "强大力量"},
		},
		{
			Name:        "导师",
			Description: "引导主角成长的智者",
			Traits:      []string{"智慧", "神秘", "耐心"},
			Abilities:   []string{"丰富经验", "洞察力"},
		},
	}
}

// createCharacterState 使用LLM创建完整的角色状态
func (ee *EvolutionEngine) createCharacterState(race models.Race, state *EvolutionState) (*CharacterState, error) {
	charID := db.GenerateID("char")

	// 如果没有种族信息，创建默认角色
	if race.Name == "" {
		return &CharacterState{
			ID:   charID,
			Name: "未命名角色",
			Role: "supporting",
			EmotionalState: EmotionalSystem{
				CurrentEmotion:     "平静",
				EmotionalIntensity: 30,
				EmpathyLevel:       50,
				EQ:                 50,
			},
			Desires: DesireSystem{
				ConsciousWant:    "生存",
				UnconsciousNeed:  "归属",
				Fear:             "孤独",
			},
			Relationships: make(map[string]*RelationshipState),
			ArcProgress:   0,
			Secrets:       make([]string, 0),
		}, nil
	}

	// 使用LLM生成角色状态
	prompt := ee.buildCharacterCreationPrompt(race, state)
	systemPrompt := `你是一位专业的人物设计师，擅长创造深刻、复杂的角色。
请根据提供的种族信息和世界设定，生成一个完整的角色情感系统。`

	result, err := ee.callWithRetry(prompt, systemPrompt)
	if err != nil {
		// LLM失败时返回默认角色
		return ee.createDefaultCharacterState(charID, race.Name), nil
	}

	// 解析LLM输出
	var charData struct {
		Name            string `json:"name"`
		Role            string `json:"role"`
		CurrentEmotion  string `json:"current_emotion"`
		ConsciousWant   string `json:"conscious_want"`
		UnconsciousNeed string `json:"unconscious_need"`
		Fear            string `json:"fear"`
		EmpathyLevel    int    `json:"empathy_level"`
		EQ              int    `json:"eq"`
	}

	if err := json.Unmarshal([]byte(result), &charData); err != nil {
		extracted := extractJSON(result)
		json.Unmarshal([]byte(extracted), &charData)
	}

	// 填充默认值
	if charData.Name == "" {
		charData.Name = race.Name
	}
	if charData.Role == "" {
		charData.Role = "supporting"
	}
	if charData.CurrentEmotion == "" {
		charData.CurrentEmotion = "平静"
	}
	if charData.ConsciousWant == "" {
		charData.ConsciousWant = "生存与安全"
	}
	if charData.UnconsciousNeed == "" {
		charData.UnconsciousNeed = "被理解"
	}
	if charData.Fear == "" {
		charData.Fear = "孤独"
	}
	if charData.EmpathyLevel == 0 {
		charData.EmpathyLevel = 50
	}
	if charData.EQ == 0 {
		charData.EQ = 50
	}

	char := &CharacterState{
		ID:   charID,
		Name: charData.Name,
		Role: charData.Role,
		EmotionalState: EmotionalSystem{
			CurrentEmotion:     charData.CurrentEmotion,
			EmotionalIntensity: 50,
			EmpathyLevel:       charData.EmpathyLevel,
			EQ:                 charData.EQ,
		},
		Desires: DesireSystem{
			ConsciousWant:    charData.ConsciousWant,
			UnconsciousNeed:  charData.UnconsciousNeed,
			Fear:             charData.Fear,
		},
		Relationships: make(map[string]*RelationshipState),
		ArcProgress:   0,
		Secrets:       make([]string, 0),
	}

	return char, nil
}

// createDefaultCharacterState 创建默认角色状态
func (ee *EvolutionEngine) createDefaultCharacterState(id, name string) *CharacterState {
	return &CharacterState{
		ID:   id,
		Name: name,
		Role: "supporting",
		EmotionalState: EmotionalSystem{
			CurrentEmotion:     "平静",
			EmotionalIntensity: 30,
			EmpathyLevel:       50,
			EQ:                 50,
		},
		Desires: DesireSystem{
			ConsciousWant:    "生存",
			UnconsciousNeed:  "归属",
			Fear:             "未知",
		},
		Relationships: make(map[string]*RelationshipState),
		ArcProgress:   0,
		Secrets:       make([]string, 0),
	}
}

// buildCharacterCreationPrompt 构建角色创建提示词
func (ee *EvolutionEngine) buildCharacterCreationPrompt(race models.Race, state *EvolutionState) string {
	var prompt strings.Builder

	prompt.WriteString("# 角色创建任务\n\n")

	// 种族信息
	prompt.WriteString("## 种族信息\n")
	prompt.WriteString(fmt.Sprintf("- 名称: %s\n", race.Name))
	prompt.WriteString(fmt.Sprintf("- 描述: %s\n", race.Description))
	if len(race.Traits) > 0 {
		prompt.WriteString(fmt.Sprintf("- 特质: %s\n", strings.Join(race.Traits, "、")))
	}
	if len(race.Abilities) > 0 {
		prompt.WriteString(fmt.Sprintf("- 能力: %s\n", strings.Join(race.Abilities, "、")))
	}

	// 世界背景
	prompt.WriteString(ee.buildWorldContextSection(state))

	// 世界观（影响角色信仰）
	prompt.WriteString(ee.buildWorldviewSection(state))

	// 超自然体系（影响角色能力）
	prompt.WriteString(ee.buildSupernaturalSection(state))

	// 语言宗教（影响角色背景）
	prompt.WriteString(ee.buildCivilizationSection(state))

	// 社会阶层（角色出身参考）
	if len(state.WorldContext.Society.Classes) > 0 {
		prompt.WriteString("\n## 社会阶层\n")
		for i, class := range state.WorldContext.Society.Classes {
			prompt.WriteString(fmt.Sprintf("%d. %s (等级:%d)\n", i+1, class.Name, class.Rank))
		}
	}

	// 政治结构（影响角色立场）
	prompt.WriteString("\n## 政治环境\n")
	prompt.WriteString(fmt.Sprintf("- 政体类型: %s\n", state.WorldContext.Society.Politics.Type))
	prompt.WriteString(fmt.Sprintf("- 权力来源: %s\n", state.WorldContext.Society.Politics.LegitimacySource))
	// 权力层级
	if state.WorldContext.Society.Politics.PowerStructure != nil {
		if len(state.WorldContext.Society.Politics.PowerStructure.Formal) > 0 {
			levelNames := make([]string, 0, len(state.WorldContext.Society.Politics.PowerStructure.Formal))
			for _, pl := range state.WorldContext.Society.Politics.PowerStructure.Formal {
				levelNames = append(levelNames, pl.Name)
			}
			prompt.WriteString(fmt.Sprintf("- 权力层级: %s\n", strings.Join(levelNames, "→")))
		}
	}

	// 经济与贸易
	prompt.WriteString("\n## 经济环境\n")
	prompt.WriteString(fmt.Sprintf("- 经济类型: %s\n", state.WorldContext.Society.Economy.Type))
	if state.WorldContext.Society.Economy.TradeNetwork != "" {
		prompt.WriteString(fmt.Sprintf("- 贸易网络: %s\n", state.WorldContext.Society.Economy.TradeNetwork))
	}
	if len(state.WorldContext.Society.Economy.Currency) > 0 {
		prompt.WriteString(fmt.Sprintf("- 货币: %s\n", strings.Join(state.WorldContext.Society.Economy.Currency, ", ")))
	}

	// 法律体系（影响角色行为约束）
	if len(state.WorldContext.Society.Laws) > 0 {
		prompt.WriteString("\n## 法律体系\n")
		for i, law := range state.WorldContext.Society.Laws {
			prompt.WriteString(fmt.Sprintf("%d. %s (%s): %s\n", i+1, law.Name, law.Type, law.Description))
		}
	}

	// 历史背景（影响角色记忆）
	if len(state.WorldContext.History.Traumas) > 0 {
		prompt.WriteString("\n## 集体创伤\n")
		for _, trauma := range state.WorldContext.History.Traumas {
			prompt.WriteString(fmt.Sprintf("- %s\n", trauma))
		}
	}

	// 历史时代（角色可能经历的）
	prompt.WriteString(ee.buildHistoryDetailsSection(state))

	prompt.WriteString("\n# 任务\n")
	prompt.WriteString("基于以上信息创建一个角色，要求：\n")
	prompt.WriteString("1. 角色与世界设定有深度关联\n")
	prompt.WriteString("2. 拥有内在的欲望-需求冲突（表层欲望vs深层需求）\n")
	prompt.WriteString("3. 有明确的恐惧（这将成为角色的弱点）\n")
	prompt.WriteString("4. 性格复杂，有优点也有缺陷\n")
	prompt.WriteString("5. 考虑角色的社会阶层出身和历史背景影响\n")
	prompt.WriteString("6. 考虑超自然体系对角色能力的影响\n")

	prompt.WriteString("\n# 输出格式（JSON）\n")
	prompt.WriteString(`{
  "name": "角色名称",
  "role": "主角/反派/配角/导师",
  "social_class": "所属社会阶层",
  "current_emotion": "当前主导情绪",
  "conscious_want": "表层欲望（他想要什么）",
  "unconscious_need": "深层需求（他真正需要什么）",
  "fear": "最大恐惧",
  "supernatural_ability": "如果适用，简述超自然能力",
  "empathy_level": 0-100,
  "eq": 0-100
}`)

	return prompt.String()
}

// designConflicts 使用LLM设计冲突
func (ee *EvolutionEngine) designConflicts(state *EvolutionState) ([]*ConflictThread, error) {
	// 构建提示词
	prompt := ee.buildConflictDesignPrompt(state)
	systemPrompt := `你是一位专业的故事策划师，擅长设计复杂、深刻的冲突。
冲突是故事的动力，请设计多层次、多维度的冲突系统。`

	result, err := ee.callWithRetry(prompt, systemPrompt)
	if err != nil {
		// LLM失败时使用默认冲突生成
		return ee.createDefaultConflicts(state), nil
	}

	// 解析LLM输出
	var conflictData struct {
		Conflicts []struct {
			Type         string   `json:"type"`
			CoreQuestion string   `json:"core_question"`
			Participants []string `json:"participants"`
			Stakes       []string `json:"stakes"`
			Escalation   []string `json:"escalation_path"`
		} `json:"conflicts"`
	}

	if err := json.Unmarshal([]byte(result), &conflictData); err != nil {
		extracted := extractJSON(result)
		json.Unmarshal([]byte(extracted), &conflictData)
	}

	conflicts := make([]*ConflictThread, 0)

	// 获取角色列表
	charIDs := make([]string, 0, len(state.Characters))
	for id := range state.Characters {
		charIDs = append(charIDs, id)
	}

	for _, c := range conflictData.Conflicts {
		// 确定参与者
		participants := c.Participants
		if len(participants) == 0 && len(charIDs) > 0 {
			// 随机分配1-2个角色
			if len(charIDs) >= 2 {
				participants = charIDs[:2]
			} else {
				participants = charIDs
			}
		}

		// 构建演化路径
		evolutionPath := make([]ConflictStage, 0)
		for i, desc := range c.Escalation {
			evolutionPath = append(evolutionPath, ConflictStage{
				Stage:       fmt.Sprintf("阶段%d", i+1),
				Description: desc,
				Intensity:   30 + i*20, // 逐渐增加强度
				Events:      []string{},
				EmotionalImpact: make(map[string]string),
			})
		}

		conflict := &ConflictThread{
			ID:               db.GenerateID("conflict"),
			Type:             c.Type,
			CoreQuestion:     c.CoreQuestion,
			Participants:     participants,
			CurrentIntensity: 40,
			EvolutionPath:    evolutionPath,
			Stakes:           c.Stakes,
			IsResolved:       false,
		}
		conflicts = append(conflicts, conflict)
	}

	// 如果没有生成任何冲突，使用默认方法
	if len(conflicts) == 0 {
		return ee.createDefaultConflicts(state), nil
	}

	return conflicts, nil
}

// createDefaultConflicts 创建默认冲突（LLM失败时使用）
func (ee *EvolutionEngine) createDefaultConflicts(state *EvolutionState) []*ConflictThread {
	conflicts := make([]*ConflictThread, 0)

	for _, soilConflict := range state.WorldContext.StorySoil.SocialConflicts {
		conflict := &ConflictThread{
			ID:               db.GenerateID("conflict"),
			Type:             soilConflict.Type,
			CoreQuestion:     soilConflict.Description,
			CurrentIntensity: soilConflict.Tension,
			EvolutionPath: []ConflictStage{
				{
					Stage:       "初始阶段",
					Description: soilConflict.Description,
					Intensity:   soilConflict.Tension,
				},
			},
			IsResolved: false,
		}
		conflicts = append(conflicts, conflict)
	}

	return conflicts
}

// buildConflictDesignPrompt 构建冲突设计提示词
func (ee *EvolutionEngine) buildConflictDesignPrompt(state *EvolutionState) string {
	var prompt strings.Builder

	prompt.WriteString("# 冲突设计任务\n\n")

	// 世界背景
	prompt.WriteString(ee.buildWorldContextSection(state))

	// 超自然体系（冲突来源）
	prompt.WriteString(ee.buildSupernaturalSection(state))

	// 已有角色
	if len(state.Characters) > 0 {
		prompt.WriteString("\n## 已有角色\n")
		for _, char := range state.Characters {
			prompt.WriteString(fmt.Sprintf("- %s: 欲望=%s, 需求=%s, 恐惧=%s\n",
				char.Name, char.Desires.ConsciousWant, char.Desires.UnconsciousNeed, char.Desires.Fear))
		}
	}

	// 社会矛盾（来自故事土壤）
	if len(state.WorldContext.StorySoil.SocialConflicts) > 0 {
		prompt.WriteString("\n## 世界中的社会矛盾\n")
		for i, c := range state.WorldContext.StorySoil.SocialConflicts {
			prompt.WriteString(fmt.Sprintf("%d. [%s] %s (张力:%d)\n",
				i+1, c.Type, c.Description, c.Tension))
		}
	}

	// 权力结构（冲突来源）
	if len(state.WorldContext.StorySoil.PowerStructures) > 0 {
		prompt.WriteString("\n## 权力结构\n")
		for i, ps := range state.WorldContext.StorySoil.PowerStructures {
			levelNames := make([]string, 0, len(ps.Formal))
			for _, pl := range ps.Formal {
				levelNames = append(levelNames, pl.Name)
			}
			prompt.WriteString(fmt.Sprintf("%d. 明面权力: %s\n", i+1, strings.Join(levelNames, ", ")))
			if len(ps.Actual) > 0 {
				entityNames := make([]string, 0, len(ps.Actual))
				for _, a := range ps.Actual {
					entityNames = append(entityNames, a.Entity)
				}
				prompt.WriteString(fmt.Sprintf("   实际掌权者: %s\n", strings.Join(entityNames, ", ")))
			}
		}
	}

	// 社会冲突（来自社会设定）
	if len(state.WorldContext.Society.Conflicts) > 0 {
		prompt.WriteString("\n## 社会冲突\n")
		for i, c := range state.WorldContext.Society.Conflicts {
			parties := strings.Join(c.Parties, " vs ")
			prompt.WriteString(fmt.Sprintf("%d. %s: %s (张力:%d)\n",
				i+1, parties, c.Description, c.Tension))
		}
	}

	// 经济状况（冲突根源）
	prompt.WriteString("\n## 经济环境\n")
	prompt.WriteString(fmt.Sprintf("- 经济类型: %s\n", state.WorldContext.Society.Economy.Type))
	if state.WorldContext.Society.Economy.TradeNetwork != "" {
		prompt.WriteString(fmt.Sprintf("- 贸易网络: %s\n", state.WorldContext.Society.Economy.TradeNetwork))
	}
	if len(state.WorldContext.Society.Economy.Currency) > 0 {
		prompt.WriteString(fmt.Sprintf("- 货币: %s\n", strings.Join(state.WorldContext.Society.Economy.Currency, ", ")))
	}
	prompt.WriteString(ee.buildGeographySection(state))

	// 法律体系（约束冲突解决方式）
	if len(state.WorldContext.Society.Laws) > 0 {
		prompt.WriteString("\n## 法律体系\n")
		for i, law := range state.WorldContext.Society.Laws {
			prompt.WriteString(fmt.Sprintf("%d. %s (%s): %s - 违法后果: %s\n",
				i+1, law.Name, law.Type, law.Description, law.Penalty))
		}
	}

	// 文化细节
	cd := state.WorldContext.StorySoil.CulturalDetails
	if len(cd.Customs) > 0 || len(cd.Taboos) > 0 || len(cd.Slang) > 0 ||
		len(cd.Holidays) > 0 || len(cd.Arts) > 0 {
		prompt.WriteString("\n## 文化背景\n")
		if len(cd.Customs) > 0 {
			prompt.WriteString(fmt.Sprintf("习俗: %s\n", strings.Join(cd.Customs, ", ")))
		}
		if len(cd.Taboos) > 0 {
			prompt.WriteString(fmt.Sprintf("禁忌: %s\n", strings.Join(cd.Taboos, ", ")))
		}
		if len(cd.Slang) > 0 {
			prompt.WriteString(fmt.Sprintf("俚语: %s\n", strings.Join(cd.Slang, ", ")))
		}
		if len(cd.Holidays) > 0 {
			prompt.WriteString(fmt.Sprintf("节日: %s\n", strings.Join(cd.Holidays, ", ")))
		}
		if len(cd.Arts) > 0 {
			prompt.WriteString(fmt.Sprintf("艺术: %s\n", strings.Join(cd.Arts, ", ")))
		}
	}

	// 未解决问题（来自历史语境）
	if len(state.WorldContext.StorySoil.HistoricalContext.UnresolvedIssues) > 0 {
		prompt.WriteString("\n## 历史遗留问题\n")
		for _, issue := range state.WorldContext.StorySoil.HistoricalContext.UnresolvedIssues {
			prompt.WriteString(fmt.Sprintf("- %s\n", issue))
		}
	}

	prompt.WriteString("\n# 任务\n")
	prompt.WriteString("设计3-5个核心冲突，要求：\n")
	prompt.WriteString("1. 冲突类型多样（内在冲突/人际冲突/社会冲突/存在冲突）\n")
	prompt.WriteString("2. 每个冲突有明确的核心问题\n")
	prompt.WriteString("3. 冲突有升级路径（escalation_path）- 展示冲突如何从轻微发展到激烈\n")
	prompt.WriteString("4. 冲突有赌注（stakes）- 明确输赢的代价\n")
	prompt.WriteString("5. 冲突与世界主题相关\n")
	prompt.WriteString("6. 利用权力结构和社会矛盾作为冲突来源\n")
	prompt.WriteString("7. 考虑资源争夺和超自然体系对冲突的影响\n")

	prompt.WriteString("\n# 输出格式（JSON）\n")
	prompt.WriteString(`{
  "conflicts": [
    {
      "type": "内在冲突/人际冲突/社会冲突/存在冲突",
      "core_question": "冲突的核心问题是什么？",
      "participants": ["角色ID"],
      "stakes": ["输了会怎样？"],
      "escalation_path": ["阶段1", "阶段2", "阶段3", "阶段4"]
    }
  ]
}`)

	return prompt.String()
}

func (ee *EvolutionEngine) generateNextConflictStage(conflict *ConflictThread, state *EvolutionState) string {
	// 基于冲突类型和当前阶段生成下一个阶段的描述
	stageTemplates := map[string][]string{
		"内在冲突": {
			"主角开始意识到内心的矛盾",
			"内在冲突逐渐显现，影响行为",
			"矛盾激化，主角面临艰难选择",
			"崩溃边缘，主角必须做出决定",
			"关键的自我认知时刻",
		},
		"人际冲突": {
			"表面上的和谐开始破裂",
			"分歧公开化，关系紧张",
			"信任崩塌，对立加剧",
			"正面冲突爆发",
			"关系面临终极考验",
		},
		"社会冲突": {
			"潜在的不满开始浮现",
			"矛盾公开化，社会分裂",
			"冲突升级，对立阵营形成",
			"全面对抗爆发",
			"社会秩序面临重构",
		},
		"存在冲突": {
			"开始质疑存在的意义",
			"怀疑加深，信念动摇",
			"价值观崩溃，寻找新方向",
			"存在主义危机爆发",
			"找到新的存在意义",
		},
	}

	templates, ok := stageTemplates[conflict.Type]
	if !ok {
		templates = []string{
			"冲突萌芽",
			"矛盾显现",
			"冲突升级",
			"对抗激烈",
			"面临解决",
		}
	}

	stageNum := len(conflict.EvolutionPath)
	if stageNum < len(templates) {
		return templates[stageNum]
	}

	return "冲突进入新阶段"
}

// generateEventsForStage 为冲突阶段生成具体事件
func (ee *EvolutionEngine) generateEventsForStage(conflict *ConflictThread, stageNum int, state *EvolutionState) []string {
	events := make([]string, 0)

	// 基于冲突参与者生成事件
	if len(conflict.Participants) > 0 {
		// 为前两个参与者生成事件
		for i := 0; i < min(2, len(conflict.Participants)); i++ {
			charID := conflict.Participants[i]
			if char, ok := state.Characters[charID]; ok {
				event := fmt.Sprintf("%s经历情感波动，对%s有了新的认识",
					char.Name, conflict.Type)
				events = append(events, event)
			}
		}
	}

	// 添加阶段性事件
	stageEvents := map[int]string{
		1: "冲突的火花被点燃",
		2: "矛盾开始显现，局势发生变化",
		3: "冲突升级，双方采取行动",
		4: "对抗达到高峰",
		5: "面临最终抉择",
	}

	if event, ok := stageEvents[stageNum]; ok {
		events = append(events, event)
	}

	return events
}

// generateEmotionalImpact 生成冲突对角色的情感影响
func (ee *EvolutionEngine) generateEmotionalImpact(conflict *ConflictThread, state *EvolutionState) map[string]string {
	impact := make(map[string]string)

	// 为每个参与者生成情感影响
	for _, charID := range conflict.Participants {
		if char, ok := state.Characters[charID]; ok {
			if conflict.Type == "内在冲突" {
				impact[charID] = fmt.Sprintf("%s的内心矛盾加深，情绪在%s与%s之间摇摆",
					char.Name,
					char.EmotionalState.CurrentEmotion,
					ee.getOppositeEmotion(char.EmotionalState.CurrentEmotion))
			} else {
				impact[charID] = fmt.Sprintf("%s因%s而产生强烈的%s情绪",
					char.Name,
					conflict.CoreQuestion,
					char.EmotionalState.CurrentEmotion)
			}
		}
	}

	return impact
}

// generateThematicRelevance 生成冲突与主题的关联
func (ee *EvolutionEngine) generateThematicRelevance(conflict *ConflictThread, state *EvolutionState) string {
	if state.WorldContext.Philosophy.CoreQuestion == "" {
		return "该冲突推动故事发展"
	}

	return fmt.Sprintf("该%s是'%s'这一核心问题的具体体现",
		conflict.Type,
		state.WorldContext.Philosophy.CoreQuestion)
}

// getOppositeEmotion 获取相反的情绪
func (ee *EvolutionEngine) getOppositeEmotion(emotion string) string {
	opposites := map[string]string{
		"希望":   "绝望",
		"爱":     "恨",
		"信任":   "怀疑",
		"平静":   "焦虑",
		"自信":   "自卑",
		"快乐":   "悲伤",
		"勇敢":   "恐惧",
		"坚定":   "动摇",
	}

	if opp, ok := opposites[emotion]; ok {
		return opp
	}
	return "矛盾"
}

// generateStoryHook 生成故事钩子（吸引人的开头）
func (ee *EvolutionEngine) generateStoryHook(world *models.WorldSetting) string {
	var hook strings.Builder

	// 基于世界类型生成不同类型的故事钩子
	switch world.Type {
	case "fantasy":
		hook.WriteString(fmt.Sprintf("在%s的魔法世界中，", world.Name))
	case "xianxia", "wuxia":
		hook.WriteString(fmt.Sprintf("在%s的修真界中，", world.Name))
	case "scifi":
		hook.WriteString(fmt.Sprintf("在%s的未来世界里，", world.Name))
	case "historical":
		hook.WriteString(fmt.Sprintf("在%s的历史洪流中，", world.Name))
	case "urban":
		hook.WriteString(fmt.Sprintf("在%s的现代都市中，", world.Name))
	default:
		hook.WriteString(fmt.Sprintf("在%s中，", world.Name))
	}

	// 添加核心问题
	if world.Philosophy.CoreQuestion != "" {
		hook.WriteString(fmt.Sprintf("一个关于'%s'的故事即将展开。",
			world.Philosophy.CoreQuestion))
	} else {
		hook.WriteString("一个引人入胜的故事即将展开。")
	}

	// 如果有故事土壤中的情节钩子，添加一个
	if len(world.StorySoil.PotentialPlotHooks) > 0 {
		firstHook := world.StorySoil.PotentialPlotHooks[0]
		hook.WriteString(fmt.Sprintf("\n一切始于：%s", firstHook.Description))
	}

	return hook.String()
}

func (ee *EvolutionEngine) updateCharacterEmotionFromConflict(char *CharacterState, conflict *ConflictThread, stage *ConflictStage) {
	// 根据冲突更新角色情感
	char.EmotionalState.CurrentEmotion = "紧张"
	char.EmotionalState.EmotionalIntensity = min(100, char.EmotionalState.EmotionalIntensity+15)
}

func (ee *EvolutionEngine) deepenInternalConflicts(char *CharacterState, state *EvolutionState) []string {
	return []string{"更深层的内在挣扎"}
}

func (ee *EvolutionEngine) generateSecret(char *CharacterState, state *EvolutionState) string {
	return "隐藏的过去"
}

func (ee *EvolutionEngine) evolveDesireSystem(char *CharacterState, state *EvolutionState) {
	// 欲望系统演化
	char.Desires.WantVsNeedGap = "欲望与需求的差距逐渐显现"
}

// generateForeshadows 使用LLM生成伏笔
func (ee *EvolutionEngine) generateForeshadows(state *EvolutionState) []*Foreshadow {
	// 构建提示词
	prompt := ee.buildForeshadowPrompt(state)
	systemPrompt := `你是一位专业的故事策划师，擅长设计精妙的伏笔。
好的伏笔应该在回顾时让人恍然大悟，但首次阅读时不会明显。`

	result, err := ee.callWithRetry(prompt, systemPrompt)
	if err != nil {
		// LLM失败时返回默认伏笔
		return ee.createDefaultForeshadows(state)
	}

	// 解析LLM输出
	var foreshadowData struct {
		Foreshadows []struct {
			Type        string   `json:"type"`
			Content     string   `json:"content"`
			Subtlety    int      `json:"subtlety"`
			PayoffHint  string   `json:"payoff_hint"`
			Theme       string   `json:"theme"`
		} `json:"foreshadows"`
	}

	if err := json.Unmarshal([]byte(result), &foreshadowData); err != nil {
		extracted := extractJSON(result)
		json.Unmarshal([]byte(extracted), &foreshadowData)
	}

	foreshadows := make([]*Foreshadow, 0)
	for _, f := range foreshadowData.Foreshadows {
		// 计算回收轮次（当前轮次+3到+6轮）
		payoffRound := state.CurrentRound + 3 + (len(foreshadows) % 4)

		foreshadow := &Foreshadow{
			ID:          db.GenerateID("foreshadow"),
			Type:        f.Type,
			Content:     f.Content,
			PlantRound:  state.CurrentRound,
			PlantScene:  fmt.Sprintf("第%d章某场景", state.CurrentRound+1),
			PayoffRound: payoffRound,
			PayoffScene: fmt.Sprintf("第%d章某场景", payoffRound+1),
			Subtlety:    f.Subtlety,
			IsPlanted:   false,
			IsPaidOff:   false,
			RelatedThemes: []string{f.Theme},
		}
		foreshadows = append(foreshadows, foreshadow)
	}

	if len(foreshadows) == 0 {
		return ee.createDefaultForeshadows(state)
	}

	return foreshadows
}

// createDefaultForeshadows 创建默认伏笔
func (ee *EvolutionEngine) createDefaultForeshadows(state *EvolutionState) []*Foreshadow {
	foreshadows := make([]*Foreshadow, 0)
	foreshadow := &Foreshadow{
		ID:          db.GenerateID("foreshadow"),
		Type:        "symbolic",
		Content:     "象征性的伏笔",
		PlantRound:  state.CurrentRound,
		PlantScene:  fmt.Sprintf("第%d章", state.CurrentRound+1),
		PayoffRound: state.CurrentRound + 4,
		PayoffScene: fmt.Sprintf("第%d章", state.CurrentRound+5),
		Subtlety:    70,
		IsPlanted:   false,
		IsPaidOff:   false,
		RelatedThemes: []string{state.ThemeEvolution.CoreTheme},
	}
	foreshadows = append(foreshadows, foreshadow)
	return foreshadows
}

// buildForeshadowPrompt 构建伏笔生成提示词
func (ee *EvolutionEngine) buildForeshadowPrompt(state *EvolutionState) string {
	var prompt strings.Builder

	prompt.WriteString("# 伏笔设计任务\n\n")

	prompt.WriteString("## 故事背景\n")
	prompt.WriteString(fmt.Sprintf("- 核心主题: %s\n", state.ThemeEvolution.CoreTheme))
	prompt.WriteString(fmt.Sprintf("- 当前轮次: %d\n", state.CurrentRound))

	// 已有冲突
	if len(state.Conflicts) > 0 {
		prompt.WriteString("\n## 核心冲突\n")
		for i, c := range state.Conflicts {
			prompt.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, c.Type, c.CoreQuestion))
		}
	}

	// 已有角色
	if len(state.Characters) > 0 {
		prompt.WriteString("\n## 角色\n")
		for _, char := range state.Characters {
			prompt.WriteString(fmt.Sprintf("- %s: 秘密=%v\n", char.Name, len(char.Secrets) > 0))
		}
	}

	// 潜在情节钩子（来自故事土壤）
	if len(state.WorldContext.StorySoil.PotentialPlotHooks) > 0 {
		prompt.WriteString("\n## 可利用的情节钩子\n")
		for i, hook := range state.WorldContext.StorySoil.PotentialPlotHooks {
			prompt.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, hook.Type, hook.Description))
		}
	}

	// 历史事件（可作为伏笔回收的历史依据）
	if len(state.WorldContext.History.Events) > 0 {
		prompt.WriteString("\n## 可引用的历史事件\n")
		for i, event := range state.WorldContext.History.Events {
			if i < 5 { // 最多列出5个
				prompt.WriteString(fmt.Sprintf("- %s (%s): %s\n", event.Name, event.Time, event.Description))
			}
		}
	}

	// 历史遗产（可作为深层伏笔）
	if len(state.WorldContext.History.Legacies) > 0 {
		prompt.WriteString("\n## 历史遗产（可作深层伏笔）\n")
		for _, legacy := range state.WorldContext.History.Legacies {
			prompt.WriteString(fmt.Sprintf("- %s\n", legacy))
		}
	}

	// 未解决问题（伏笔回收的目标）
	if len(state.WorldContext.StorySoil.HistoricalContext.UnresolvedIssues) > 0 {
		prompt.WriteString("\n## 待解决的历史问题\n")
		for _, issue := range state.WorldContext.StorySoil.HistoricalContext.UnresolvedIssues {
			prompt.WriteString(fmt.Sprintf("- %s\n", issue))
		}
	}

	prompt.WriteString("\n# 任务\n")
	prompt.WriteString("设计3-5个伏笔，要求：\n")
	prompt.WriteString("1. 伏笔类型多样（象征式/对话式/情节式/角色式）\n")
	prompt.WriteString("2. 隐蔽程度（subtlety）0-100，越高越难发现\n")
	prompt.WriteString("3. 伏笔与核心主题相关\n")
	prompt.WriteString("4. 提示伏笔的回收方向（payoff_hint）\n")
	prompt.WriteString("5. 伏笔应该是可回顾验证的\n")
	prompt.WriteString("6. 考虑利用历史事件和情节钩子作为伏笔基础\n")

	prompt.WriteString("\n# 输出格式（JSON）\n")
	prompt.WriteString(`{
  "foreshadows": [
    {
      "type": "象征式/对话式/情节式/角色式",
      "content": "伏笔内容描述",
      "subtlety": 0-100,
      "payoff_hint": "如何回收这个伏笔？",
      "theme": "关联的主题"
    }
  ]
}`)

	return prompt.String()
}

// deepenTheme 使用LLM深化主题
func (ee *EvolutionEngine) deepenTheme(state *EvolutionState) []ThematicLayer {
	// 构建提示词
	prompt := ee.buildThemeDeepenPrompt(state)
	systemPrompt := `你是一位专业的故事策划师，擅长主题设计和哲学思考。
好的故事应该在娱乐之余传达深刻的思考。`

	result, err := ee.callWithRetry(prompt, systemPrompt)
	if err != nil {
		// LLM失败时返回默认主题层次
		return []ThematicLayer{
			{
				Layer:     "deep",
				Expression: fmt.Sprintf("第%d章主题探索", state.CurrentRound),
				Chapter:   state.CurrentRound,
				Deepened:   true,
			},
		}
	}

	// 解析LLM输出
	var themeData struct {
		Layers []struct {
			Layer      string `json:"layer"`
			Expression string `json:"expression"`
			Depth      string `json:"depth"`
		} `json:"layers"`
	}

	if err := json.Unmarshal([]byte(result), &themeData); err != nil {
		extracted := extractJSON(result)
		json.Unmarshal([]byte(extracted), &themeData)
	}

	layers := make([]ThematicLayer, 0)
	for _, l := range themeData.Layers {
		layer := ThematicLayer{
			Layer:      l.Depth,
			Expression: l.Expression,
			Chapter:    state.CurrentRound,
			Deepened:   true,
		}
		layers = append(layers, layer)
	}

	if len(layers) == 0 {
		return []ThematicLayer{
			{
				Layer:     "deep",
				Expression: fmt.Sprintf("第%d章主题探索", state.CurrentRound),
				Chapter:   state.CurrentRound,
				Deepened:   true,
			},
		}
	}

	return layers
}

// buildThemeDeepenPrompt 构建主题深化提示词
func (ee *EvolutionEngine) buildThemeDeepenPrompt(state *EvolutionState) string {
	var prompt strings.Builder

	prompt.WriteString("# 主题深化任务\n\n")

	prompt.WriteString("## 故事背景\n")
	prompt.WriteString(fmt.Sprintf("- 核心主题: %s\n", state.ThemeEvolution.CoreTheme))
	prompt.WriteString(fmt.Sprintf("- 当前轮次: %d\n", state.CurrentRound))
	prompt.WriteString(fmt.Sprintf("- 叙事深度: %d/10\n", state.NarrativeDepth))

	// 已有主题层次
	if len(state.ThemeEvolution.ThematicLayers) > 0 {
		prompt.WriteString("\n## 已有主题层次\n")
		for i, l := range state.ThemeEvolution.ThematicLayers {
			prompt.WriteString(fmt.Sprintf("%d. [%s] %s (第%d章)\n", i+1, l.Layer, l.Expression, l.Chapter))
		}
	}

	prompt.WriteString("\n# 任务\n")
	prompt.WriteString("深化主题表达，要求：\n")
	prompt.WriteString("1. 设计2-3个新的主题层次\n")
	prompt.WriteString("2. 层次应逐步深入（surface → middle → deep → philosophical）\n")
	prompt.WriteString("3. 每个层次有具体的表达方式\n")
	prompt.WriteString("4. 与核心主题保持一致\n")

	prompt.WriteString("\n# 输出格式（JSON）\n")
	prompt.WriteString(`{
  "layers": [
    {
      "layer": "层次描述",
      "expression": "如何在故事中表达这个层次？",
      "depth": "surface/middle/deep/philosophical"
    }
  ]
}`)

	return prompt.String()
}

// generatePlotTwist 使用LLM生成情节转折
func (ee *EvolutionEngine) generatePlotTwist(state *EvolutionState) string {
	// 构建提示词
	prompt := ee.buildPlotTwistPrompt(state)
	systemPrompt := `你是一位专业的故事策划师，擅长设计令人震惊但合理的情节转折。
最好的情节转折是回顾时发现一切早有暗示。`

	result, err := ee.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return "意外的情节转折"
	}

	// 解析LLM输出
	var twistData struct {
		Twist string `json:"twist"`
	}

	if err := json.Unmarshal([]byte(result), &twistData); err != nil {
		extracted := extractJSON(result)
		json.Unmarshal([]byte(extracted), &twistData)
	}

	if twistData.Twist != "" {
		return twistData.Twist
	}

	return "意外的情节转折"
}

// buildPlotTwistPrompt 构建情节转折提示词
func (ee *EvolutionEngine) buildPlotTwistPrompt(state *EvolutionState) string {
	var prompt strings.Builder

	prompt.WriteString("# 情节转折设计任务\n\n")

	prompt.WriteString("## 故事背景\n")
	prompt.WriteString(fmt.Sprintf("- 核心主题: %s\n", state.ThemeEvolution.CoreTheme))
	prompt.WriteString(fmt.Sprintf("- 当前轮次: %d\n", state.CurrentRound))

	// 已有冲突
	if len(state.Conflicts) > 0 {
		prompt.WriteString("\n## 核心冲突\n")
		for i, c := range state.Conflicts {
			if !c.IsResolved {
				prompt.WriteString(fmt.Sprintf("%d. %s: %s (强度:%d)\n", i+1, c.Type, c.CoreQuestion, c.CurrentIntensity))
			}
		}
	}

	// 已有伏笔
	plantedForeshadows := 0
	for _, f := range state.Foreshadowing {
		if f.IsPlanted && !f.IsPaidOff {
			plantedForeshadows++
		}
	}
	if plantedForeshadows > 0 {
		prompt.WriteString(fmt.Sprintf("\n## 可用伏笔\n"))
		prompt.WriteString(fmt.Sprintf("有%d个已种下但未回收的伏笔可利用\n", plantedForeshadows))
	}

	// 近期事件（可作为转折触发点）
	if len(state.WorldContext.StorySoil.HistoricalContext.RecentEvents) > 0 {
		prompt.WriteString("\n## 近期事件（可作为转折触发点）\n")
		for _, event := range state.WorldContext.StorySoil.HistoricalContext.RecentEvents {
			prompt.WriteString(fmt.Sprintf("- %s: %s (%d年前)\n", event.Event, event.Impact, event.YearsAgo))
		}
	}

	// 集体记忆（影响角色行为）
	if state.WorldContext.StorySoil.HistoricalContext.CollectiveMemory != "" {
		prompt.WriteString("\n## 集体记忆\n")
		prompt.WriteString(state.WorldContext.StorySoil.HistoricalContext.CollectiveMemory)
		prompt.WriteString("\n")
	}

	// 潜在情节钩子
	if len(state.WorldContext.StorySoil.PotentialPlotHooks) > 0 {
		prompt.WriteString("\n## 可利用的情节钩子\n")
		for i, hook := range state.WorldContext.StorySoil.PotentialPlotHooks {
			prompt.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, hook.Type, hook.Description))
		}
	}

	prompt.WriteString("\n# 任务\n")
	prompt.WriteString("设计一个情节转折，要求：\n")
	prompt.WriteString("1. 意外但合理（回顾时发现早有暗示）\n")
	prompt.WriteString("2. 显著改变故事方向\n")
	prompt.WriteString("3. 影响多个角色\n")
	prompt.WriteString("4. 与核心主题相关\n")
	prompt.WriteString("5. 考虑利用历史事件和集体记忆作为转折基础\n")

	prompt.WriteString("\n# 输出格式（JSON）\n")
	prompt.WriteString(`{
  "twist": "描述这个情节转折，包括触发点、变化和影响"
}`)

	return prompt.String()
}

// buildWorldContextSection 构建世界背景提示词部分
func (ee *EvolutionEngine) buildWorldContextSection(state *EvolutionState) string {
	var prompt strings.Builder

	prompt.WriteString("## 世界背景\n")
	prompt.WriteString(fmt.Sprintf("- 世界名称: %s\n", state.WorldContext.Name))
	prompt.WriteString(fmt.Sprintf("- 世界类型: %s\n", state.WorldContext.Type))
	prompt.WriteString(fmt.Sprintf("- 世界规模: %s\n", state.WorldContext.Scale))
	if state.WorldContext.Style != "" {
		prompt.WriteString(fmt.Sprintf("- 风格倾向: %s\n", state.WorldContext.Style))
	}
	prompt.WriteString(fmt.Sprintf("- 核心主题: %s\n", state.WorldContext.Philosophy.CoreQuestion))
	prompt.WriteString(fmt.Sprintf("- 最高善: %s, 最大恶: %s\n",
		state.WorldContext.Philosophy.ValueSystem.HighestGood,
		state.WorldContext.Philosophy.ValueSystem.UltimateEvil))

	// 道德困境（冲突来源）
	if len(state.WorldContext.Philosophy.ValueSystem.MoralDilemmas) > 0 {
		prompt.WriteString("\n### 道德困境\n")
		for i, d := range state.WorldContext.Philosophy.ValueSystem.MoralDilemmas {
			prompt.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, d.Dilemma, d.Description))
		}
	}

	// 主题列表
	if len(state.WorldContext.Philosophy.Themes) > 0 {
		prompt.WriteString("\n### 可探索主题\n")
		for _, t := range state.WorldContext.Philosophy.Themes {
			prompt.WriteString(fmt.Sprintf("- %s: %s\n", t.Name, t.ExplorationAngle))
		}
	}

	return prompt.String()
}

// buildGeographySection 构建地理信息提示词部分
func (ee *EvolutionEngine) buildGeographySection(state *EvolutionState) string {
	if len(state.WorldContext.Geography.Regions) == 0 {
		return ""
	}

	var prompt strings.Builder
	prompt.WriteString("\n## 地理环境\n")

	// 列出主要区域
	prompt.WriteString("### 主要区域\n")
	for i, region := range state.WorldContext.Geography.Regions {
		if i >= 8 { // 最多显示8个区域
			prompt.WriteString(fmt.Sprintf("... 还有%d个区域\n", len(state.WorldContext.Geography.Regions)-i))
			break
		}
		prompt.WriteString(fmt.Sprintf("- %s (%s): %s\n", region.Name, region.Type, region.Description))
		if len(region.Resources) > 0 {
			prompt.WriteString(fmt.Sprintf("  区域资源: %s\n", strings.Join(region.Resources, ", ")))
		}
		if len(region.Risks) > 0 {
			prompt.WriteString(fmt.Sprintf("  风险: %s\n", strings.Join(region.Risks, ", ")))
		}
	}

	// 全局资源分布
	if state.WorldContext.Geography.Resources != nil {
		prompt.WriteString("\n### 资源分布\n")
		if len(state.WorldContext.Geography.Resources.Basic) > 0 {
			prompt.WriteString(fmt.Sprintf("基础资源: %s\n", strings.Join(state.WorldContext.Geography.Resources.Basic, ", ")))
		}
		if len(state.WorldContext.Geography.Resources.Strategic) > 0 {
			prompt.WriteString(fmt.Sprintf("战略资源: %s\n", strings.Join(state.WorldContext.Geography.Resources.Strategic, ", ")))
		}
		if len(state.WorldContext.Geography.Resources.Rare) > 0 {
			prompt.WriteString(fmt.Sprintf("稀有资源: %s\n", strings.Join(state.WorldContext.Geography.Resources.Rare, ", ")))
		}
	}

	// 气候信息
	if state.WorldContext.Geography.Climate != nil {
		prompt.WriteString(fmt.Sprintf("\n### 气候\n%s\n", state.WorldContext.Geography.Climate.Type))
		if len(state.WorldContext.Geography.Climate.Features) > 0 {
			prompt.WriteString(fmt.Sprintf("特征: %s\n", strings.Join(state.WorldContext.Geography.Climate.Features, ", ")))
		}
	}

	return prompt.String()
}

// buildWorldviewSection 构建世界观信息提示词部分
func (ee *EvolutionEngine) buildWorldviewSection(state *EvolutionState) string {
	var prompt strings.Builder
	hasContent := false

	// 宇宙论
	if state.WorldContext.Worldview.Cosmology.Origin != "" ||
		state.WorldContext.Worldview.Cosmology.Structure != "" ||
		state.WorldContext.Worldview.Cosmology.Eschatology != "" {
		hasContent = true
		prompt.WriteString("\n## 世界观\n")
		if state.WorldContext.Worldview.Cosmology.Origin != "" {
			prompt.WriteString(fmt.Sprintf("### 宇宙论\n- 起源: %s\n", state.WorldContext.Worldview.Cosmology.Origin))
		}
		if state.WorldContext.Worldview.Cosmology.Structure != "" {
			prompt.WriteString(fmt.Sprintf("- 结构: %s\n", state.WorldContext.Worldview.Cosmology.Structure))
		}
		if state.WorldContext.Worldview.Cosmology.Eschatology != "" {
			prompt.WriteString(fmt.Sprintf("- 终极命运: %s\n", state.WorldContext.Worldview.Cosmology.Eschatology))
		}
	}

	// 形而上学
	if state.WorldContext.Worldview.Metaphysics.SoulExists ||
		state.WorldContext.Worldview.Metaphysics.FateExists {
		if !hasContent {
			prompt.WriteString("\n## 世界观\n")
			hasContent = true
		}
		prompt.WriteString("\n### 形而上学\n")
		if state.WorldContext.Worldview.Metaphysics.SoulExists {
			prompt.WriteString(fmt.Sprintf("- 灵魂存在: %s\n", state.WorldContext.Worldview.Metaphysics.SoulNature))
			prompt.WriteString(fmt.Sprintf("- 来世: %s\n", state.WorldContext.Worldview.Metaphysics.Afterlife))
		}
		if state.WorldContext.Worldview.Metaphysics.FateExists {
			prompt.WriteString(fmt.Sprintf("- 命运与自由意志: %s\n", state.WorldContext.Worldview.Metaphysics.FateRelShip))
		}
	}

	return prompt.String()
}

// buildSupernaturalSection 构建超自然体系信息提示词部分
func (ee *EvolutionEngine) buildSupernaturalSection(state *EvolutionState) string {
	if state.WorldContext.Laws.Supernatural == nil ||
		!state.WorldContext.Laws.Supernatural.Exists {
		return ""
	}

	var prompt strings.Builder
	sn := state.WorldContext.Laws.Supernatural

	prompt.WriteString("\n## 超自然体系\n")
	prompt.WriteString(fmt.Sprintf("类型: %s\n", sn.Type))

	if sn.Settings != nil {
		// 魔法体系
		if sn.Settings.MagicSystem != nil {
			prompt.WriteString("\n### 魔法体系\n")
			prompt.WriteString(fmt.Sprintf("- 来源: %s\n", sn.Settings.MagicSystem.Source))
			prompt.WriteString(fmt.Sprintf("- 代价: %s\n", sn.Settings.MagicSystem.Cost))
			if len(sn.Settings.MagicSystem.Limitation) > 0 {
				prompt.WriteString(fmt.Sprintf("- 限制: %s\n", strings.Join(sn.Settings.MagicSystem.Limitation, ", ")))
			}
		}

		// 修真体系
		if sn.Settings.CultivationSystem != nil {
			prompt.WriteString("\n### 修真体系\n")
			if len(sn.Settings.CultivationSystem.Realms) > 0 {
				prompt.WriteString(fmt.Sprintf("- 境界: %s\n", strings.Join(sn.Settings.CultivationSystem.Realms, "→")))
			}
			prompt.WriteString(fmt.Sprintf("- 资源体系: %s\n", sn.Settings.CultivationSystem.ResourceSystem))
			prompt.WriteString(fmt.Sprintf("- 瓶颈: %s\n", sn.Settings.CultivationSystem.Bottleneck))
		}

		// 异能体系
		if sn.Settings.SuperpowerSystem != nil {
			prompt.WriteString("\n### 异能体系\n")
			prompt.WriteString(fmt.Sprintf("- 起源: %s\n", sn.Settings.SuperpowerSystem.Origin))
			prompt.WriteString(fmt.Sprintf("- 类型: %s\n", sn.Settings.SuperpowerSystem.Type))
			if len(sn.Settings.SuperpowerSystem.Limit) > 0 {
				prompt.WriteString(fmt.Sprintf("- 限制: %s\n", strings.Join(sn.Settings.SuperpowerSystem.Limit, ", ")))
			}
		}
	}

	return prompt.String()
}

// buildCivilizationSection 构建文明信息提示词部分
func (ee *EvolutionEngine) buildCivilizationSection(state *EvolutionState) string {
	var prompt strings.Builder
	hasContent := false

	// 种族关系（潜在冲突来源）
	if len(state.WorldContext.Civilization.Races) > 0 {
		hasContent = true
		hasRaceRelations := false
		for _, race := range state.WorldContext.Civilization.Races {
			if len(race.Relations) > 0 {
				hasRaceRelations = true
				break
			}
		}
		if hasRaceRelations {
			prompt.WriteString("\n## 种族关系\n")
			for _, race := range state.WorldContext.Civilization.Races {
				if len(race.Relations) > 0 {
					prompt.WriteString(fmt.Sprintf("%s 与其他种族的关系:\n", race.Name))
					for otherRace, relation := range race.Relations {
						prompt.WriteString(fmt.Sprintf("  - %s: %s\n", otherRace, relation))
					}
				}
			}
		}
	}

	// 语言
	if len(state.WorldContext.Civilization.Languages) > 0 {
		if !hasContent {
			hasContent = true
			prompt.WriteString("\n## 语言文化\n")
		} else {
			prompt.WriteString("\n### 语言\n")
		}
		for _, lang := range state.WorldContext.Civilization.Languages {
			prompt.WriteString(fmt.Sprintf("- %s (%s): 使用者=%s\n", lang.Name, lang.Type, lang.Speakers))
			if len(lang.Features) > 0 {
				prompt.WriteString(fmt.Sprintf("  特征: %s\n", strings.Join(lang.Features, ", ")))
			}
		}
	}

	// 宗教（包括组织结构）
	if len(state.WorldContext.Civilization.Religions) > 0 {
		if !hasContent {
			prompt.WriteString("\n## 宗教信仰\n")
			hasContent = true
		} else {
			prompt.WriteString("\n### 宗教信仰\n")
		}
		for _, rel := range state.WorldContext.Civilization.Religions {
			prompt.WriteString(fmt.Sprintf("- %s (%s)\n", rel.Name, rel.Type))
			if rel.Cosmology != "" {
				prompt.WriteString(fmt.Sprintf("  宇宙观: %s\n", rel.Cosmology))
			}
			if len(rel.Ethics) > 0 {
				prompt.WriteString(fmt.Sprintf("  伦理: %s\n", strings.Join(rel.Ethics, ", ")))
			}
			if len(rel.Practices) > 0 {
				prompt.WriteString(fmt.Sprintf("  仪式: %s\n", strings.Join(rel.Practices, ", ")))
			}
			// 宗教组织（权力结构）
			if rel.Organization != nil {
				prompt.WriteString(fmt.Sprintf("  组织: %s (领导者: %s)\n", rel.Organization.Type, rel.Organization.Leader))
				if len(rel.Organization.Factions) > 0 {
					prompt.WriteString(fmt.Sprintf("  派系: %s\n", strings.Join(rel.Organization.Factions, ", ")))
				}
			}
		}
	}

	return prompt.String()
}

// buildHistoryDetailsSection 构建历史详情提示词部分
func (ee *EvolutionEngine) buildHistoryDetailsSection(state *EvolutionState) string {
	var prompt strings.Builder
	hasContent := false

	// 世界起源
	if state.WorldContext.History.Origin != "" {
		hasContent = true
		prompt.WriteString("\n## 世界起源\n")
		prompt.WriteString(state.WorldContext.History.Origin)
		prompt.WriteString("\n")
	}

	// 历史时代
	if len(state.WorldContext.History.Eras) > 0 {
		if !hasContent {
			prompt.WriteString("\n## 历史时代\n")
			hasContent = true
		} else {
			prompt.WriteString("\n### 历史时代\n")
		}
		for _, era := range state.WorldContext.History.Eras {
			prompt.WriteString(fmt.Sprintf("- %s (%s): %s\n", era.Name, era.Period, era.Description))
		}
	}

	return prompt.String()
}

// 质量评估方法
func (ee *EvolutionEngine) evaluateCharacterQuality(state *EvolutionState) int {
	return 70 + len(state.Characters)*3
}

func (ee *EvolutionEngine) evaluateConflictQuality(state *EvolutionState) int {
	totalIntensity := 0
	for _, c := range state.Conflicts {
		totalIntensity += c.CurrentIntensity
	}
	return min(100, 50+totalIntensity/len(state.Conflicts)/2)
}

func (ee *EvolutionEngine) evaluateForeshadowQuality(state *EvolutionState) int {
	return 60 + len(state.Foreshadowing)*2
}

func (ee *EvolutionEngine) evaluateThemeQuality(state *EvolutionState) int {
	return 60 + state.NarrativeDepth*4
}

func (ee *EvolutionEngine) evaluatePlotQuality(state *EvolutionState) int {
	return 70
}

func (ee *EvolutionEngine) calculateOverallQuality(state *EvolutionState) int {
	return (ee.evaluateCharacterQuality(state) +
		ee.evaluateConflictQuality(state) +
		ee.evaluateForeshadowQuality(state) +
		ee.evaluateThemeQuality(state)) / 4
}

func (ee *EvolutionEngine) buildEvolutionPrompt(state *EvolutionState) string {
	return fmt.Sprintf("分析当前叙事状态并提出改进建议。轮次: %d/%d, 角色: %d, 冲突: %d",
		state.CurrentRound, state.MaxRounds, len(state.Characters), len(state.Conflicts))
}

func (ee *EvolutionEngine) callWithRetry(prompt, systemPrompt string) (string, error) {
	fmt.Println("\n========== LLM DEBUG [EVOLUTION] (JSON) ==========")
	fmt.Printf("System Prompt:\n%s\n\n", systemPrompt)
	fmt.Printf("User Prompt:\n%s\n", truncateForDebugEvo(prompt, 2000))
	fmt.Println("====================================================")

	fmt.Println("🔄 调用LLM...")
	startTime := time.Now()

	result, err := ee.client.GenerateJSONWithParams(
		prompt,
		systemPrompt,
		ee.mapping.Temperature,
		ee.mapping.MaxTokens,
	)

	elapsed := time.Since(startTime)
	fmt.Printf("⏱️  耗时: %.1f秒\n", elapsed.Seconds())

	if err != nil {
		fmt.Printf("❌ 调用失败: %v\n", err)
		fmt.Println("====================================================\n")
		return "", err
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("❌ 序列化失败: %v\n", err)
		fmt.Println("====================================================\n")
		return "", err
	}

	fmt.Printf("✅ 响应成功\n")
	fmt.Printf("Response:\n%s\n", truncateForDebugEvo(string(jsonBytes), 3000))
	fmt.Println("====================================================\n")

	return string(jsonBytes), nil
}

// truncateForDebugEvo 截断过长的调试输出（演化引擎版本）
func truncateForDebugEvo(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (截断，总长度: " + fmt.Sprintf("%d", len(s)) + " 字符)"
}

func (ee *EvolutionEngine) characterStateSlice(m map[string]*CharacterState) []*CharacterState {
	result := make([]*CharacterState, 0, len(m))
	for _, c := range m {
		result = append(result, c)
	}
	return result
}

// ============================================
// 新增数据结构
// ============================================

// StoryArchitecture 故事架构
type StoryArchitecture struct {
	NarrativeMode    string   `json:"narrative_mode"`    // 叙事模式：群像剧/个人成长/对抗抽象力量
	CoreConflictType string   `json:"core_conflict_type"` // 核心冲突类型
	CharacterRoster  CharacterRosterSpec `json:"character_roster"`  // 角色阵容规划
	MainDirection    string   `json:"main_direction"`    // 主要矛盾方向
	ExpectedEnding   string   `json:"expected_ending"`   // 预期结局方向
}

// CharacterRosterSpec 角色阵容规划
type CharacterRosterSpec struct {
	TotalCharacters    int      `json:"total_characters"`    // 总角色数
	ProtagonistCount   int      `json:"protagonist_count"`   // 主角数量
	AntagonistCount    int      `json:"antagonist_count"`    // 反派数量
	SupportingCount    int      `json:"supporting_count"`    // 配角数量
	NetworkStructure   string   `json:"network_structure"`   // 网络结构：星形/网状/链式
	KeyRelationships   []string `json:"key_relationships"`   // 关键关系描述
}

// RelationshipNetwork 关系网络
type RelationshipNetwork struct {
	Nodes      map[string]*CharacterState           `json:"nodes"`       // 所有角色
	Edges      map[string]*Relationship             `json:"edges"`       // 关系边（双向）
	NetworkType string                               `json:"network_type"` // 网络类型
	CenterNode string                               `json:"center_node"` // 中心节点（主角）
	Evolution  []*RelationshipEvolutionStage        `json:"evolution"`   // 关系演化历史
}

// Relationship 关系
type Relationship struct {
	From          string  `json:"from"`
	To            string  `json:"to"`
	Type          string  `json:"type"`          // "盟友"/"对手"/"复杂"/"师生"/"亲情"/"爱情"
	Tension       int     `json:"tension"`       // 0-100，紧张度
	Potential     string  `json:"potential"`     // 这个关系可能如何发展
	CurrentState  string  `json:"current_state"` // 当前状态描述
}

// RelationshipEvolutionStage 关系演化阶段
type RelationshipEvolutionStage struct {
	Stage     int    `json:"stage"`     // 阶段序号
	Change     string `json:"change"`     // 发生了什么变化
	Trigger    string `json:"trigger"`    // 触发事件
	ResultState map[string]string `json:"result_state"` // 演化后的关系状态
}

// ForeshadowPlan 伏笔计划
type ForeshadowPlan struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`        // "symbolic"/"dialogue"/"event"/"character"
	Content     string  `json:"content"`     // 伏笔的内容
	
	// 种植信息
	PlantChapter    int    `json:"plant_chapter"`
	PlantScene      int    `json:"plant_scene"`
	PlantMethod     string `json:"plant_method"`    // 如何植入
	Subtlety        int    `json:"subtlety"`        // 0-100，含蓄程度
	
	// 回收信息
	PayoffChapter   int    `json:"payoff_chapter"`
	PayoffScene     int    `json:"payoff_scene"`
	PayoffMethod    string `json:"payoff_method"`    // 如何揭示
	
	// 连接逻辑
	Connection      string `json:"connection"`       // 这个伏笔如何连接种植和回收
	
	// 状态追踪
	IsPlanted       bool   `json:"is_planted"`
	IsPaidOff       bool   `json:"is_paid_off"`
}

// GlobalOutline 全局大纲（关键事件序列）
type GlobalOutline struct {
	Opening     string            `json:"opening"`     // 开局
	KeyEvents   []KeyEvent        `json:"key_events"`   // 关键事件序列
	Climax      string            `json:"climax"`      // 高潮
	Resolution  string            `json:"resolution"`  // 结局
	ForeshadowLinks map[string]string `json:"foreshadow_links"` // 伏笔链接（事件ID -> 伏笔ID）
}

// KeyEvent 关键事件
type KeyEvent struct {
	ID           string   `json:"id"`
	Sequence     int      `json:"sequence"`     // 事件顺序
	Name         string   `json:"name"`         // 事件名称
	Description  string   `json:"description"`  // 事件描述
	InvolvedCharacters []string `json:"involved_characters"` // 涉及的角色
	Purpose      string   `json:"purpose"`      // 事件目的
	Foreshadows  []string `json:"foreshadows"`  // 种植的伏笔ID
	Reveals      []string `json:"reveals"`      // 回收的伏笔ID
}

// ChapterPlan 章节规划
type ChapterPlan struct {
	TotalChapters    int                    `json:"total_chapters"`
	ChapterSequence  []ChapterSynopsis      `json:"chapter_sequence"` // 章节序列
}

// ChapterSynopsis 章节概要
type ChapterSynopsis struct {
	Chapter      int      `json:"chapter"`
	Title        string   `json:"title"`
	Purpose      string   `json:"purpose"`
	KeyEvents    []string `json:"key_events"`   // 本章包含的关键事件
	RelationshipChanges []string `json:"relationship_changes"` // 预期的关系变化
	ForeshadowOps ForeshadowOperations `json:"foreshadow_ops"` // 伏笔操作
}

// ForeshadowOperations 伏笔操作
type ForeshadowOperations struct {
	Plant  []ForeshadowPlantOp  `json:"plant"`  // 种植操作
	Payoff  []ForeshadowPayoffOp `json:"payoff"` // 回收操作
}

// ForeshadowPlantOp 种植操作
type ForeshadowPlantOp struct {
	ForeshadowID string `json:"foreshadow_id"`
	Scene         int    `json:"scene"`
	Method        string `json:"method"`
}

// ForeshadowPayoffOp 回收操作
type ForeshadowPayoffOp struct {
	ForeshadowID string `json:"foreshadow_id"`
	Scene         int    `json:"scene"`
	Method        string `json:"method"`
}

// CharacterEvolutionTracker 角色演化追踪器
type CharacterEvolutionTracker struct {
	CharacterID        string                   `json:"character_id"`

	// 情感轨迹
	EmotionalJourney   []EmotionalState         `json:"emotional_journey"`

	// 关系轨迹
	RelationshipHistory map[string][]RelationshipHistoryEntry `json:"relationship_history"`

	// 知识轨迹
	KnowledgeGrowth    []KnowledgePiece          `json:"knowledge_growth"`

	// 内在冲突演化
	InternalConflictProgress []ConflictStage    `json:"internal_conflict_progress"`

	// 关键转折点
	TurningPoints      []TurningPoint            `json:"turning_points"`

	// 本章变化（在细纲生成时填充）
	ChapterChanges map[string]*ChapterCharacterChange `json:"chapter_changes"`
}

// EmotionalState 情感状态
type EmotionalState struct {
	Emotion     string `json:"emotion"`
	Intensity   int    `json:"intensity"`   // 0-100
	Context     string `json:"context"`     // 触发原因
	Timestamp   int    `json:"timestamp"`   // 第几轮演化
}

// RelationshipHistoryEntry 关系历史条目（用于追踪关系演化）
type RelationshipHistoryEntry struct {
	Type        string `json:"type"`
	Tension     int    `json:"tension"`
	Description string `json:"description"`
	Timestamp   int    `json:"timestamp"`
}

// KnowledgePiece 知识片段
type KnowledgePiece struct {
	Content     string `json:"content"`
	Source      string `json:"source"`   // 从哪里得知的
	Importance  int    `json:"importance"` // 1-10
	Timestamp   int    `json:"timestamp"`
}

// TurningPoint 转折点
type TurningPoint struct {
	Chapter     int    `json:"chapter"`
	Scene       int    `json:"scene"`
	Event       string `json:"event"`
	Significance string `json:"significance"`
}

// ChapterCharacterChange 章节角色变化
type ChapterCharacterChange struct {
	Chapter         int                      `json:"chapter"`
	EmotionalChange  string                   `json:"emotional_change"`
	NewKnowledge     []string                 `json:"new_knowledge"`
	RelationshipChanges map[string]string   `json:"relationship_changes"`
	InternalConflict   string                   `json:"internal_conflict"`
}
