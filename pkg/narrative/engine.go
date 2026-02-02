// Package narrative 叙事器
// 负责故事大纲、章节规划、场景序列和角色弧光规划
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

// NarrativeStructure 叙事结构类型
type NarrativeStructure string

const (
	StructureThreeAct       NarrativeStructure = "three_act"        // 三幕剧结构
	StructureHerosJourney   NarrativeStructure = "heros_journey"    // 英雄之旅
	StructureSaveTheCat     NarrativeStructure = "save_the_cat"     // 救猫咪节拍表
	StructureKishotenketsu  NarrativeStructure = "kishotenketsu"     // 起承转合
	StructureFreytagPyramid NarrativeStructure = "freytag_pyramid"  // 弗赖塔格金字塔
)

// CreateParams 创建叙事蓝图参数
type CreateParams struct {
	WorldID    string `json:"world_id"`    // 世界ID
	StoryType  string `json:"story_type"`  // 故事类型
	Theme      string `json:"theme"`       // 核心主题
	Protagonist string `json:"protagonist"` // 主角概念
	Length     string `json:"length"`      // 篇幅预期：short/medium/long
	ChapterCount int  `json:"chapter_count"` // 章节数量（可选）
	Structure   NarrativeStructure `json:"structure"` // 叙事结构（可选，默认三幕剧）
}

// OutlineInput 生成大纲输入
type OutlineInput struct {
	WorldSummary string `json:"world_summary"`
	StoryType    string `json:"story_type"`
	Theme        string `json:"theme"`
	Protagonist  string `json:"protagonist"`
	Length       string `json:"length"`
	Structure    NarrativeStructure `json:"structure"`
}

// ChapterPlanInput 生成章节规划输入
type ChapterPlanInput struct {
	Outline      string `json:"outline"`
	ChapterCount int    `json:"chapter_count"`
}

// SceneInput 生成场景序列输入
type SceneInput struct {
	Chapter        string `json:"chapter"`
	ChapterPurpose string `json:"chapter_purpose"`
	PreviousSummary string `json:"previous_summary"`
	CharacterStates string `json:"character_states"`
}

// CharacterArcInput 角色弧光输入
type CharacterArcInput struct {
	CharacterInfo string `json:"character_info"`
	Theme         string `json:"theme"`
	StoryType     string `json:"story_type"`
}

// ============================================
// 冲突系统
// ============================================

// CoreConflict 核心冲突
type CoreConflict struct {
	Type           string   `json:"type"`            // 人与人/与社会/与自己/与自然
	Description    string   `json:"description"`     // 冲突描述
	EscalationPath []string `json:"escalation_path"` // 冲突升级路径
	Resolution     string   `json:"resolution"`      // 冲突解决方式
}

// ============================================
// 多种叙事结构输出
// ============================================

// OutlineOutput 大纲输出（通用结构）
type OutlineOutput struct {
	StructureType  NarrativeStructure `json:"structure_type"`
	ThreeAct       *ThreeActOutput   `json:"three_act,omitempty"`
	HerosJourney   *HerosJourneyOutput `json:"heros_journey,omitempty"`
	SaveTheCat     *SaveTheCatOutput `json:"save_the_cat,omitempty"`
	Kishotenketsu  *KishotenketsuOutput `json:"kishotenketsu,omitempty"`
	FreytagPyramid *FreytagPyramidOutput `json:"freytag_pyramid,omitempty"`
	CoreConflicts  []CoreConflict    `json:"core_conflicts"`
}

// ThreeActOutput 三幕剧结构输出
type ThreeActOutput struct {
	Act1 Act1Output `json:"act1"`
	Act2 Act2Output `json:"act2"`
	Act3 Act3Output `json:"act3"`
}

type Act1Output struct {
	Setup            string `json:"setup"`
	IncitingIncident  string `json:"inciting_incident"`
	PlotPoint1       string `json:"plot_point1"`
}

type Act2Output struct {
	RisingAction []string `json:"rising_action"`
	Midpoint     string   `json:"midpoint"`
	AllIsLost    string   `json:"all_is_lost"`
	PlotPoint2   string   `json:"plot_point2"`
}

type Act3Output struct {
	Climax     string `json:"climax"`
	Resolution string `json:"resolution"`
}

// HerosJourneyOutput 英雄之旅结构输出（坎贝尔12阶段）
type HerosJourneyOutput struct {
	OrdinaryWorld      string `json:"ordinary_world"`       // 1. 平凡世界
	CallToAdventure    string `json:"call_to_adventure"`     // 2. 冒险召唤
	Refusal            string `json:"refusal"`                // 3. 拒绝召唤
	MeetingMentor      string `json:"meeting_mentor"`         // 4. 遇见导师
	CrossingThreshold  string `json:"crossing_threshold"`     // 5. 跨越第一道门槛
	TestsAllies        string `json:"tests_allies"`           // 6. 试炼、盟友、敌人
	ApproachInmostCave string `json:"approach_inmost_cave"`   // 7. 接近最深处的洞穴
	Ordeal             string `json:"ordeal"`                  // 8. 严峻考验
	Reward             string `json:"reward"`                  // 9. 奖赏
	TheRoadBack        string `json:"the_road_back"`          // 10. 归途
	Resurrection       string `json:"resurrection"`            // 11. 复活
	ReturnWithElixir   string `json:"return_with_elixir"`     // 12. 带着灵药回归
}

// SaveTheCatOutput 救猫咪节拍表输出（布莱克·斯奈德）
type SaveTheCatOutput struct {
	OpeningImage   string `json:"opening_image"`    // 1. 开篇画面
	ThemeStated    string `json:"theme_stated"`     // 2. 主题陈述
	SetUp          string `json:"set_up"`           // 3. 铺垫
	Catalyst       string `json:"catalyst"`         // 4. 触发事件
	Debate         string `json:"debate"`           // 5. 争论
	BreakIntoTwo   string `json:"break_into_two"`   // 6. 第二幕衔接点
	BStory         string `json:"b_story"`          // 7. B故事
	FunAndGames    string `json:"fun_and_games"`    // 8. 游戏时间
	Midpoint       string `json:"midpoint"`         // 9. 中点
	BadGuysCloseIn string `json:"bad_guys_close_in"` // 10. 坏人逼近
	AllIsLost      string `json:"all_is_lost"`      // 11. 一无所有
	DarkNightOfSoul string `json:"dark_night_of_soul"` // 12. 灵魂黑夜
	BreakIntoThree string `json:"break_into_three"` // 13. 第三幕衔接点
	Finale         string `json:"finale"`           // 14. 终局
	FinalImage     string `json:"final_image"`      // 15. 结束画面
}

// KishotenketsuOutput 起承转合结构输出（东方叙事）
type KishotenketsuOutput struct {
	Ki   string `json:"ki"`   // 起：介绍角色和设定
	Sho  string `json:"sho"`  // 承：发展事件和 complication
	Ten  string `json:"ten"`  // 转：转折点，改变方向
	Ketsu string `json:"ketsu"` // 合：结局，收束所有线索
}

// FreytagPyramidOutput 弗赖塔格金字塔输出
type FreytagPyramidOutput struct {
	Exposition    string `json:"exposition"`     // 说明：介绍背景
	IncitingIncident string `json:"inciting_incident"` // 激发事件
	RisingAction  string `json:"rising_action"`  // 上升动作：一系列事件
	Climax        string `json:"climax"`         // 高潮：转折点
	FallingAction string `json:"falling_action"` // 下降动作：后果
	Resolution    string `json:"resolution"`     // 结局：解决
}

// ChapterPlanOutput 章节规划输出
type ChapterPlanOutput struct {
	Chapters []ChapterPlanItem `json:"chapters"`
}

type ChapterPlanItem struct {
	Chapter         int      `json:"chapter"`
	Title           string   `json:"title"`
	Purpose         string   `json:"purpose"`
	KeyScenes       []string `json:"key_scenes"`
	PlotAdvancement string   `json:"plot_advancement"`
	ArcProgress     string   `json:"arc_progress"`
	EndingHook      string   `json:"ending_hook"`
	EstimatedWords  int      `json:"estimated_words"`
}

// SceneOutput 场景输出
type SceneOutput struct {
	Scenes []SceneItem `json:"scenes"`
}

type SceneItem struct {
	Sequence       int      `json:"sequence"`
	Purpose        string   `json:"purpose"`
	Location       string   `json:"location"`
	Characters     []string `json:"characters"`
	Action         string   `json:"action"`
	DialogueFocus  string   `json:"dialogue_focus"`
	Mood           string   `json:"mood"`
	ExpectedLength int      `json:"expected_length"`
}

// CharacterArcOutput 角色弧光输出
type CharacterArcOutput struct {
	ArcType       string           `json:"arc_type"`
	StartState    CharacterStateIO `json:"start_state"`
	EndState      CharacterStateIO `json:"end_state"`
	TurningPoints []TurningPointIO `json:"turning_points"`
}

type CharacterStateIO struct {
	Personality []string `json:"personality"`
	Motivation  string   `json:"motivation"`
	Emotion     string   `json:"emotion"`
}

type TurningPointIO struct {
	Chapter int    `json:"chapter"`
	Event   string `json:"event"`
	Change  string `json:"change"`
}

// NarrativeEngine 叙事器（系统的大脑）
type NarrativeEngine struct {
	db      db.Database
	cfg     *config.Config
	client  *llm.Client
	mapping *config.ModuleMapping
	evolution *EvolutionEngine // 演化引擎
}

// New 创建叙事器
func New() (*NarrativeEngine, error) {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 创建LLM客户端
	client, mapping, err := llm.NewClientForModule("narrative_engine")
	if err != nil {
		return nil, fmt.Errorf("创建LLM客户端失败: %w", err)
	}

	// 创建演化引擎
	evolution, err := NewEvolutionEngine()
	if err != nil {
		return nil, fmt.Errorf("创建演化引擎失败: %w", err)
	}

	return &NarrativeEngine{
		db:      db.Get(),
		cfg:     cfg,
		client:  client,
		mapping: mapping,
		evolution: evolution,
	}, nil
}

// EvolutionConfig 演化配置
type EvolutionConfig struct {
	EnableEvolution bool              `json:"enable_evolution"` // 是否启用动态演化
	MaxRounds       int               `json:"max_rounds"`       // 最大演化轮次
	RoundTypes      []EvolutionRound  `json:"round_types"`      // 自定义演化轮次
	AutoStopWhen    int               `json:"auto_stop_when"`   // 自动停止条件（质量分数）
}

// CreateBlueprintThroughEvolution 通过动态演化创建叙事蓝图
// 这是叙事器作为"系统大脑"的主要入口
func (ne *NarrativeEngine) CreateBlueprintThroughEvolution(params CreateParams, config EvolutionConfig) (*models.NarrativeBlueprint, *EvolutionState, error) {
	// 1. 创建初始演化状态
	evolutionState, err := ne.evolution.CreateEvolutionState(params.WorldID)
	if err != nil {
		return nil, nil, fmt.Errorf("创建演化状态失败: %w", err)
	}

	// 设置演化配置
	if config.MaxRounds > 0 {
		evolutionState.MaxRounds = config.MaxRounds
	}

	evolutionResults := make([]*EvolutionResult, 0)

	// 2. 执行多轮动态演化（仅在启用时）
	if config.EnableEvolution {
		roundTypes := config.RoundTypes
		if len(roundTypes) == 0 {
			// 默认演化序列
			roundTypes = []EvolutionRound{
				RoundCharacterCreation,  // 角色创建
				RoundConflictDesign,     // 冲突设计
				RoundCharacterDeepen,    // 角色深化
				RoundConflictEvolution,  // 冲突演化
				RoundForeshadowPlant,    // 种下伏笔
				RoundThemeDeepen,        // 主题深化
				RoundConflictEvolution,  // 冲突再演化
				RoundForeshadowWeave,    // 编织伏笔
				RoundPlotTwist,          // 情节转折
				RoundResolutionPlan,     // 结局规划
			}
		}

		for _, roundType := range roundTypes {
			if evolutionState.CurrentRound >= evolutionState.MaxRounds {
				break
			}

			result, err := ne.evolution.Evolve(evolutionState, roundType)
			if err != nil {
				return nil, nil, fmt.Errorf("演化轮次%s失败: %w", roundType, err)
			}

			evolutionResults = append(evolutionResults, result)

			// 检查自动停止条件
			if config.AutoStopWhen > 0 && result.QualityScore >= config.AutoStopWhen {
				break
			}
		}
	}

	// 3. 基于演化状态生成叙事蓝图
	blueprint := ne.buildBlueprintFromEvolution(evolutionState, params)

	// 4. 保存到数据库
	if err := ne.db.SaveNarrativeBlueprint(blueprint); err != nil {
		return nil, nil, fmt.Errorf("保存叙事蓝图失败: %w", err)
	}

	return blueprint, evolutionState, nil
}

// buildBlueprintFromEvolution 从演化状态构建叙事蓝图
func (ne *NarrativeEngine) buildBlueprintFromEvolution(state *EvolutionState, params CreateParams) *models.NarrativeBlueprint {
	blueprint := &models.NarrativeBlueprint{
		ID:        db.GenerateID("narrative"),
		WorldID:   params.WorldID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 1. 从冲突系统生成故事大纲
	fmt.Println("  📚 构建故事大纲...")
	blueprint.StoryOutline = ne.buildOutlineFromConflicts(state)
	fmt.Println("  ✓ 故事大纲完成")

	// 2. 从演化状态生成章节规划
	chapterCount := params.ChapterCount
	if chapterCount == 0 {
		chapterCount = ne.defaultChapterCount(params.Length)
	}
	fmt.Printf("  📖 生成 %d 章规划...\n", chapterCount)
	blueprint.ChapterPlans = ne.buildChapterPlansFromEvolution(state, chapterCount)
	fmt.Println("  ✓ 章节规划完成")

	// 3. 从角色状态生成场景指令
	blueprint.Scenes = ne.buildScenesFromEvolution(state, blueprint.ChapterPlans)

	// 4. 从角色情感系统生成角色弧光
	fmt.Println("  👥 构建角色弧光...")
	blueprint.CharacterArcs = ne.buildCharacterArcsFromEvolution(state)
	fmt.Println("  ✓ 角色弧光完成")

	// 5. 从主题演化生成主题计划
	fmt.Println("  🎨 构建主题规划...")
	blueprint.ThemePlan = ne.buildThemePlanFromEvolution(state)
	fmt.Println("  ✓ 主题规划完成")

	return blueprint
}

// buildOutlineFromConflicts 从冲突系统构建故事大纲
func (ne *NarrativeEngine) buildOutlineFromConflicts(state *EvolutionState) models.StoryOutline {
	// 找到主要冲突（强度最高的）
	mainConflict := state.findMainConflict()

	structure := ne.determineStructureFromConflicts(state.Conflicts)

	outline := models.StoryOutline{
		StructureType: string(structure),
	}

	// 如果没有冲突，返回默认大纲
	if len(mainConflict.EvolutionPath) == 0 {
		return ne.createDefaultOutline(state)
	}

	// 构建setup：基于世界设定和角色
	setup := ne.buildSetupFromState(state)

	// 构建第一幕
	outline.Act1 = models.Act1{
		Setup:            setup,
		IncitingIncident: mainConflict.EvolutionPath[0].Description,
		PlotPoint1:       ne.buildPlotPoint1(state, mainConflict),
	}

	// 构建第二幕
	risingAction := make([]string, 0)
	midpointIndex := len(mainConflict.EvolutionPath) / 2

	for i := 1; i < len(mainConflict.EvolutionPath); i++ {
		stage := mainConflict.EvolutionPath[i]
		if i == midpointIndex {
			// 这个阶段作为中点
			continue
		}
		risingAction = append(risingAction, stage.Description)
	}

	outline.Act2 = models.Act2{
		RisingAction: risingAction,
		Midpoint:     ne.buildMidpoint(state, mainConflict),
		AllIsLost:    ne.buildAllIsLost(state, mainConflict),
		PlotPoint2:   ne.buildPlotPoint2(state, mainConflict),
	}

	// 构建第三幕
	outline.Act3 = models.Act3{
		Climax:     ne.buildClimax(state, mainConflict),
		Resolution: ne.buildResolution(state, mainConflict),
	}

	return outline
}

// createDefaultOutline 创建默认大纲
func (ne *NarrativeEngine) createDefaultOutline(state *EvolutionState) models.StoryOutline {
	return models.StoryOutline{
		StructureType: "three_act",
		Act1: models.Act1{
			Setup:            ne.buildSetupFromState(state),
			IncitingIncident: "打破平衡的事件发生",
			PlotPoint1:       "主角踏上旅程",
		},
		Act2: models.Act2{
			RisingAction: []string{"面对挑战", "遭遇挫折", "获得成长"},
			Midpoint:     "故事的重大转折",
			AllIsLost:    "主角面临最低点",
			PlotPoint2:   "准备最终对决",
		},
		Act3: models.Act3{
			Climax:     "最终对抗",
			Resolution: "冲突得到解决，主角获得成长",
		},
	}
}

// buildSetupFromState 基于演化状态构建setup
func (ne *NarrativeEngine) buildSetupFromState(state *EvolutionState) string {
	var setup strings.Builder

	setup.WriteString(fmt.Sprintf("在%s的世界中，", state.WorldContext.Name))

	// 描述主要角色
	if len(state.Characters) > 0 {
		charNames := make([]string, 0)
		for _, char := range state.Characters {
			if len(charNames) < 3 { // 最多列出3个主角
				charNames = append(charNames, char.Name)
			}
		}
		setup.WriteString(strings.Join(charNames, "、"))
		setup.WriteString("等角色各自怀揣着不同的欲望与秘密。")
	}

	// 描述核心问题
	if state.WorldContext.Philosophy.CoreQuestion != "" {
		setup.WriteString(fmt.Sprintf("世界面临着一个根本问题：%s",
			state.WorldContext.Philosophy.CoreQuestion))
	}

	return setup.String()
}

// buildPlotPoint1 构建第一情节点
func (ne *NarrativeEngine) buildPlotPoint1(state *EvolutionState, conflict *ConflictThread) string {
	if len(conflict.Participants) == 0 {
		return "主角被迫卷入冲突，无法再置身事外"
	}

	return fmt.Sprintf("%s因%s而被迫采取行动，踏上改变的旅程",
		strings.Join(conflict.Participants, "与"),
		conflict.CoreQuestion)
}

// buildMidpoint 构建中点
func (ne *NarrativeEngine) buildMidpoint(state *EvolutionState, conflict *ConflictThread) string {
	prompt := fmt.Sprintf(`你是故事结构专家。请为故事设计"中点"（Midpoint）这一关键情节。

# 冲突信息
冲突类型：%s
核心问题：%s
当前强度：%d

# 中点（Midpoint）的定义
中点是故事中的重大转折点，通常发生在故事的一半处。在这个时刻：
- 主角对冲突有了全新的认识或发现
- 局势发生根本性变化，故事从此进入"第二幕的下半场"
- 主角可能获得重要信息、失去重要支持，或遭遇意外挫折

# 任务要求
请描述这个中点事件，必须包含：
1. **具体发生了什么**（明确的事件或发现）
2. **主角的认知变化**（主角如何重新理解冲突）
3. **局势的根本转变**（故事方向如何改变）
4. **80-150字**

# 输出格式
直接输出描述，不要前缀。

# 示例
❌ 错误：中点转折
✅ 正确：Thalric在执行任务时发现，目标人物同样保留着情感和记忆。这个发现彻底动摇了他的信念——原来"切除情感"并非唯一的出路。他开始质疑"无情者"教会的根本教义，故事从此从"如何成为无情者"转向"是否应该成为无情者"。`,
		conflict.Type, conflict.CoreQuestion, conflict.CurrentIntensity)

	response, err := ne.callLLM(prompt)
	if err != nil {
		return "中点转折：主角对冲突有了新的认识，局势发生根本变化"
	}
	result := strings.TrimSpace(response)
	if len(result) == 0 {
		return "中点转折：主角对冲突有了新的认识，局势发生根本变化"
	}
	return result
}

// buildAllIsLost 构建一无所有时刻
func (ne *NarrativeEngine) buildAllIsLost(state *EvolutionState, conflict *ConflictThread) string {
	prompt := fmt.Sprintf(`你是故事结构专家。请为故事设计"一无所有"（All Is Lost）这一关键时刻。

# 冲突信息
冲突类型：%s
核心问题：%s
当前强度：%d

# 一无所有时刻的定义
这是主角最绝望的时刻，通常发生在高潮之前。在这个时刻：
- 主角遭遇彻底失败，失去一切依靠
- 看似没有任何胜利的可能
- 主角的内心防线崩溃，绝望感达到顶峰
- 但这个绝望是"触底反弹"的前奏

# 任务要求
请描述这个时刻，必须包含：
1. **主角失去了什么**（具体的损失：人物、希望、信念等）
2. **绝望的表现**（主角如何崩溃、放弃或陷入绝望）
3. **局势的严峻性**（为什么看起来毫无希望）
4. **80-120字**

# 输出格式
直接输出描述，不要前缀。

# 示例
❌ 错误：主角失败
✅ 正确：Thalric最珍视的机械幼兽被教会无情处死，因为他违反了"保留情感"的禁令。这一刻，Thalric彻底崩溃——他努力遵守的所有规则、他压抑的所有痛苦，换来的却是失去最后的情感寄托。他蜷缩在冰冷的操作台上，第一次真正理解了"无情"的代价：那不是力量的提升，而是人性的丧失。`,
		conflict.Type, conflict.CoreQuestion, conflict.CurrentIntensity)

	response, err := ne.callLLM(prompt)
	if err != nil {
		return "冲突达到最高潮，主角面临最严峻的考验"
	}
	result := strings.TrimSpace(response)
	if len(result) == 0 {
		return "冲突达到最高潮，主角面临最严峻的考验"
	}
	return result
}

// buildPlotPoint2 构建第二情节点
func (ne *NarrativeEngine) buildPlotPoint2(state *EvolutionState, conflict *ConflictThread) string {
	prompt := fmt.Sprintf(`你是故事结构专家。请为故事设计"第二情节点"（Plot Point 2）。

# 冲突信息
冲突类型：%s
核心问题：%s
当前强度：%d

# 第二情节点的定义
第二情节点发生在"一无所有"之后、高潮之前。在这个时刻：
- 主角从绝望中找到新的希望或力量
- 主角重整旗鼓，整合所有资源和教训
- 决定进行最后的决战，不再犹豫
- 故事的节奏加速，直奔高潮

# 任务要求
请描述这个时刻，必须包含：
1. **主角找到了什么**（新的希望、新的理解、新的力量来源）
2. **重整旗鼓的过程**（主角如何整合资源、如何改变策略）
3. **决心的形成**（为什么这次决意战斗到底）
4. **80-120字**

# 输出格式
直接输出描述，不要前缀。

# 示例
❌ 错误：主角准备决战
✅ 正确：在失去机械幼兽的痛苦中，Thalric反而找到了答案——真正的力量不是"切除情感"，而是"驾驭情感"。他回忆起所有被压抑的痛苦时刻，意识到这些痛苦正是让他成为人的原因。他站起身，第一次不是试图成为"无情者"，而是作为"有情者"迎接战斗。`,
		conflict.Type, conflict.CoreQuestion, conflict.CurrentIntensity)

	response, err := ne.callLLM(prompt)
	if err != nil {
		return "主角重整旗鼓，整合所有资源，准备进行最终对抗"
	}
	result := strings.TrimSpace(response)
	if len(result) == 0 {
		return "主角重整旗鼓，整合所有资源，准备进行最终对抗"
	}
	return result
}

// buildClimax 构建高潮
func (ne *NarrativeEngine) buildClimax(state *EvolutionState, conflict *ConflictThread) string {
	prompt := fmt.Sprintf(`你是故事结构专家。请为故事设计"高潮"（Climax）。

# 冲突信息
冲突类型：%s
核心问题：%s
当前强度：%d

# 高潮的定义
高潮是故事最紧张、最激烈的时刻，所有线索在此汇聚：
- 主角与反派的最终对决
- 所有伏笔和铺垫在此爆发
- 主角必须面对最终的考验，无法逃避
- 故事的核心主题在此得到最强烈的表达

# 任务要求
请描述这个高潮场景，必须包含：
1. **对抗的场景**（在哪里、如何对抗）
2. **冲突的爆发**（具体发生了什么）
3. **哲学层面的对抗**（不仅是物理对抗，更是价值观、信念的对抗）
4. **80-150字**

# 输出格式
直接输出描述，不要前缀。

# 示例
❌ 错误：主角和反派打了一架
✅ 正确：Thalric与教会长老在水晶大厅对峙。这不仅是武力的对抗，更是两种哲学的决战：长老代表"切除情感=完美"的信念，Thalric则捍卫"保留情感=人性"的立场。当长老以绝对优势压制Thalric时，Thalric没有试图变得"无情"，而是完全释放自己的痛苦——那些曾经被视为"弱点"的情感，此刻成为超越"无情者"的力量源泉。`,
		conflict.Type, conflict.CoreQuestion, conflict.CurrentIntensity)

	response, err := ne.callLLM(prompt)
	if err != nil {
		return "高潮：所有线索汇聚，冲突在激烈的对抗中迎来最终爆发"
	}
	result := strings.TrimSpace(response)
	if len(result) == 0 {
		return "高潮：所有线索汇聚，冲突在激烈的对抗中迎来最终爆发"
	}
	return result
}

// buildResolution 构建结局
func (ne *NarrativeEngine) buildResolution(state *EvolutionState, conflict *ConflictThread) string {
	prompt := fmt.Sprintf(`你是故事结构专家。请为故事设计"结局"（Resolution）。

# 冲突信息
冲突类型：%s
核心问题：%s
当前强度：%d

# 结局的定义
结局是高潮之后的余波，展示故事的最终结果：
- 冲突如何解决（主角的胜利、失败，或者某种融合）
- 主角的成长（主角获得了什么、失去了什么）
- 世界的变化（故事世界如何因主角的旅程而改变）
- 给读者的余味（希望、反思、或复杂的情感）

# 任务要求
请描述这个结局，必须包含：
1. **冲突的解决方式**（具体的解决过程和结果）
2. **主角的成长**（主角获得了什么新理解）
3. **世界的改变**（故事世界如何变化）
4. **80-150字**

# 输出格式
直接输出描述，不要前缀。

# 示例
❌ 错误：主角赢了，大家都很开心
✅ 正确：Thalric没有杀死长老，而是以自己的"有情"击败了长老的"无情"。长老无法理解Thalric为何能在拥有情感的情况下战胜自己，这个认知崩溃导致长老自行晶体化并粉碎。"无情者"教会瓦解，Thalric没有成为新的领袖，而是选择离开——世界不再需要"无情"或"有情"的标签，每个个体都能自由选择。`,
		conflict.Type, conflict.CoreQuestion, conflict.CurrentIntensity)

	response, err := ne.callLLM(prompt)
	if err != nil {
		return "冲突得到解决，主角获得成长，世界迎来新的平衡"
	}
	result := strings.TrimSpace(response)
	if len(result) == 0 {
		return "冲突得到解决，主角获得成长，世界迎来新的平衡"
	}
	return result
}

// buildChapterPlansFromEvolution 从演化状态构建章节规划
func (ne *NarrativeEngine) buildChapterPlansFromEvolution(state *EvolutionState, chapterCount int) []models.ChapterPlan {
	// 使用LLM生成章节规划
	chapterPlans := ne.generateChapterPlansWithLLM(state, chapterCount)

	plans := make([]models.ChapterPlan, chapterCount)
	for i, plan := range chapterPlans {
		plans[i] = models.ChapterPlan{
			Chapter:         i + 1,
			Title:           plan.Title,
			Purpose:         plan.Purpose,
			KeyScenes:       plan.KeyScenes,
			PlotAdvancement: plan.PlotAdvancement,
			ArcProgress:     plan.ArcProgress,
			EndingHook:      plan.EndingHook,
			WordCount:       plan.EstimatedWords,
			Status:          "pending",
		}
	}

	return plans
}

// generateChapterPlansWithLLM 使用LLM生成章节规划
func (ne *NarrativeEngine) generateChapterPlansWithLLM(state *EvolutionState, chapterCount int) []ChapterPlanItem {
	// 构建提示词
	prompt := ne.buildChapterPlanPrompt(state, chapterCount)
	systemPrompt := `你是一位专业的故事策划师，擅长设计引人入胜的章节规划。
每一章都应该有明确的目的、推动情节发展、并展示角色成长。`

	result, err := ne.callWithRetry(prompt, systemPrompt)
	if err != nil {
		// LLM失败时返回简化版本
		return ne.createFallbackChapterPlans(chapterCount)
	}

	// 解析输出
	var output ChapterPlanOutput
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		extracted := extractJSON(result)
		if err := json.Unmarshal([]byte(extracted), &output); err != nil {
			return ne.createFallbackChapterPlans(chapterCount)
		}
	}

	if len(output.Chapters) == 0 {
		return ne.createFallbackChapterPlans(chapterCount)
	}

	// 确保章节数量匹配
	if len(output.Chapters) < chapterCount {
		// 补充缺失的章节
		for i := len(output.Chapters); i < chapterCount; i++ {
			output.Chapters = append(output.Chapters, ChapterPlanItem{
				Chapter:         i + 1,
				Title:           fmt.Sprintf("第%d章", i+1),
				Purpose:         "章节发展",
				KeyScenes:       []string{"关键场景"},
				PlotAdvancement: "情节推进",
				ArcProgress:     "角色发展",
				EndingHook:      "悬念结尾",
				EstimatedWords:  5000,
			})
		}
	}

	return output.Chapters
}

// createFallbackChapterPlans 创建备用章节规划
func (ne *NarrativeEngine) createFallbackChapterPlans(chapterCount int) []ChapterPlanItem {
	plans := make([]ChapterPlanItem, chapterCount)
	for i := 0; i < chapterCount; i++ {
		plans[i] = ChapterPlanItem{
			Chapter:         i + 1,
			Title:           fmt.Sprintf("第%d章", i+1),
			Purpose:         "本章推动情节发展",
			KeyScenes:       []string{"开场场景", "发展场景", "转折场景"},
			PlotAdvancement: "主要情节向前推进",
			ArcProgress:     "角色弧光发展",
			EndingHook:      "留下悬念",
			EstimatedWords:  5000,
		}
	}
	return plans
}

// buildChapterPlanPrompt 构建章节规划提示词
func (ne *NarrativeEngine) buildChapterPlanPrompt(state *EvolutionState, chapterCount int) string {
	var prompt strings.Builder

	prompt.WriteString("# 章节规划任务\n\n")

	prompt.WriteString("## 故事背景\n")
	prompt.WriteString(fmt.Sprintf("- 世界类型: %s\n", state.WorldContext.Type))
	prompt.WriteString(fmt.Sprintf("- 世界规模: %s\n", state.WorldContext.Scale))
	if state.WorldContext.Style != "" {
		prompt.WriteString(fmt.Sprintf("- 风格倾向: %s\n", state.WorldContext.Style))
	}
	prompt.WriteString(fmt.Sprintf("- 核心主题: %s\n", state.ThemeEvolution.CoreTheme))
	prompt.WriteString(fmt.Sprintf("- 章节数量: %d\n", chapterCount))

	// 地理环境（场景地点参考）
	if len(state.WorldContext.Geography.Regions) > 0 {
		prompt.WriteString("\n## 可用地点\n")
		regionNames := make([]string, 0, min(8, len(state.WorldContext.Geography.Regions)))
		for i, region := range state.WorldContext.Geography.Regions {
			if i >= 8 {
				break
			}
			regionNames = append(regionNames, fmt.Sprintf("%s(%s)", region.Name, region.Type))
		}
		prompt.WriteString(strings.Join(regionNames, "、"))
		prompt.WriteString("\n")
	}

	// 核心冲突
	if len(state.Conflicts) > 0 {
		prompt.WriteString("\n## 核心冲突\n")
		for i, c := range state.Conflicts {
			prompt.WriteString(fmt.Sprintf("%d. %s: %s (强度:%d)\n", i+1, c.Type, c.CoreQuestion, c.CurrentIntensity))
			if len(c.EvolutionPath) > 0 {
				prompt.WriteString(fmt.Sprintf("   演化路径: %s", c.EvolutionPath[0].Description))
				for j := 1; j < len(c.EvolutionPath); j++ {
					prompt.WriteString(fmt.Sprintf(" → %s", c.EvolutionPath[j].Description))
				}
				prompt.WriteString("\n")
			}
		}
	}

	// 主要角色
	if len(state.Characters) > 0 {
		prompt.WriteString("\n## 主要角色\n")
		for _, char := range state.Characters {
			prompt.WriteString(fmt.Sprintf("- %s (%s): 欲望=%s, 需求=%s, 恐惧=%s\n",
				char.Name, char.Role, char.Desires.ConsciousWant, char.Desires.UnconsciousNeed, char.Desires.Fear))
		}
	}

	// 伏笔（已在演化中种下）
	if len(state.Foreshadowing) > 0 {
		prompt.WriteString(fmt.Sprintf("\n## 已种下的伏笔 (%d个)\n", len(state.Foreshadowing)))
		for i, f := range state.Foreshadowing {
			if i >= 5 { // 最多显示5个
				prompt.WriteString(fmt.Sprintf("... 还有%d个伏笔\n", len(state.Foreshadowing)-i))
				break
			}
			prompt.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, f.Type, f.Content))
		}
	}

	// 潜在情节钩子
	if len(state.WorldContext.StorySoil.PotentialPlotHooks) > 0 {
		prompt.WriteString("\n## 可利用的情节钩子\n")
		for i, hook := range state.WorldContext.StorySoil.PotentialPlotHooks {
			if i >= 3 { // 最多显示3个
				break
			}
			prompt.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, hook.Type, hook.Description))
		}
	}

	prompt.WriteString("\n# 任务\n")
	prompt.WriteString(fmt.Sprintf("为这%d章的故事设计详细的章节规划，要求：\n", chapterCount))
	prompt.WriteString("1. 每章有吸引人的标题\n")
	prompt.WriteString("2. 每章有明确的目的（为什么要写这一章？）\n")
	prompt.WriteString("3. 列出每章的关键场景（3-5个），考虑使用可用地点\n")
	prompt.WriteString("4. 说明每章如何推进情节\n")
	prompt.WriteString("5. 说明角色弧光如何发展\n")
	prompt.WriteString("6. 每章结尾有吸引读者继续阅读的悬念\n")
	prompt.WriteString("7. 考虑伏笔回收和情节钩子的利用\n")

	prompt.WriteString("\n# 输出格式（JSON）\n")
	prompt.WriteString(`{
  "chapters": [
    {
      "chapter": 1,
      "title": "章节标题",
      "purpose": "本章目的描述",
      "key_scenes": ["场景1描述", "场景2描述", "场景3描述"],
      "plot_advancement": "情节如何推进",
      "arc_progress": "角色弧光如何发展",
      "ending_hook": "结尾悬念",
      "estimated_words": 5000
    }
  ]
}`)

	return prompt.String()
}

// buildScenesFromEvolution 从演化状态构建场景指令
func (ne *NarrativeEngine) buildScenesFromEvolution(state *EvolutionState, plans []models.ChapterPlan) []models.SceneInstruction {
	scenes := make([]models.SceneInstruction, 0)
	globalSequence := 0 // 全局场景序号

	totalScenes := len(plans) * (3 + len(state.Characters)/2)

	fmt.Printf("  🎬 开始生成 %d 个场景...\n", totalScenes)

	sceneIndex := 0
	for chapterIdx := 0; chapterIdx < len(plans); chapterIdx++ {
		plan := plans[chapterIdx]
		sceneCount := 3 + len(state.Characters)/2 // 根据角色数量决定场景数

		for i := 0; i < sceneCount; i++ {
			sceneIndex++
			globalSequence++ // 全局序号递增

			if sceneIndex%5 == 1 || sceneIndex == totalScenes {
				fmt.Printf("    [%d/%d] 生成第%d章场景%d...\n", sceneIndex, totalScenes, plan.Chapter, i+1)
			}

			scene := models.SceneInstruction{
				Chapter:        plan.Chapter,
				Scene:          i + 1,          // 章内场景编号
				Sequence:       globalSequence, // 全局场景序号
				Purpose:        ne.determineScenePurpose(state, plan.Chapter, i),
				Location:       ne.selectLocationForScene(state, plan.Chapter, i),
				Characters:     ne.selectCharactersForScene(state, plan.Chapter, i),
				POVCharacter:   ne.selectPOVCharacter(state),
				Action:         ne.determineSceneAction(state, plan.Chapter, i),
				DialogueFocus:  ne.determineDialogueFocus(state, plan.Chapter, i),
				ExpectedLength: ne.estimateSceneLength(state, plan.Chapter, i),
				Mood:           ne.determineSceneMood(state, plan.Chapter, i),
				Status:         "pending",
			}
			scenes = append(scenes, scene)
		}
	}

	fmt.Printf("  ✓ 场景生成完成\n")
	return scenes
}

// buildCharacterArcsFromEvolution 从角色情感系统构建角色弧光
func (ne *NarrativeEngine) buildCharacterArcsFromEvolution(state *EvolutionState) map[string]*models.ArcPlan {
	arcs := make(map[string]*models.ArcPlan)

	for charID, charState := range state.Characters {
		arc := &models.ArcPlan{
			ArcType: ne.determineArcType(charState),
			StartState: models.CharacterState{
				Personality: []string{charState.EmotionalState.CurrentEmotion},
				Motivation:  charState.Desires.ConsciousWant,
				Emotion:     charState.EmotionalState.CurrentEmotion,
			},
			EndState: models.CharacterState{
				Personality: []string{"成长后的性格"},
				Motivation:  charState.Desires.UnconsciousNeed,
				Emotion:     "平静",
			},
			TurningPoints:  ne.buildTurningPoints(charState, state),
			CurrentProgress: int(charState.ArcProgress * 100),
		}
		arcs[charID] = arc
	}

	return arcs
}

// buildThemePlanFromEvolution 从主题演化构建主题计划
func (ne *NarrativeEngine) buildThemePlanFromEvolution(state *EvolutionState) models.ThemePlan {
	themePlan := models.ThemePlan{
		CoreTheme:    state.ThemeEvolution.CoreTheme,
		Threading:    make([]models.ThemeThreading, 0),
		Symbols:      make([]models.Symbol, 0),
		Motifs:       make([]string, 0),
	}

	// 从主题演化层次构建贯穿
	for _, layer := range state.ThemeEvolution.ThematicLayers {
		themePlan.Threading = append(themePlan.Threading, models.ThemeThreading{
			Chapter:    layer.Chapter,
			Expression: layer.Expression,
			Depth:      layer.Layer,
		})
	}

	// 从象征追踪器构建符号
	for _, symbol := range state.ThemeEvolution.SymbolTracker {
		themePlan.Symbols = append(themePlan.Symbols, models.Symbol{
			Name:        symbol.Name,
			Meaning:     symbol.Meaning,
			Appearances: symbol.Appearances,
		})
	}

	// 如果没有符号，调用LLM生成
	if len(themePlan.Symbols) == 0 {
		themePlan.Symbols = ne.generateSymbols(state)
	}

	// 从母题进展构建母题列表
	for motif := range state.ThemeEvolution.MotifProgress {
		themePlan.Motifs = append(themePlan.Motifs, motif)
	}

	// 如果没有母题，调用LLM生成
	if len(themePlan.Motifs) == 0 {
		themePlan.Motifs = ne.generateMotifs(state)
	}

	return themePlan
}

// generateSymbols 生成故事中的象征符号
func (ne *NarrativeEngine) generateSymbols(state *EvolutionState) []models.Symbol {
	if state.ThemeEvolution.CoreTheme == "" {
		return nil
	}

	characters := ne.getMainCharacters(state, 3)

	prompt := fmt.Sprintf(`你是主题设计专家。请为故事设计3-5个象征符号。

# 核心主题
%s

# 主要角色
%s

# 象征符号的定义
象征符号是故事中反复出现的物体、地点、颜色或自然元素，它们承载着深层的主题意义。每次出现都可能强化或改变其含义。

# 任务要求
请设计3-5个象征符号，每个包含：
1. **名称**（具体的事物）
2. **含义**（象征什么主题、情感或概念）
3. **30-60字**

# 输出格式
请用JSON数组格式输出：
[
  {
    "name": "符号名称",
    "meaning": "象征意义"
  }
]

# 示例
[
  {
    "name": "晶体化",
    "meaning": "象征着"无情"和"完美"——角色们通过晶体化仪式切除情感，但晶体化也让他们失去了人性。当Thalric选择保留痛苦时，他身上的晶体开始碎裂，这象征着他对"完美无情"的放弃"
  },
  {
    "name": "机械幼兽",
    "meaning": "象征着"无辜"和"羁绊"。这只由Thalric抚养的机械兽是他仅存的情感纽带。杀死它意味着彻底切除情感，而保护它则意味着保留人性"
  }
]`,
		state.ThemeEvolution.CoreTheme, characters)

	response, err := ne.callLLM(prompt)
	if err != nil {
		return nil
	}

	// 尝试解析JSON
	var symbols []struct {
		Name    string `json:"name"`
		Meaning string `json:"meaning"`
	}

	if err := json.Unmarshal([]byte(response), &symbols); err != nil {
		// JSON解析失败，返回空
		return nil
	}

	result := make([]models.Symbol, 0, len(symbols))
	for _, s := range symbols {
		result = append(result, models.Symbol{
			Name:        s.Name,
			Meaning:     s.Meaning,
			Appearances: []int{1}, // 默认在第1章出现
		})
	}

	return result
}

// generateMotifs 生成故事中的母题
func (ne *NarrativeEngine) generateMotifs(state *EvolutionState) []string {
	if state.ThemeEvolution.CoreTheme == "" {
		return nil
	}

	characters := ne.getMainCharacters(state, 3)

	prompt := fmt.Sprintf(`你是主题设计专家。请为故事设计3-5个母题（motifs）。

# 核心主题
%s

# 主要角色
%s

# 母题（Motif）的定义
母题是故事中反复出现的模式、元素或思想，比如：
- 反复出现的"牺牲与救赎"场景
- 角色多次面临"无情 vs 仁慈"的抉择
- 反复出现的"镜子"、"水"、"火"等意象
- 特定的对话模式或行为模式

母题与符号的区别：
- 符号是具体的物体或元素
- 母题是反复出现的模式、思想或情境

# 任务要求
请设计3-5个母题，每个包含：
1. **母题的描述**（反复出现的模式是什么）
2. **如何服务主题**（这个母题如何强化或探索核心主题）
3. **30-50字/个**

# 输出格式
请用JSON数组格式输出，每个元素是一个字符串：
[
  "母题1的描述",
  "母题2的描述",
  ...
]

# 示例
[
  "痛苦的镜子：角色们多次在看到他人痛苦时感到自己的创伤被唤起，这些场景象征着"痛苦是连接人类的纽带，而非需要切除的累赘"",
  "选择的十字路口：每个角色都多次面临"切除情感"或"保留痛苦"的抉择，这些选择场景构成了故事的核心冲突",
  "破碎与愈合：角色们的身体和情感都经历了"破碎-愈合"的循环，象征真正的成长不是避免痛苦，而是在痛苦中找到力量"
]`,
		state.ThemeEvolution.CoreTheme, characters)

	response, err := ne.callLLM(prompt)
	if err != nil {
		return nil
	}

	// 尝试解析JSON
	var motifs []string
	if err := json.Unmarshal([]byte(response), &motifs); err != nil {
		// JSON解析失败，返回空
		return nil
	}

	return motifs
}

// ============================================
// 辅助方法（演化状态扩展）
// ============================================

// findMainConflict 找到主要冲突
func (s *EvolutionState) findMainConflict() *ConflictThread {
	if len(s.Conflicts) == 0 {
		return &ConflictThread{}
	}

	mainConflict := s.Conflicts[0]
	maxIntensity := mainConflict.CurrentIntensity

	for _, conflict := range s.Conflicts {
		if conflict.CurrentIntensity > maxIntensity {
			mainConflict = conflict
			maxIntensity = conflict.CurrentIntensity
		}
	}

	return mainConflict
}

// getConflictForChapter 获取指定章节的主要冲突
func (s *EvolutionState) getConflictForChapter(chapter int) *ConflictThread {
	if len(s.Conflicts) == 0 {
		return nil
	}

	// 根据章节轮询分配冲突，确保每个冲突都有发展的空间
	idx := (chapter - 1) % len(s.Conflicts)
	return s.Conflicts[idx]
}

// determineStructureFromConflicts 根据冲突确定叙事结构
func (ne *NarrativeEngine) determineStructureFromConflicts(conflicts []*ConflictThread) NarrativeStructure {
	// 检查是否有"与自己"类型的冲突（内在冲突）
	hasInternalConflict := false
	for _, c := range conflicts {
		if c.Type == "与自己" || c.Type == "internal" {
			hasInternalConflict = true
			break
		}
	}

	if hasInternalConflict {
		return StructureHerosJourney // 内在冲突适合英雄之旅
	}

	return StructureThreeAct
}

// 以下为简化实现
func (ne *NarrativeEngine) determineChapterPurpose(state *EvolutionState, chapterIndex int) string {
	return fmt.Sprintf("第%d章目的", chapterIndex+1)
}

func (ne *NarrativeEngine) extractKeyScenesForChapter(state *EvolutionState, chapterIndex int) []string {
	return []string{"场景1", "场景2", "场景3"}
}

func (ne *NarrativeEngine) determinePlotAdvancement(state *EvolutionState, chapterIndex int) string {
	return "推进情节发展"
}

func (ne *NarrativeEngine) determineArcProgress(state *EvolutionState, chapterIndex int) string {
	return "角色弧光进展"
}

func (ne *NarrativeEngine) generateEndingHook(state *EvolutionState, chapterIndex int) string {
	return "结尾悬念"
}

func (ne *NarrativeEngine) estimateChapterWords(length string) int {
	switch length {
	case "short":
		return 3000
	case "medium":
		return 5000
	case "long":
		return 8000
	default:
		return 4000
	}
}

func (ne *NarrativeEngine) determineScenePurpose(state *EvolutionState, chapter, sceneIndex int) string {
	// 获取章节信息
	chapterTitle := fmt.Sprintf("第%d章", chapter)

	// 获取冲突信息
	conflictInfo := ""
	if len(state.Conflicts) > 0 {
		conflict := state.Conflicts[0]
		conflictInfo = fmt.Sprintf("核心冲突：%s（%s）", conflict.Type, conflict.CoreQuestion)
	}

	// 获取角色信息
	characterInfo := ""
	if len(state.Characters) > 0 {
		charNames := make([]string, 0, min(3, len(state.Characters)))
		for _, char := range state.Characters {
			if len(charNames) >= 3 {
				break
			}
			charNames = append(charNames, char.Name)
		}
		characterInfo = fmt.Sprintf("主要角色：%s", strings.Join(charNames, "、"))
	}

	prompt := fmt.Sprintf(`你是场景设计专家。请为故事中的某个场景设计其"目的"（Purpose）。

# 章节信息
%s

# 冲突信息
%s

# 角色信息
%s

# 场景目的的定义
场景目的回答"为什么需要这个场景"这个问题。每个场景都应该有明确的存在理由：
- 推进情节（展示新的信息、改变局势）
- 展示角色（揭示角色性格、动机、成长）
- 建立氛围（营造情绪、建立基调）
- 铺垫伏笔（为后续情节埋下线索）
- 深化主题（通过具体事件表达故事主题）

# 任务要求
请为第%d章的第%d个场景设计目的，必须包含：
1. **这个场景要实现什么**（具体的目标）
2. **如何服务故事**（如何推进情节、展示角色、或深化主题）
3. **30-60字**

# 输出格式
直接输出描述，不要前缀。

# 示例
❌ 错误：展示冲突
✅ 正确：Thalric被迫执行一项无情的任务，这场景展示他内心"情感"与"职责"的挣扎，同时揭示"无情者"教会的残酷本质，为后续的觉醒埋下伏笔。`,
		chapterTitle, conflictInfo, characterInfo, chapter, sceneIndex)

	response, err := ne.callLLM(prompt)
	if err != nil {
		// 降级方案
		sceneTypes := []string{
			"开场：建立场景氛围",
			"发展：推进情节",
			"冲突：展示矛盾",
			"转折：意外变化",
			"高潮：情感爆发",
			"收尾：留下悬念",
		}
		idx := sceneIndex % len(sceneTypes)
		return sceneTypes[idx]
	}

	result := strings.TrimSpace(response)
	if len(result) == 0 {
		sceneTypes := []string{
			"开场：建立场景氛围",
			"发展：推进情节",
			"冲突：展示矛盾",
			"转折：意外变化",
			"高潮：情感爆发",
			"收尾：留下悬念",
		}
		idx := sceneIndex % len(sceneTypes)
		return sceneTypes[idx]
	}
	return result
}

func (ne *NarrativeEngine) selectLocationForScene(state *EvolutionState, chapter, sceneIndex int) string {
	// 轮换使用不同地理区域
	if len(state.WorldContext.Geography.Regions) > 0 {
		// 根据章节和场景索引选择区域，使地点分布更均匀
		idx := (chapter + sceneIndex) % len(state.WorldContext.Geography.Regions)
		region := state.WorldContext.Geography.Regions[idx]
		// 返回区域名称和类型
		return fmt.Sprintf("%s(%s)", region.Name, region.Type)
	}

	return "默认地点"
}

func (ne *NarrativeEngine) selectCharactersForScene(state *EvolutionState, chapter, sceneIndex int) []string {
	characters := make([]string, 0)
	for charID := range state.Characters {
		characters = append(characters, charID)
	}
	return characters
}

func (ne *NarrativeEngine) selectPOVCharacter(state *EvolutionState) string {
	for charID := range state.Characters {
		return charID
	}
	return ""
}

func (ne *NarrativeEngine) determineSceneAction(state *EvolutionState, chapter, sceneIndex int) string {
	// 获取场景目的作为上下文
	scenePurpose := ne.determineScenePurpose(state, chapter, sceneIndex)

	// 获取冲突信息
	conflictInfo := ""
	if len(state.Conflicts) > 0 {
		conflict := state.Conflicts[0]
		conflictInfo = fmt.Sprintf("核心冲突：%s", conflict.Type)
	}

	// 获取角色信息
	characterInfo := ""
	if len(state.Characters) > 0 {
		charNames := make([]string, 0, min(2, len(state.Characters)))
		for _, char := range state.Characters {
			if len(charNames) >= 2 {
				break
			}
			charNames = append(charNames, char.Name)
		}
		characterInfo = fmt.Sprintf("在场角色：%s", strings.Join(charNames, "、"))
	}

	prompt := fmt.Sprintf(`你是场景设计专家。请为故事中的某个场景设计其"行动"（Action）。

# 场景目的
%s

# 冲突信息
%s

# 角色信息
%s

# 场景行动的定义
场景行动描述"这个场景中发生了什么"。它应该包含：
- 具体的行动或事件（角色做了什么、发生了什么）
- 情感变化（角色的情感如何转变）
- 与情节的连接（这个行动如何推进故事）
- 场景的氛围（紧张、温馨、悬疑等）

# 任务要求
请为第%d章的第%d个场景设计行动，必须包含：
1. **具体的行动或事件**（明确的动作或发生的事情）
2. **情感变化**（角色情感的转变）
3. **与情节的连接**（如何推进故事）
4. **80-150字**

# 输出格式
直接输出描述，不要前缀。

# 示例
❌ 错误：角色们讨论问题
✅ 正确：Thalric站在机械幼兽的笼子前，手中握着教会下达的处决令。他的手指颤抖着，回忆起这只机械兽陪伴他度过无数孤独夜晚的时光。最终，他选择撕碎处决令，将机械幼兽释放。这个决定标志着他第一次公开违抗教会的命令，内心的情感战胜了教条的束缚。`,
		scenePurpose, conflictInfo, characterInfo, chapter, sceneIndex)

	response, err := ne.callLLM(prompt)
	if err != nil {
		return "展示角色互动，推动情节发展"
	}

	result := strings.TrimSpace(response)
	if len(result) == 0 {
		return "展示角色互动，推动情节发展"
	}
	return result
}

func (ne *NarrativeEngine) determineDialogueFocus(state *EvolutionState, chapter, sceneIndex int) string {
	// 确定对话重点
	focuses := []string{
		"探讨冲突的核心问题",
		"揭示角色的内心挣扎",
		"展示不同立场的碰撞",
		"传递关键信息",
		"深化角色关系",
		"暗示未来的发展",
		"回顾过去的经历",
		"表达情感变化",
	}

	// 基于冲突和角色选择合适的对话重点
	conflict := state.getConflictForChapter(chapter)
	if conflict != nil && conflict.Type == "内在冲突" {
		return focuses[1] // 角色内心挣扎
	}

	if conflict != nil && conflict.Type == "人际冲突" {
		return focuses[2] // 不同立场碰撞
	}

	// 根据场景索引循环使用不同的重点
	idx := (chapter + sceneIndex) % len(focuses)
	return focuses[idx]
}

func (ne *NarrativeEngine) estimateSceneLength(state *EvolutionState, chapter, sceneIndex int) int {
	return 800
}

func (ne *NarrativeEngine) determineSceneMood(state *EvolutionState, chapter, sceneIndex int) string {
	// 根据场景位置和冲突强度决定氛围
	moods := []string{
		"平静", "紧张", "悬疑", "温馨", "压抑", "激昂", "诡异", "庄重",
	}
	idx := (chapter + sceneIndex) % len(moods)

	// 如果有活跃冲突，增加紧张感
	for _, conflict := range state.Conflicts {
		if !conflict.IsResolved && conflict.CurrentIntensity > 70 {
			return "紧张"
		}
	}

	return moods[idx]
}

func (ne *NarrativeEngine) determineArcType(char *CharacterState) string {
	if char.Desires.ConsciousWant != char.Desires.UnconsciousNeed {
		return "growth" // 成长弧光
	}
	return "flat"
}

func (ne *NarrativeEngine) buildTurningPoints(char *CharacterState, state *EvolutionState) []models.TurningPoint {
	points := make([]models.TurningPoint, 0)

	// 从冲突线程中提取转折点
	for _, conflict := range state.Conflicts {
		// 检查这个角色是否参与了冲突
		participated := false
		for _, participant := range conflict.Participants {
			if participant == char.Name {
				participated = true
				break
			}
		}

		if !participated {
			continue
		}

		// 获取章节数量，默认12章
		chapterCount := 12

		// 为每个演化阶段生成转折点
		for i, stage := range conflict.EvolutionPath {
			// 将转折点均匀分布到各个章节
			chapter := (i * chapterCount / len(conflict.EvolutionPath)) + 1
			if chapter > chapterCount {
				chapter = chapterCount
			}

			// 调用LLM生成具体的角色变化描述
			changeDesc := ne.generateCharacterChange(
				char.Name,
				char.EmotionalState.CurrentEmotion,
				char.Desires.ConsciousWant,
				stage.Description,
				conflict.Type,
			)

			points = append(points, models.TurningPoint{
				Chapter: chapter,
				Event:   stage.Description,
				Change:  changeDesc,
			})
		}
	}

	return points
}

// generateCharacterChange 生成角色在转折点处的具体变化描述
func (ne *NarrativeEngine) generateCharacterChange(charName, currentEmotion, consciousWant, event, conflictType string) string {
	prompt := fmt.Sprintf(`你是角色弧光设计专家。请描述角色在某个转折点处的具体变化。

# 角色信息
- 姓名：%s
- 当前情感：%s
- 欲望：%s

# 转折点事件
%s

# 冲突类型
%s

# 任务要求
请描述这个事件如何改变了角色，必须包含：
1. **认知变化**（角色对某事有了新的理解）
2. **情感变化**（角色的情感状态发生了什么转变）
3. **行为倾向变化**（角色之后的行为会有什么不同）
4. **50-80字**

# 输出格式
直接输出描述，不要前缀。

# 示例
❌ 错误：角色状态变化
✅ 正确：Thalric的犹豫暴露了他内心的矛盾——他并非真的"无情"，而是在用冷酷掩饰脆弱。这次失败让他开始怀疑"切除情感"是否是正确的道路，他的内心冲突从"如何变得无情"转向"是否应该变得无情"。`,
		charName, currentEmotion, consciousWant, event, conflictType)

	response, err := ne.callLLM(prompt)
	if err != nil {
		return "角色状态发生变化"
	}

	result := strings.TrimSpace(response)
	if len(result) == 0 {
		return "角色状态发生变化"
	}
	return result
}

// CreateBlueprint 创建叙事蓝图（集成动态演化）
func (ne *NarrativeEngine) CreateBlueprint(params CreateParams) (*models.NarrativeBlueprint, error) {
	// 首先执行动态演化，生成丰富的叙事内容
	evolutionConfig := EvolutionConfig{
		EnableEvolution: true,
		MaxRounds:       8, // 执行8轮演化
		AutoStopWhen:    85, // 质量达到85分时停止
	}

	blueprint, _, err := ne.CreateBlueprintThroughEvolution(params, evolutionConfig)
	if err != nil {
		return nil, fmt.Errorf("动态演化失败: %w", err)
	}

	// 保存到数据库
	if err := ne.db.SaveNarrativeBlueprint(blueprint); err != nil {
		return nil, fmt.Errorf("保存叙事蓝图失败: %w", err)
	}

	return blueprint, nil
}

// defaultStructure 根据故事类型返回默认叙事结构
func (ne *NarrativeEngine) defaultStructure(storyType string) NarrativeStructure {
	switch storyType {
	case "成长", "adventure":
		return StructureHerosJourney
	case "商业片", "动作", "comedy":
		return StructureSaveTheCat
	case "东方", "武侠", "仙侠":
		return StructureKishotenketsu
	case "悲剧", "正剧":
		return StructureFreytagPyramid
	default:
		return StructureThreeAct
	}
}

// AssignCharacterArc 为蓝图分配角色弧光
func (ne *NarrativeEngine) AssignCharacterArc(blueprintID, characterID string, arcPlan *models.ArcPlan) error {
	blueprint, err := ne.db.GetNarrativeBlueprint(blueprintID)
	if err != nil {
		return fmt.Errorf("获取叙事蓝图失败: %w", err)
	}

	if blueprint.CharacterArcs == nil {
		blueprint.CharacterArcs = make(map[string]*models.ArcPlan)
	}

	blueprint.CharacterArcs[characterID] = arcPlan
	blueprint.UpdatedAt = time.Now()

	return ne.db.SaveNarrativeBlueprint(blueprint)
}

// buildWorldSummary 构建世界设定摘要
func (ne *NarrativeEngine) buildWorldSummary(world *models.WorldSetting) string {
	summary := fmt.Sprintf("世界名称: %s\n类型: %s\n规模: %s\n\n",
		world.Name, world.Type, world.Scale)

	summary += fmt.Sprintf("【哲学】核心问题: %s\n", world.Philosophy.CoreQuestion)
	summary += fmt.Sprintf("【价值观】最高善: %s, 最大恶: %s\n",
		world.Philosophy.ValueSystem.HighestGood,
		world.Philosophy.ValueSystem.UltimateEvil)

	// 故事土壤
	if len(world.StorySoil.SocialConflicts) > 0 {
		summary += fmt.Sprintf("【社会冲突】%d个主要矛盾\n", len(world.StorySoil.SocialConflicts))
		for i, conflict := range world.StorySoil.SocialConflicts {
			if i < 2 { // 只列出前两个
				summary += fmt.Sprintf("  - %s: %s\n", conflict.Type, conflict.Description)
			}
		}
	}

	if len(world.StorySoil.PotentialPlotHooks) > 0 {
		summary += fmt.Sprintf("【情节钩子】%d个潜在故事点\n", len(world.StorySoil.PotentialPlotHooks))
	}

	// 地理环境
	if len(world.Geography.Regions) > 0 {
		summary += fmt.Sprintf("【地理】%d个区域，气候类型: %s\n",
			len(world.Geography.Regions),
			func() string {
				if world.Geography.Climate != nil {
					return world.Geography.Climate.Type
				}
				return "未知"
			}())
	}

	// 文明
	if len(world.Civilization.Races) > 0 {
		summary += "【种族】"
		for i, race := range world.Civilization.Races {
			if i > 0 {
				summary += ", "
			}
			summary += race.Name
		}
		summary += "\n"
	}

	return summary
}

// defaultChapterCount 根据篇幅返回默认章节数
func (ne *NarrativeEngine) defaultChapterCount(length string) int {
	switch length {
	case "short":
		return 10
	case "medium":
		return 30
	case "long":
		return 60
	default:
		return 20
	}
}

// planTheme 规划主题贯穿
func (ne *NarrativeEngine) planTheme(coreTheme string, chapterCount int) models.ThemePlan {
	plan := models.ThemePlan{
		CoreTheme: coreTheme,
		Threading:  make([]models.ThemeThreading, 0),
		Symbols:    make([]models.Symbol, 0),
		Motifs:     []string{},
	}

	// 为每章规划主题深度
	for i := 1; i <= chapterCount; i += chapterCount / 5 {
		depth := "surface"
		if i > chapterCount/2 {
			depth = "deep"
		} else if i > chapterCount/4 {
			depth = "philosophical"
		}

		plan.Threading = append(plan.Threading, models.ThemeThreading{
			Chapter:    i,
			Expression: fmt.Sprintf("第%d章主题探索", i),
			Depth:      depth,
		})
	}

	return plan
}

// updatePreviousSummary 更新前情摘要
func (ne *NarrativeEngine) updatePreviousSummary(chapter models.ChapterPlan, scenes *SceneOutput) string {
	summary := fmt.Sprintf("第%d章《%s》：%s。关键场景：",
		chapter.Chapter, chapter.Title, chapter.Purpose)
	for i, scene := range scenes.Scenes {
		if i > 0 {
			summary += " → "
		}
		summary += fmt.Sprintf("%s(%s)", scene.Location, scene.Purpose)
	}
	summary += fmt.Sprintf("。本章推进了：%s", chapter.PlotAdvancement)
	return summary
}

// callWithRetry 调用LLM并自动重试
func (ne *NarrativeEngine) callWithRetry(prompt, systemPrompt string) (string, error) {
	retryConfig := ne.cfg.System.Retry
	maxAttempts := retryConfig.MaxAttempts
	var lastErr error

	fmt.Println("\n========== LLM DEBUG (JSON) ==========")
	fmt.Printf("System Prompt:\n%s\n\n", systemPrompt)
	fmt.Printf("User Prompt:\n%s\n", truncateForDebug(prompt, 2000))
	fmt.Println("====================================")

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("🔄 尝试 %d/%d...\n", attempt, maxAttempts)
		startTime := time.Now()

		result, err := ne.client.GenerateJSONWithParams(
			prompt,
			systemPrompt,
			ne.mapping.Temperature,
			ne.mapping.MaxTokens,
		)

		elapsed := time.Since(startTime)
		fmt.Printf("⏱️  耗时: %.1f秒\n", elapsed.Seconds())

		if err == nil {
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				fmt.Printf("❌ 序列化结果失败: %v\n", err)
				return "", fmt.Errorf("序列化结果失败: %w", err)
			}
			fmt.Printf("✅ 响应成功\n")
			fmt.Printf("Response:\n%s\n", truncateForDebug(string(jsonBytes), 3000))
			fmt.Println("====================================\n")
			return string(jsonBytes), nil
		}

		fmt.Printf("❌ 调用失败: %v\n", err)
		lastErr = err

		if attempt < maxAttempts {
			delay := time.Duration(retryConfig.InitialDelay*attempt) * time.Second
			if delay > time.Duration(retryConfig.MaxDelay)*time.Second {
				delay = time.Duration(retryConfig.MaxDelay) * time.Second
			}
			fmt.Printf("⏳ 等待 %.1f 秒后重试...\n", delay.Seconds())
			time.Sleep(delay)
		}
	}

	fmt.Printf("❌ LLM调用失败（重试%d次后）: %v\n", maxAttempts, lastErr)
	fmt.Println("====================================\n")
	return "", fmt.Errorf("LLM调用失败（重试%d次后）: %w", maxAttempts, lastErr)
}

// extractJSON 从文本中提取JSON内容
func extractJSON(s string) string {
	// 查找 ```json```
	start := -1
	end := -1

	jsonStart := []byte("```json")
	if idx := indexOf(s, jsonStart); idx >= 0 {
		start = idx + len(jsonStart)
		if idx := indexOf(s[start:], []byte("```")); idx >= 0 {
			end = start + idx
			return s[start:end]
		}
	}

	// 查找 ````
	if idx := indexOf(s, []byte("```")); idx >= 0 {
		start = idx + 3
		if idx := indexOf(s[start:], []byte("```")); idx >= 0 {
			end = start + idx
			return s[start:end]
		}
	}

	// 查找 { }
	if idx := indexOf(s, []byte("{")); idx >= 0 {
		start = idx
		if idx := lastIndexOf(s, []byte("}")); idx >= 0 {
			end = idx + 1
			return s[start:end]
		}
	}

	return s
}

func indexOf(s string, sep []byte) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		match := true
		for j := 0; j < len(sep); j++ {
			if s[i+j] != sep[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func lastIndexOf(s string, sep []byte) int {
	for i := len(s) - len(sep); i >= 0; i-- {
		match := true
		for j := 0; j < len(sep); j++ {
			if s[i+j] != sep[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getMainCharacters 获取主要角色列表的字符串描述
func (ne *NarrativeEngine) getMainCharacters(state *EvolutionState, maxCount int) string {
	if len(state.Characters) == 0 {
		return "暂无角色"
	}

	var names []string
	count := min(maxCount, len(state.Characters))
	for _, char := range state.Characters {
		if len(names) >= count {
			break
		}
		names = append(names, char.Name)
	}

	result := strings.Join(names, "、")
	if len(state.Characters) > maxCount {
		result += fmt.Sprintf(" 等%d人", len(state.Characters))
	}
	return result
}

// callLLM 调用LLM的辅助函数
func (ne *NarrativeEngine) callLLM(prompt string) (string, error) {
	fmt.Println("\n========== LLM DEBUG (TEXT) ==========")
	fmt.Printf("User Prompt:\n%s\n", truncateForDebug(prompt, 2000))
	fmt.Println("====================================")

	fmt.Println("🔄 调用LLM...")
	startTime := time.Now()

	response, err := ne.client.GenerateWithParams(
		prompt,
		"", // 系统提示，可以留空
		ne.mapping.Temperature,
		ne.mapping.MaxTokens,
	)

	elapsed := time.Since(startTime)
	fmt.Printf("⏱️  耗时: %.1f秒\n", elapsed.Seconds())

	if err != nil {
		fmt.Printf("❌ 调用失败: %v\n", err)
		fmt.Println("====================================\n")
		return "", err
	}

	fmt.Printf("✅ 响应成功\n")
	fmt.Printf("Response:\n%s\n", truncateForDebug(response, 3000))
	fmt.Println("====================================\n")

	return response, nil
}

// truncateForDebug 截断过长的调试输出
func truncateForDebug(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (截断，总长度: " + fmt.Sprintf("%d", len(s)) + " 字符)"
}
