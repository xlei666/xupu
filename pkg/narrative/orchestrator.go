// Package narrative 叙事器 - 系统的大脑
// 真正的链式演化编排器
package narrative

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xlei/xupu/internal/models"
)

// Orchestrator 演化编排器
type Orchestrator struct {
	engine *EvolutionEngine
}

// NewOrchestrator 创建编排器
func NewOrchestrator(engine *EvolutionEngine) *Orchestrator {
	return &Orchestrator{
		engine: engine,
	}
}

// ExecuteFullEvolution 执行完整的演化流程（约200轮LLM）
func (o *Orchestrator) ExecuteFullEvolution(worldID string, chapterCount int) (*EvolutionState, error) {
	fmt.Println("🔄 [初始化] 正在初始化演化状态...")
	// 初始化演化状态
	state, err := o.engine.CreateEvolutionState(worldID)
	if err != nil {
		return nil, fmt.Errorf("初始化演化状态失败: %w", err)
	}
	fmt.Printf("✓ 初始化完成 (轮次: %d)\n\n", state.CurrentRound)

	// 阶段1：故事架构设计（10-15轮）
	fmt.Println("🏗️  [阶段1/7] 故事架构设计 (10-15轮LLM)...")
	fmt.Println("  ├─ 分析世界设定，确定叙事模式")
	fmt.Println("  ├─ 规划角色阵容架构")
	fmt.Println("  └─ 确定核心矛盾线索")
	if err := o.phase1_StoryArchitecture(state); err != nil {
		return nil, fmt.Errorf("故事架构设计失败: %w", err)
	}
	fmt.Printf("✓ 阶段1完成 (当前轮次: %d)\n\n", state.CurrentRound)

	// 阶段2：角色创建与关系网络（40-50轮）
	fmt.Println("👥 [阶段2/7] 角色创建与关系网络 (40-50轮LLM)...")
	fmt.Printf("  ├─ 创建 %d 个角色 (每角色3轮)\n", state.StoryArchitecture.CharacterRoster.TotalCharacters)
	fmt.Println("  ├─ 构建关系网络 (5-8轮)")
	fmt.Println("  ├─ 演化关系网络 (5-10轮)")
	fmt.Println("  └─ 自动识别主角")
	if err := o.phase2_CharactersAndRelationships(state); err != nil {
		return nil, fmt.Errorf("角色创建失败: %w", err)
	}
	fmt.Printf("✓ 阶段2完成 (当前轮次: %d)\n", state.CurrentRound)
	if state.RelationshipNetwork.CenterNode != "" {
		protagonist := state.Characters[state.RelationshipNetwork.CenterNode]
		fmt.Printf("  ✓ 主角识别: %s\n\n", protagonist.Name)
	} else {
		fmt.Println()
	}

	// 阶段3：伏笔系统设计（10-15轮）
	fmt.Println("🔮 [阶段3/7] 伏笔系统设计 (10-15轮LLM)...")
	fmt.Println("  ├─ 规划伏笔网络 (5-8轮)")
	fmt.Println("  └─ 验证伏笔完整性 (5-7轮)")
	if err := o.phase3_ForeshadowPlanning(state); err != nil {
		return nil, fmt.Errorf("伏笔系统设计失败: %w", err)
	}
	fmt.Printf("✓ 阶段3完成 - 规划了 %d 个伏笔 (当前轮次: %d)\n\n", len(state.ForeshadowPlan), state.CurrentRound)

	// 阶段4：冲突系统设计（20-30轮）
	fmt.Println("⚔️  [阶段4/7] 冲突系统设计 (20-30轮LLM)...")
	fmt.Printf("  ├─ 设计 %d 个核心冲突 (每冲突2轮)\n", len(state.Characters)+2)
	fmt.Println("  └─ 构建冲突层级 (3-5轮)")
	if err := o.phase4_ConflictSystem(state); err != nil {
		return nil, fmt.Errorf("冲突系统设计失败: %w", err)
	}
	fmt.Printf("✓ 阶段4完成 - 设计了 %d 个冲突 (当前轮次: %d)\n\n", len(state.Conflicts), state.CurrentRound)

	// 阶段5：生成主要故事大纲（15-20轮）
	fmt.Println("📖 [阶段5/7] 生成主要故事大纲 (15-20轮LLM)...")
	fmt.Println("  ├─ 规划故事走向 (1轮)")
	fmt.Println("  ├─ 设计关键事件序列 (1轮)")
	fmt.Println("  ├─ 设计高潮和结局 (1轮)")
	fmt.Println("  └─ 构建伏笔链接")
	if err := o.phase5_GlobalOutline(state); err != nil {
		return nil, fmt.Errorf("故事大纲生成失败: %w", err)
	}
	fmt.Printf("✓ 阶段5完成 - 设计了 %d 个关键事件 (当前轮次: %d)\n\n", len(state.GlobalOutline.KeyEvents), state.CurrentRound)

	// 阶段6：章节规划（10-15轮）
	fmt.Printf("📚 [阶段6/7] 章节规划 (10-15轮LLM)...\n")
	fmt.Printf("  ├─ 将关键事件分配到 %d 个章节 (5-8轮)\n", chapterCount)
	fmt.Println("  └─ 优化章节序列和连接 (5-7轮)")
	if err := o.phase6_ChapterPlanning(state, chapterCount); err != nil {
		return nil, fmt.Errorf("章节规划失败: %w", err)
	}
	fmt.Printf("✓ 阶段6完成 - 规划了 %d 个章节 (当前轮次: %d)\n\n", len(state.ChapterPlan.ChapterSequence), state.CurrentRound)

	// 阶段7：细纲生成（每章10-15轮，在生成时按需执行）
	fmt.Println("🎯 [阶段7/7] 细纲生成系统 (按需执行)")
	fmt.Println("  阶段7不是一次性执行，而是在生成每章细纲时按需调用")
	fmt.Println("  每章细纲生成包括：")
	fmt.Println("    • 设计场景序列 (2-3轮)")
	fmt.Println("    • 生成场景详细指令 (每场景1轮)")
	fmt.Println("    • 追踪角色演化 (1轮)")
	fmt.Println("    • 规划伏笔操作")
	fmt.Println("    • 估算字数和写作指导")
	fmt.Printf("  预计每章需要: 10-15轮LLM\n\n")

	// 阶段7：细纲生成（每章10-15轮，在生成时按需执行）
	// 这个阶段不是一次性执行，而是按需生成
	// 这里只设置标志
	// state.CurrentRound = 0 // 重置轮次计数器，为细纲生成准备

	return state, nil
}

// phase1_StoryArchitecture 阶段1：故事架构设计
func (o *Orchestrator) phase1_StoryArchitecture(state *EvolutionState) error {
	// 1.1 分析世界设定，确定叙事模式（3-4轮）
	narrativeMode, err := o.analyzeWorldAndDetermineMode(state)
	if err != nil {
		return err
	}

	// 1.2 规划角色阵容架构（3-4轮）
	characterRoster, err := o.planCharacterRoster(state, narrativeMode)
	if err != nil {
		return err
	}

	// 1.3 确定核心矛盾线索（4-6轮）
	coreConflicts, err := o.identifyCoreConflictDirections(state, narrativeMode, characterRoster)
	if err != nil {
		return err
	}

	// 保存架构信息
	state.StoryArchitecture = &StoryArchitecture{
		NarrativeMode:     narrativeMode,
		CoreConflictType:  coreConflicts,
		CharacterRoster:  characterRoster,
		MainDirection:    "",
		ExpectedEnding:    "",
	}

	state.logAction(state.CurrentRound, "story_architecture", "故事架构设计完成", []string{
		fmt.Sprintf("叙事模式: %s", narrativeMode),
		fmt.Sprintf("角色数量: %d", characterRoster.TotalCharacters),
	})

	return nil
}

// phase2_CharactersAndRelationships 阶段2：角色创建与关系网络（40-50轮）
func (o *Orchestrator) phase2_CharactersAndRelationships(state *EvolutionState) error {
	roster := state.StoryArchitecture.CharacterRoster

	// 2.1 逐个创建角色（每个角色3-4轮）
	for i := 0; i < roster.TotalCharacters; i++ {
		character, err := o.createCharacterWithDepth(state, i)
		if err != nil {
			return err
		}
		state.Characters[character.ID] = character
	}

	// 2.2 构建关系网络（5-8轮）
	network, err := o.buildRelationshipNetwork(state)
	if err != nil {
		return err
	}
	state.RelationshipNetwork = network

	// 2.3 演化关系网络（5-10轮）
	if err := o.evolveRelationshipNetwork(state); err != nil {
		return err
	}

	// 识别主角
	protagonist := o.identifyProtagonist(state)
	state.RelationshipNetwork.CenterNode = protagonist

	// 初始化角色演化追踪
	state.CharacterEvolution = make(map[string]*CharacterEvolutionTracker)
	for charID := range state.Characters {
		state.CharacterEvolution[charID] = &CharacterEvolutionTracker{
			CharacterID:         charID,
			EmotionalJourney:    []EmotionalState{},
			RelationshipHistory:  make(map[string][]RelationshipHistoryEntry),
			KnowledgeGrowth:      []KnowledgePiece{},
			TurningPoints:        []TurningPoint{},
			ChapterChanges:       make(map[string]*ChapterCharacterChange),
		}
	}

	return nil
}

// phase3_ForeshadowPlanning 阶段3：伏笔系统设计（10-15轮）
func (o *Orchestrator) phase3_ForeshadowPlanning(state *EvolutionState) error {
	// 3.1 规划伏笔网（5-8轮）
	foreshadowPlan, err := o.planForeshadowNetwork(state)
	if err != nil {
		return err
	}

	// 3.2 验证伏笔的完整性（5-7轮）
	if err := o.validateForeshadowPlan(state, foreshadowPlan); err != nil {
		return err
	}

	state.ForeshadowPlan = foreshadowPlan

	return nil
}

// phase4_ConflictSystem 阶段4：冲突系统设计（20-30轮）
func (o *Orchestrator) phase4_ConflictSystem(state *EvolutionState) error {
	// 4.1 设计核心冲突（每个冲突4-5轮）
	conflicts, err := o.designCoreConflicts(state)
	if err != nil {
		return err
	}
	state.Conflicts = conflicts

	// 4.2 构建冲突层级（3-5轮）
	if err := o.buildConflictHierarchy(state); err != nil {
		return err
	}

	return nil
}

// phase5_GlobalOutline 阶段5：生成主要故事大纲（15-20轮）
func (o *Orchestrator) phase5_GlobalOutline(state *EvolutionState) error {
	// 5.1 规划故事走向（结合伏笔）（6-8轮）
	opening, direction, err := o.planStoryDirection(state)
	if err != nil {
		return err
	}

	// 5.2 设计关键事件序列（8-10轮）
	keyEvents, err := o.designKeyEvents(state, opening, direction)
	if err != nil {
		return err
	}

	// 5.3 验证大纲的连贯性
	climax, resolution, err := o.designClimaxAndResolution(state, keyEvents)
	if err != nil {
		return err
	}

	// 构建伏笔链接
	foreshadowLinks := o.buildForeshadowLinks(state, keyEvents)

	state.GlobalOutline = &GlobalOutline{
		Opening:          opening,
		KeyEvents:        keyEvents,
		Climax:           climax,
		Resolution:       resolution,
		ForeshadowLinks: foreshadowLinks,
	}

	return nil
}

// phase6_ChapterPlanning 阶段6：章节规划（10-15轮）
func (o *Orchestrator) phase6_ChapterPlanning(state *EvolutionState, chapterCount int) error {
	// 6.1 将关键事件分配到章节（5-8轮）
	chapterSequence, err := o.assignEventsToChapters(state, chapterCount)
	if err != nil {
		return err
	}

	// 6.2 确定章节序列和连接（5-7轮）
	if err := o.refineChapterSequence(state, chapterSequence); err != nil {
		return err
	}

	state.ChapterPlan = &ChapterPlan{
		TotalChapters:   chapterCount,
		ChapterSequence: chapterSequence,
	}

	return nil
}

// ============ 阶段1的具体实现 ============

// analyzeWorldAndDetermineMode 分析世界设定，确定叙事模式（3-4轮LLM）
func (o *Orchestrator) analyzeWorldAndDetermineMode(state *EvolutionState) (string, error) {
	state.CurrentRound++

	// 第1轮：分析世界设定的核心特征
	worldAnalysisPrompt := o.buildWorldAnalysisPrompt(state)
	systemPrompt := o.buildSystemPrompt("story_architecture_analyzer")

	response, err := o.engine.callWithRetry(worldAnalysisPrompt, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("世界分析失败: %w", err)
	}

	var result struct {
		CoreTensions    []string `json:"core_tensions"`
		StoryPotential  []string `json:"story_potential"`
		Scale            string   `json:"scale"`
		Complexity       string   `json:"complexity"`
		SuggestedModes   []string `json:"suggested_modes"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "", fmt.Errorf("解析世界分析结果失败: %w", err)
	}

	state.logAction(state.CurrentRound, "world_analysis", "世界设定分析", []string{
		fmt.Sprintf("核心张力: %v", result.CoreTensions),
		fmt.Sprintf("建议模式: %v", result.SuggestedModes),
	})

	// 第2轮：确定最适合的叙事模式
	modeDeterminationPrompt := o.buildModeDeterminationPrompt(state, &result)
	modeResponse, err := o.engine.callWithRetry(modeDeterminationPrompt, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("叙事模式确定失败: %w", err)
	}

	var modeResult struct {
		SelectedMode   string   `json:"selected_mode"`
		Reasoning      string   `json:"reasoning"`
		Considerations []string `json:"considerations"`
	}
	if err := json.Unmarshal([]byte(modeResponse), &modeResult); err != nil {
		return "", fmt.Errorf("解析模式确定结果失败: %w", err)
	}

	state.logAction(state.CurrentRound, "mode_determination", "叙事模式确定", []string{
		fmt.Sprintf("选定模式: %s", modeResult.SelectedMode),
		fmt.Sprintf("理由: %s", modeResult.Reasoning),
	})

	return modeResult.SelectedMode, nil
}

// planCharacterRoster 规划角色阵容架构（3-4轮LLM）
func (o *Orchestrator) planCharacterRoster(state *EvolutionState, mode string) (CharacterRosterSpec, error) {
	state.CurrentRound++

	// 第1轮：基于叙事模式确定角色阵容
	rosterPrompt := o.buildRosterPlanningPrompt(state, mode)
	systemPrompt := o.buildSystemPrompt("character_roster_planner")

	response, err := o.engine.callWithRetry(rosterPrompt, systemPrompt)
	if err != nil {
		return CharacterRosterSpec{}, fmt.Errorf("角色阵容规划失败: %w", err)
	}

	var result struct {
		TotalCharacters  int      `json:"total_characters"`
		ProtagonistCount int      `json:"protagonist_count"`
		AntagonistCount  int      `json:"antagonist_count"`
		SupportingCount  int      `json:"supporting_count"`
		NetworkStructure string   `json:"network_structure"`
		KeyRelationships  []string `json:"key_relationships"`
		CharacterTypes   []string `json:"character_types"`
		Reasoning        string   `json:"reasoning"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return CharacterRosterSpec{}, fmt.Errorf("解析角色阵容结果失败: %w", err)
	}

	state.logAction(state.CurrentRound, "roster_planning", "角色阵容规划", []string{
		fmt.Sprintf("总角色数: %d", result.TotalCharacters),
		fmt.Sprintf("网络结构: %s", result.NetworkStructure),
		fmt.Sprintf("角色类型: %v", result.CharacterTypes),
	})

	return CharacterRosterSpec{
		TotalCharacters:  result.TotalCharacters,
		ProtagonistCount: result.ProtagonistCount,
		AntagonistCount:  result.AntagonistCount,
		SupportingCount:  result.SupportingCount,
		NetworkStructure: result.NetworkStructure,
		KeyRelationships:  result.KeyRelationships,
	}, nil
}

// identifyCoreConflictDirections 确定核心矛盾线索（4-6轮LLM）
func (o *Orchestrator) identifyCoreConflictDirections(state *EvolutionState, mode string, roster CharacterRosterSpec) (string, error) {
	state.CurrentRound++

	// 第1轮：识别潜在的核心冲突方向
	conflictPrompt := o.buildConflictIdentificationPrompt(state, mode, roster)
	systemPrompt := o.buildSystemPrompt("conflict_architect")

	response, err := o.engine.callWithRetry(conflictPrompt, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("冲突识别失败: %w", err)
	}

	var result struct {
		PrimaryConflicts   []string `json:"primary_conflicts"`
		SecondaryConflicts []string `json:"secondary_conflicts"`
		ThematicCore       string   `json:"thematic_core"`
		ConflictDirection  string   `json:"conflict_direction"`
		Reasoning          string   `json:"reasoning"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "", fmt.Errorf("解析冲突识别结果失败: %w", err)
	}

	state.logAction(state.CurrentRound, "conflict_identification", "核心冲突识别", []string{
		fmt.Sprintf("主要冲突: %v", result.PrimaryConflicts),
		fmt.Sprintf("冲突方向: %s", result.ConflictDirection),
		fmt.Sprintf("主题核心: %s", result.ThematicCore),
	})

	// 第2轮：深化冲突方向
	deepenPrompt := o.buildConflictDeepeningPrompt(state, &result)
	deepenResponse, err := o.engine.callWithRetry(deepenPrompt, systemPrompt)
	if err != nil {
		return result.ConflictDirection, nil // 返回初步结果
	}

	var deepenResult struct {
		RefinedDirection string   `json:"refined_direction"`
		ConflictLayers   []string `json:"conflict_layers"`
		EvolutionPath    []string `json:"evolution_path"`
	}
	if err := json.Unmarshal([]byte(deepenResponse), &deepenResult); err == nil {
		state.logAction(state.CurrentRound, "conflict_deepening", "冲突方向深化", []string{
			fmt.Sprintf("精炼方向: %s", deepenResult.RefinedDirection),
			fmt.Sprintf("冲突层级: %v", deepenResult.ConflictLayers),
		})
		return deepenResult.RefinedDirection, nil
	}

	return result.ConflictDirection, nil
}

// ============ 阶段2的具体实现 ============

// createCharacterWithDepth 创建角色，包含多轮LLM调用（每个角色3-4轮）
func (o *Orchestrator) createCharacterWithDepth(state *EvolutionState, index int) (*CharacterState, error) {
	state.CurrentRound++

	charID := fmt.Sprintf("char_%d", index)

	// 第1轮：创建角色基本信息
	character, err := o.createCharacterBasicInfo(state, charID, index)
	if err != nil {
		return nil, err
	}

	// 第2轮：深化角色内在冲突
	state.CurrentRound++
	if err := o.deepenCharacterInternalConflict(state, character); err != nil {
		return nil, err
	}

	// 第3轮：确定角色在关系网络中的定位
	state.CurrentRound++
	if err := o.positionCharacterInNetwork(state, character); err != nil {
		return nil, err
	}

	return character, nil
}

// createCharacterBasicInfo 创建角色基本信息（第1轮）
func (o *Orchestrator) createCharacterBasicInfo(state *EvolutionState, charID string, index int) (*CharacterState, error) {
	prompt := o.buildCharacterCreationPrompt(state, index)
	systemPrompt := o.buildSystemPrompt("character_creator")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("角色创建失败: %w", err)
	}

	var result struct {
		Name            string   `json:"name"`
		Role            string   `json:"role"`
		Age             int      `json:"age"`
		Background      string   `json:"background"`
		Personality     []string `json:"personality"`
		ConsciousWant   string   `json:"conscious_want"`
		UnconsciousNeed string   `json:"unconscious_need"`
		CoreTraits      []string `json:"core_traits"`
		Flaws           []string `json:"flaws"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析角色创建结果失败: %w", err)
	}

	character := &CharacterState{
		ID:   charID,
		Name: result.Name,
		Role: result.Role,
		EmotionalState: EmotionalSystem{
			CurrentEmotion:     "平静",
			EmotionalIntensity: 50,
			EmotionalStack:     []string{},
		},
		Desires: DesireSystem{
			ConsciousWant:   result.ConsciousWant,
			UnconsciousNeed: result.UnconsciousNeed,
		},
		Relationships:     make(map[string]*RelationshipState),
		ArcProgress:       0.0,
		InternalConflicts: []string{},
		Secrets:          []string{},
	}

	state.logAction(state.CurrentRound, "character_creation", "创建角色", []string{
		fmt.Sprintf("角色名: %s", result.Name),
		fmt.Sprintf("角色: %s", result.Role),
		fmt.Sprintf("意识欲望: %s", result.ConsciousWant),
	})

	return character, nil
}

// deepenCharacterInternalConflict 深化角色内在冲突（第2轮）
func (o *Orchestrator) deepenCharacterInternalConflict(state *EvolutionState, character *CharacterState) error {
	prompt := o.buildCharacterDeepeningPrompt(state, character)
	systemPrompt := o.buildSystemPrompt("character_psychologist")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return fmt.Errorf("角色深化失败: %w", err)
	}

	// 调试：打印原始响应
	fmt.Printf("  [DEBUG] 原始响应长度: %d\n", len(response))

	var result struct {
		InternalConflicts []string `json:"internal_conflicts"`
		Secrets           []string `json:"secrets"`
		Fears             []string `json:"fears"`
		Triggers          []string `json:"triggers"`
		MaskingBehaviors  []string `json:"masking_behaviors"`
		WantVsNeedGap     string   `json:"want_vs_need_gap"`
		Desires           []string `json:"desires"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// 打印前500个字符用于调试
		preview := response
		if len(preview) > 500 {
			preview = preview[:500]
		}
		return fmt.Errorf("解析角色深化结果失败: %w\n原始响应前500字符: %s", err, preview)
	}

	character.InternalConflicts = result.InternalConflicts
	character.Secrets = result.Secrets
	character.Desires.Fear = strings.Join(result.Fears, "; ")
	if len(result.Triggers) > 0 {
		character.EmotionalState.Triggers = result.Triggers
	}
	if len(result.MaskingBehaviors) > 0 {
		character.Desires.MaskingBehavior = result.MaskingBehaviors
	}
	character.Desires.WantVsNeedGap = result.WantVsNeedGap

	state.logAction(state.CurrentRound, "character_deepening", "角色深化", []string{
		fmt.Sprintf("角色: %s", character.Name),
		fmt.Sprintf("内在冲突: %v", result.InternalConflicts),
		fmt.Sprintf("恐惧: %v", result.Fears),
	})

	return nil
}

// positionCharacterInNetwork 确定角色在关系网络中的定位（第3轮）
func (o *Orchestrator) positionCharacterInNetwork(state *EvolutionState, character *CharacterState) error {
	// 这里需要知道其他已存在的角色，但由于是逐个创建，
	// 这个方法在建立关系网络时会更有意义
	// 暂时只记录角色的初始定位

	state.logAction(state.CurrentRound, "character_positioning", "角色定位", []string{
		fmt.Sprintf("角色: %s", character.Name),
		fmt.Sprintf("角色类型: %s", character.Role),
	})

	return nil
}

// buildRelationshipNetwork 构建关系网络（5-8轮LLM）
func (o *Orchestrator) buildRelationshipNetwork(state *EvolutionState) (*RelationshipNetwork, error) {
	state.CurrentRound++

	// 第1轮：分析所有角色的潜在关系
	prompt := o.buildRelationshipAnalysisPrompt(state)
	systemPrompt := o.buildSystemPrompt("relationship_architect")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("关系分析失败: %w", err)
	}

	var result struct {
		Relationships []struct {
			CharA            string   `json:"char_a"`
			CharB            string   `json:"char_b"`
			RelationType     string   `json:"relation_type"`
			Tension          int      `json:"tension"`
			Description      string   `json:"description"`
			PowerDynamic     string   `json:"power_dynamic"`
			SharedHistory    string   `json:"shared_history"`
			UnspokenTension  string   `json:"unspoken_tension"`
		} `json:"relationships"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析关系分析结果失败: %w", err)
	}

	network := &RelationshipNetwork{
		Nodes:       make(map[string]*CharacterState),
		Edges:       make(map[string]*Relationship),
		NetworkType: state.StoryArchitecture.CharacterRoster.NetworkStructure,
		CenterNode:  "",
	}

	// 创建节点（使用已有的角色）
	for charID, char := range state.Characters {
		network.Nodes[charID] = char
	}

	// 创建边（关系）
	for _, rel := range result.Relationships {
		// 创建关系键（确保一致性）
		relKey := getRelationshipKey(rel.CharA, rel.CharB)

		// 初始化角色的关系状态
		if state.Characters[rel.CharA].Relationships[rel.CharB] == nil {
			state.Characters[rel.CharA].Relationships[rel.CharB] = &RelationshipState{
				TargetCharacterID: rel.CharB,
				VisibleEmotion:    rel.Tension,
				HiddenEmotion:     rel.Tension,
				PowerDynamic:      rel.PowerDynamic,
				SharedHistory:     []string{rel.SharedHistory},
				UnspokenTension:   []string{rel.UnspokenTension},
				SecretsFrom:       []string{},
			}
		}

		if state.Characters[rel.CharB].Relationships[rel.CharA] == nil {
			state.Characters[rel.CharB].Relationships[rel.CharA] = &RelationshipState{
				TargetCharacterID: rel.CharA,
				VisibleEmotion:    rel.Tension,
				HiddenEmotion:     rel.Tension,
				PowerDynamic:      rel.PowerDynamic,
				SharedHistory:     []string{rel.SharedHistory},
				UnspokenTension:   []string{rel.UnspokenTension},
				SecretsFrom:       []string{},
			}
		}

		network.Edges[relKey] = &Relationship{
			From:      rel.CharA,
			To:        rel.CharB,
			Type:      rel.RelationType,
			Tension:   rel.Tension,
			Potential: rel.Description,
		}
	}

	state.logAction(state.CurrentRound, "relationship_network", "关系网络构建", []string{
		fmt.Sprintf("建立关系: %d个", len(result.Relationships)),
	})

	return network, nil
}

// getRelationshipKey 获取关系键（确保两个角色的顺序一致）
func getRelationshipKey(charA, charB string) string {
	if charA < charB {
		return charA + "_" + charB
	}
	return charB + "_" + charA
}

// evolveRelationshipNetwork 演化关系网络（5-10轮LLM）
func (o *Orchestrator) evolveRelationshipNetwork(state *EvolutionState) error {
	state.CurrentRound++

	// 分析关系将如何随故事演化
	prompt := o.buildRelationshipEvolutionPrompt(state)
	systemPrompt := o.buildSystemPrompt("relationship_evolutionist")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return fmt.Errorf("关系演化失败: %w", err)
	}

	var result struct {
		Evolutions []struct {
			RelationID   string   `json:"relation_id"`
			InitialState string   `json:"initial_state"`
			Evolution    []string `json:"evolution"`
			FinalState   string   `json:"final_state"`
			TurningPoint string   `json:"turning_point"`
		} `json:"evolutions"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return fmt.Errorf("解析关系演化结果失败: %w", err)
	}

	state.logAction(state.CurrentRound, "relationship_evolution", "关系网络演化", []string{
		fmt.Sprintf("演化路径数: %d", len(result.Evolutions)),
	})

	return nil
}

// identifyProtagonist 自动识别主角
func (o *Orchestrator) identifyProtagonist(state *EvolutionState) string {
	// 分析：
	// - 谁的关系网络最密集
	// - 谁的内在冲突最复杂
	// - 谁的欲望系统最强

	maxScore := 0
	protagonistID := ""

	for charID, char := range state.Characters {
		score := 0

		// 关系网络密度
		score += len(char.Relationships) * 10

		// 内在冲突复杂度
		score += len(char.InternalConflicts) * 15

		// 欲望系统强度
		if char.Desires.ConsciousWant != "" {
			score += 10
		}
		if char.Desires.UnconsciousNeed != "" {
			score += 15
		}

		// 秘密数量
		score += len(char.Secrets) * 5

		if score > maxScore {
			maxScore = score
			protagonistID = charID
		}
	}

	if protagonistID != "" {
		state.logAction(state.CurrentRound, "protagonist_identification", "主角识别", []string{
			fmt.Sprintf("主角: %s (%s)", state.Characters[protagonistID].Name, protagonistID),
			fmt.Sprintf("得分: %d", maxScore),
		})
	}

	return protagonistID
}

// ============ 阶段3的具体实现 ============

// planForeshadowNetwork 规划伏笔网（5-8轮LLM）
func (o *Orchestrator) planForeshadowNetwork(state *EvolutionState) ([]*ForeshadowPlan, error) {
	state.CurrentRound++

	// 第1轮：识别可能的伏笔类型和位置
	prompt := o.buildForeshadowPlanningPrompt(state)
	systemPrompt := o.buildSystemPrompt("foreshadow_architect")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("伏笔规划失败: %w", err)
	}

	var result struct {
		Foreshadows []struct {
			ID            string   `json:"id"`
			Type          string   `json:"type"`
			Content       string   `json:"content"`
			PlantChapter  int      `json:"plant_chapter"`
			PlantScene    int      `json:"plant_scene"`
			PlantMethod   string   `json:"plant_method"`
			PayoffChapter int      `json:"payoff_chapter"`
			PayoffScene   int      `json:"payoff_scene"`
			PayoffMethod  string   `json:"payoff_method"`
			Connection    string   `json:"connection"`
			Subtlety      int      `json:"subtlety"`
			RelatedThemes []string `json:"related_themes"`
		} `json:"foreshadows"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析伏笔规划结果失败: %w", err)
	}

	// 转换为ForeshadowPlan
	plans := make([]*ForeshadowPlan, 0, len(result.Foreshadows))
	for i, fs := range result.Foreshadows {
		if fs.ID == "" {
			fs.ID = fmt.Sprintf("foreshadow_%d", i)
		}
		plans = append(plans, &ForeshadowPlan{
			ID:            fs.ID,
			Type:          fs.Type,
			Content:       fs.Content,
			PlantChapter:  fs.PlantChapter,
			PlantScene:    fs.PlantScene,
			PlantMethod:   fs.PlantMethod,
			PayoffChapter: fs.PayoffChapter,
			PayoffScene:   fs.PayoffScene,
			PayoffMethod:  fs.PayoffMethod,
			Connection:    fs.Connection,
			Subtlety:      fs.Subtlety,
		})
	}

	state.logAction(state.CurrentRound, "foreshadow_planning", "伏笔网络规划", []string{
		fmt.Sprintf("规划伏笔数: %d", len(plans)),
	})

	return plans, nil
}

// validateForeshadowPlan 验证伏笔的完整性（5-7轮LLM）
func (o *Orchestrator) validateForeshadowPlan(state *EvolutionState, plan []*ForeshadowPlan) error {
	state.CurrentRound++

	// 验证所有伏笔都能被回收
	prompt := o.buildForeshadowValidationPrompt(state, plan)
	systemPrompt := o.buildSystemPrompt("foreshadow_validator")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return fmt.Errorf("伏笔验证失败: %w", err)
	}

	var result struct {
		IsValid      bool     `json:"is_valid"`
		Issues       []string `json:"issues"`
		Suggestions  []string `json:"suggestions"`
		MissingPayoffs []string `json:"missing_payoffs"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return fmt.Errorf("解析伏笔验证结果失败: %w", err)
	}

	state.logAction(state.CurrentRound, "foreshadow_validation", "伏笔完整性验证", []string{
		fmt.Sprintf("验证结果: %v", result.IsValid),
		fmt.Sprintf("发现问题: %d", len(result.Issues)),
	})

	return nil
}

// ============ 阶段4的具体实现 ============

// designCoreConflicts 设计核心冲突（每个冲突4-5轮LLM）
func (o *Orchestrator) designCoreConflicts(state *EvolutionState) ([]*ConflictThread, error) {
	conflicts := make([]*ConflictThread, 0)

	// 根据角色阵容确定冲突数量
	conflictCount := len(state.Characters) + 2 // 角色数量+2个额外冲突

	for i := 0; i < conflictCount; i++ {
		state.CurrentRound++

		// 第1轮：设计单个冲突
		prompt := o.buildConflictDesignPrompt(state, i)
		systemPrompt := o.buildSystemPrompt("conflict_designer")

		response, err := o.engine.callWithRetry(prompt, systemPrompt)
		if err != nil {
			return nil, fmt.Errorf("冲突设计失败(冲突%d): %w", i, err)
		}

		var result struct {
			Type           string   `json:"type"`
			CoreQuestion   string   `json:"core_question"`
			Participants   []string `json:"participants"`
			Stakes         []string `json:"stakes"`
			ThematicRelevance string `json:"thematic_relevance"`
			CurrentIntensity int    `json:"current_intensity"`
			IsExternal     bool     `json:"is_external"`
		}
		if err := json.Unmarshal([]byte(response), &result); err != nil {
			return nil, fmt.Errorf("解析冲突设计结果失败: %w", err)
		}

		conflict := &ConflictThread{
			ID:                 fmt.Sprintf("conflict_%d", i),
			Type:               result.Type,
			CoreQuestion:       result.CoreQuestion,
			Participants:       result.Participants,
			Stakes:             result.Stakes,
			ThematicRelevance:  result.ThematicRelevance,
			CurrentIntensity:   result.CurrentIntensity,
			IsResolved:         false,
			EvolutionPath:      []ConflictStage{},
		}

		// 第2轮：设计冲突演化路径
		state.CurrentRound++
		if err := o.designConflictEvolution(state, conflict); err != nil {
			return nil, err
		}

		conflicts = append(conflicts, conflict)

		state.logAction(state.CurrentRound, "conflict_design", "冲突设计", []string{
			fmt.Sprintf("冲突类型: %s", conflict.Type),
			fmt.Sprintf("核心问题: %s", conflict.CoreQuestion),
		})
	}

	return conflicts, nil
}

// designConflictEvolution 设计冲突演化路径
func (o *Orchestrator) designConflictEvolution(state *EvolutionState, conflict *ConflictThread) error {
	prompt := o.buildConflictEvolutionPrompt(state, conflict)
	systemPrompt := o.buildSystemPrompt("conflict_evolutionist")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return fmt.Errorf("冲突演化设计失败: %w", err)
	}

	var result struct {
		Stages []struct {
			Stage           string   `json:"stage"`
			Description     string   `json:"description"`
			Events          []string `json:"events"`
			EmotionalImpact string   `json:"emotional_impact"`
			ThematicDepth   int      `json:"thematic_depth"`
		} `json:"stages"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return fmt.Errorf("解析冲突演化结果失败: %w", err)
	}

	// 转换为ConflictStage
	for i, stage := range result.Stages {
		conflict.EvolutionPath = append(conflict.EvolutionPath, ConflictStage{
			Stage:           fmt.Sprintf("阶段%d", i+1),
			Description:     stage.Description,
			Intensity:       7, // 默认强度
			Events:          stage.Events,
			EmotionalImpact: make(map[string]string), // 空的map
		})
	}

	return nil
}

// buildConflictHierarchy 构建冲突层级（3-5轮LLM）
func (o *Orchestrator) buildConflictHierarchy(state *EvolutionState) error {
	state.CurrentRound++

	// 分析冲突之间的关系和层级
	prompt := o.buildConflictHierarchyPrompt(state)
	systemPrompt := o.buildSystemPrompt("conflict_hierarchist")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return fmt.Errorf("冲突层级构建失败: %w", err)
	}

	var result struct {
		PrimaryConflicts   []string `json:"primary_conflicts"`
		SecondaryConflicts []string `json:"secondary_conflicts"`
		TertiaryConflicts  []string `json:"tertiary_conflicts"`
		Relationships      []string `json:"relationships"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return fmt.Errorf("解析冲突层级结果失败: %w", err)
	}

	state.logAction(state.CurrentRound, "conflict_hierarchy", "冲突层级构建", []string{
		fmt.Sprintf("主要冲突: %d个", len(result.PrimaryConflicts)),
		fmt.Sprintf("次要冲突: %d个", len(result.SecondaryConflicts)),
	})

	return nil
}

// ============ 阶段5的具体实现 ============

// planStoryDirection 规划故事走向（6-8轮LLM）
func (o *Orchestrator) planStoryDirection(state *EvolutionState) (string, string, error) {
	state.CurrentRound++

	// 第1轮：确定故事开篇
	prompt := o.buildStoryOpeningPrompt(state)
	systemPrompt := o.buildSystemPrompt("story_architect")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return "", "", fmt.Errorf("故事开篇规划失败: %w", err)
	}

	var result struct {
		Opening      string   `json:"opening"`
		Direction    string   `json:"direction"`
		Themes       []string `json:"themes"`
		KeyElements  []string `json:"key_elements"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "", "", fmt.Errorf("解析故事开篇结果失败: %w", err)
	}

	state.logAction(state.CurrentRound, "story_opening", "故事开篇规划", []string{
		fmt.Sprintf("开篇: %s", result.Opening),
		fmt.Sprintf("方向: %s", result.Direction),
	})

	return result.Opening, result.Direction, nil
}

// designKeyEvents 设计关键事件序列（8-10轮LLM）
func (o *Orchestrator) designKeyEvents(state *EvolutionState, opening, direction string) ([]KeyEvent, error) {
	state.CurrentRound++

	// 第1轮：设计关键事件序列
	prompt := o.buildKeyEventsPrompt(state, opening, direction)
	systemPrompt := o.buildSystemPrompt("plot_designer")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("关键事件设计失败: %w", err)
	}

	var result struct {
		Events []struct {
			ID          string   `json:"id"`
			Name        string   `json:"name"`
			Type        string   `json:"type"`
			Chapter     int      `json:"chapter"`
			Description string   `json:"description"`
			Conflicts   []string `json:"conflicts"`
			Characters  []string `json:"characters"`
			Foreshadowing []string `json:"foreshadowing"`
		} `json:"events"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析关键事件结果失败: %w", err)
	}

	events := make([]KeyEvent, 0, len(result.Events))
	for i, event := range result.Events {
		if event.ID == "" {
			event.ID = fmt.Sprintf("event_%d", i)
		}
		events = append(events, KeyEvent{
			ID:                  event.ID,
			Sequence:            i + 1,
			Name:                event.Name,
			Description:         event.Description,
			InvolvedCharacters:  event.Characters,
		})
	}

	state.logAction(state.CurrentRound, "key_events_design", "关键事件设计", []string{
		fmt.Sprintf("事件数: %d", len(events)),
	})

	return events, nil
}

// designClimaxAndResolution 设计高潮和结局
func (o *Orchestrator) designClimaxAndResolution(state *EvolutionState, events []KeyEvent) (string, string, error) {
	state.CurrentRound++

	prompt := o.buildClimaxPrompt(state, events)
	systemPrompt := o.buildSystemPrompt("climax_designer")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return "", "", fmt.Errorf("高潮结局设计失败: %w", err)
	}

	var result struct {
		Climax     string `json:"climax"`
		Resolution string `json:"resolution"`
		Aftermath  string `json:"aftermath"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "", "", fmt.Errorf("解析高潮结局结果失败: %w", err)
	}

	state.logAction(state.CurrentRound, "climax_design", "高潮结局设计", []string{
		fmt.Sprintf("高潮: %s", result.Climax),
		fmt.Sprintf("结局: %s", result.Resolution),
	})

	return result.Climax, result.Resolution, nil
}

// buildForeshadowLinks 构建伏笔链接（事件ID -> 伏笔ID）
func (o *Orchestrator) buildForeshadowLinks(state *EvolutionState, events []KeyEvent) map[string]string {
	links := make(map[string]string)

	// 为每个事件关联相关的伏笔（基于伏笔计划的章节匹配）
	for _, event := range events {
		for _, plan := range state.ForeshadowPlan {
			if plan.PlantChapter == event.Sequence { // 使用Sequence作为章节号
				links[event.ID] = plan.ID
			}
		}
	}

	return links
}

// ============ 阶段6的具体实现 ============

// assignEventsToChapters 将关键事件分配到章节（5-8轮LLM）
func (o *Orchestrator) assignEventsToChapters(state *EvolutionState, chapterCount int) ([]ChapterSynopsis, error) {
	state.CurrentRound++

	// 第1轮：将关键事件分配到章节
	prompt := o.buildChapterAssignmentPrompt(state, chapterCount)
	systemPrompt := o.buildSystemPrompt("chapter_planner")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("章节分配失败: %w", err)
	}

	var result struct {
		Chapters []struct {
			Chapter         int      `json:"chapter"`
			Title           string   `json:"title"`
			Purpose         string   `json:"purpose"`
			KeyEvents       []string `json:"key_events"`
			Conflicts       []string `json:"conflicts"`
			Characters      []string `json:"characters"`
			ArcProgress     string   `json:"arc_progress"`
			EmotionalTone   string   `json:"emotional_tone"`
		} `json:"chapters"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析章节分配结果失败: %w", err)
	}

	chapters := make([]ChapterSynopsis, 0, len(result.Chapters))
	for _, chapter := range result.Chapters {
		chapters = append(chapters, ChapterSynopsis{
			Chapter:            chapter.Chapter,
			Title:              chapter.Title,
			Purpose:            chapter.Purpose,
			KeyEvents:          chapter.KeyEvents,
			RelationshipChanges: []string{}, // 空的
			ForeshadowOps:      ForeshadowOperations{}, // 空的
		})
	}

	state.logAction(state.CurrentRound, "chapter_assignment", "章节分配", []string{
		fmt.Sprintf("章节数: %d", len(chapters)),
	})

	return chapters, nil
}

// refineChapterSequence 确定章节序列和连接（5-7轮LLM）
func (o *Orchestrator) refineChapterSequence(state *EvolutionState, sequence []ChapterSynopsis) error {
	state.CurrentRound++

	// 优化章节之间的连接和过渡
	prompt := o.buildChapterRefinementPrompt(state, sequence)
	systemPrompt := o.buildSystemPrompt("chapter_refiner")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return fmt.Errorf("章节优化失败: %w", err)
	}

	var result struct {
		Transitions    []string `json:"transitions"`
		Pacing         []string `json:"pacing"`
		Improvements   []string `json:"improvements"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return fmt.Errorf("解析章节优化结果失败: %w", err)
	}

	state.logAction(state.CurrentRound, "chapter_refinement", "章节序列优化", []string{
		fmt.Sprintf("过渡数: %d", len(result.Transitions)),
		fmt.Sprintf("改进建议: %d", len(result.Improvements)),
	})

	return nil
}

// ============ 阶段7：细纲生成（按需执行） ============

// GenerateChapterDetailOutline 生成单章的细纲（10-15轮LLM）
func (o *Orchestrator) GenerateChapterDetailOutline(state *EvolutionState, chapterNum int) (*ChapterDetailOutline, error) {
	fmt.Printf("\n🎯 [开始] 生成第%d章细纲...\n", chapterNum)
	fmt.Printf("  当前总轮次: %d\n", state.CurrentRound)

	// 获取章节规划
	if state.ChapterPlan == nil || len(state.ChapterPlan.ChapterSequence) == 0 {
		return nil, fmt.Errorf("章节规划不存在")
	}

	var chapterSynopsis *ChapterSynopsis
	for _, ch := range state.ChapterPlan.ChapterSequence {
		if ch.Chapter == chapterNum {
			chapterSynopsis = &ch
			break
		}
	}

	if chapterSynopsis == nil {
		return nil, fmt.Errorf("未找到章节%d的规划", chapterNum)
	}

	fmt.Printf("  章节标题: %s\n", chapterSynopsis.Title)
	fmt.Printf("  章节目的: %s\n", chapterSynopsis.Purpose)

	// 第1-2轮：设计场景序列
	state.CurrentRound++
	fmt.Printf("\n  [轮次 %d] 设计场景序列...\n", state.CurrentRound)
	sceneSequence, err := o.designSceneSequence(state, chapterSynopsis)
	if err != nil {
		return nil, err
	}
	fmt.Printf("  ✓ 规划了 %d 个场景\n", len(sceneSequence))

	// 第3-10轮：为每个场景生成详细指令
	scenes := make([]*SceneDetailInstruction, 0, len(sceneSequence))
	for i, scene := range sceneSequence {
		state.CurrentRound++
		fmt.Printf("  [轮次 %d] 生成场景%d详情 (%s)...\n", state.CurrentRound, i+1, scene.Type)
		detail, err := o.generateSceneDetailInstruction(state, chapterSynopsis, scene, i)
		if err != nil {
			return nil, fmt.Errorf("场景%d详情生成失败: %w", i, err)
		}
		scenes = append(scenes, detail)
		fmt.Printf("  ✓ 场景%d完成: %s (POV: %s)\n", i+1, detail.Location, detail.POVCharacter)
	}

	// 第11-12轮：追踪角色演化
	state.CurrentRound++
	fmt.Printf("\n  [轮次 %d] 追踪角色演化...\n", state.CurrentRound)
	characterEvolution, err := o.trackChapterCharacterEvolution(state, chapterNum, scenes)
	if err != nil {
		return nil, err
	}
	fmt.Printf("  ✓ 追踪了 %d 个角色的演化\n", len(characterEvolution))

	// 第13-14轮：规划伏笔操作
	state.CurrentRound++
	fmt.Printf("  [轮次 %d] 规划伏笔操作...\n", state.CurrentRound)
	foreshadowTracking, err := o.planChapterForeshadowing(state, chapterNum, scenes)
	if err != nil {
		return nil, err
	}
	fmt.Printf("  ✓ 种植: %d个 | 回收: %d个 | 进行中: %d个\n",
		len(foreshadowTracking.Planted),
		len(foreshadowTracking.PaidOff),
		len(foreshadowTracking.Active))

	// 第15轮：确定章节字数和写作指导
	state.CurrentRound++
	wordCount, guidance := o.estimateChapterMetrics(state, chapterSynopsis, scenes)
	fmt.Printf("  [轮次 %d] 估算字数: %d\n", state.CurrentRound, wordCount)

	outline := &ChapterDetailOutline{
		Chapter:              chapterNum,
		Title:                chapterSynopsis.Title,
		Purpose:              chapterSynopsis.Purpose,
		Tone:                 "中等", // 默认基调
		KeyEvents:            chapterSynopsis.KeyEvents,
		EstimatedWordCount:   wordCount,
		Scenes:               scenes,
		CharacterEvolution:   characterEvolution,
		ForeshadowingTracking: *foreshadowTracking,
	}

	// 将写作指导应用到每个场景
	for _, scene := range scenes {
		scene.WritingGuidance = *guidance
	}

	fmt.Printf("\n✓ 第%d章细纲生成完成! (使用了 %d 轮)\n", chapterNum, state.CurrentRound-(state.CurrentRound-15))

	return outline, nil
}

// designSceneSequence 设计场景序列（2-3轮LLM）
func (o *Orchestrator) designSceneSequence(state *EvolutionState, chapter *ChapterSynopsis) ([]struct {
	Sequence int
	Type     string
	Purpose  string
}, error) {
	prompt := o.buildSceneSequencePrompt(state, chapter)
	systemPrompt := o.buildSystemPrompt("scene_sequence_designer")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("场景序列设计失败: %w", err)
	}

	var result struct {
		Scenes []struct {
			Sequence int    `json:"sequence"`
			Type     string `json:"type"`
			Purpose  string `json:"purpose"`
		} `json:"scenes"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析场景序列结果失败: %w", err)
	}

	scenes := make([]struct {
		Sequence int
		Type     string
		Purpose  string
	}, 0, len(result.Scenes))
	for _, s := range result.Scenes {
		scenes = append(scenes, struct {
			Sequence int
			Type     string
			Purpose  string
		}{
			Sequence: s.Sequence,
			Type:     s.Type,
			Purpose:  s.Purpose,
		})
	}

	return scenes, nil
}

// generateSceneDetailInstruction 生成场景详细指令
func (o *Orchestrator) generateSceneDetailInstruction(state *EvolutionState, chapter *ChapterSynopsis, scene interface{}, index int) (*SceneDetailInstruction, error) {
	prompt := o.buildSceneDetailPrompt(state, chapter, scene, index)
	systemPrompt := o.buildSystemPrompt("scene_detail_designer")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("场景详情生成失败: %w", err)
	}

	var result struct {
		Location            string                       `json:"location"`
		Time                string                       `json:"time"`
		POVCharacter        string                       `json:"pov_character"`
		Characters          []string                     `json:"characters"`
		MainAction          string                       `json:"main_action"`
		DialogueFocus       string                       `json:"dialogue_focus"`
		CharacterChanges    map[string]*CharacterStateChange `json:"character_changes"`
		RelationshipChanges []RelationshipDelta          `json:"relationship_changes"`
		ForeshadowPlant     []ForeshadowPlantInScene    `json:"foreshadow_plant"`
		ForeshadowPayoff    []ForeshadowPayoffInScene   `json:"foreshadow_payoff"`
		Constraints         SceneConstraints             `json:"constraints"`
		Atmosphere          SceneAtmosphere              `json:"atmosphere"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析场景详情结果失败: %w", err)
	}

	sceneSeq := index + 1

	return &SceneDetailInstruction{
		Sequence:              sceneSeq,
		Purpose:               fmt.Sprintf("%v", scene),
		Location:              result.Location,
		Time:                  result.Time,
		POVCharacter:          result.POVCharacter,
		Characters:            result.Characters,
		SceneType:             result.MainAction, // 简化，实际应该有专门的类型字段
		MainAction:            result.MainAction,
		DialogueFocus:         result.DialogueFocus,
		CharacterStateChanges: result.CharacterChanges,
		RelationshipChanges:   result.RelationshipChanges,
		Foreshadowing: ForeshadowInScene{
			Plant:  result.ForeshadowPlant,
			Payoff: result.ForeshadowPayoff,
		},
		Constraints: result.Constraints,
		Atmosphere:  result.Atmosphere,
	}, nil
}

// trackChapterCharacterEvolution 追踪章节角色演化
func (o *Orchestrator) trackChapterCharacterEvolution(state *EvolutionState, chapterNum int, scenes []*SceneDetailInstruction) (map[string]*ChapterCharacterEvolution, error) {
	prompt := o.buildCharacterEvolutionPrompt(state, chapterNum, scenes)
	systemPrompt := o.buildSystemPrompt("character_evolution_tracker")

	response, err := o.engine.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("角色演化追踪失败: %w", err)
	}

	var result struct {
		Evolutions []struct {
			CharacterID         string            `json:"character_id"`
			EmotionalArc        []string          `json:"emotional_arc"`
			GrowthSummary       string            `json:"growth_summary"`
			RelationshipChanges map[string]string `json:"relationship_changes"`
		} `json:"evolutions"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析角色演化结果失败: %w", err)
	}

	evolutionMap := make(map[string]*ChapterCharacterEvolution)
	for _, evo := range result.Evolutions {
		evolutionMap[evo.CharacterID] = &ChapterCharacterEvolution{
			CharacterID:         evo.CharacterID,
			EmotionalArc:        evo.EmotionalArc,
			GrowthSummary:       evo.GrowthSummary,
			RelationshipChanges: evo.RelationshipChanges,
		}
	}

	return evolutionMap, nil
}

// planChapterForeshadowing 规划章节伏笔操作
func (o *Orchestrator) planChapterForeshadowing(state *EvolutionState, chapterNum int, scenes []*SceneDetailInstruction) (*ForeshadowTracking, error) {
	// 分析本章中种植和回收的伏笔
	planted := make([]string, 0)
	paidOff := make([]string, 0)
	active := make([]string, 0)

	// 从伏笔计划中找出本章相关伏笔
	for _, plan := range state.ForeshadowPlan {
		if plan.PlantChapter == chapterNum {
			planted = append(planted, plan.ID)
		}
		if plan.PayoffChapter == chapterNum {
			paidOff = append(paidOff, plan.ID)
		}
		if plan.PlantChapter < chapterNum && plan.PayoffChapter > chapterNum {
			active = append(active, plan.ID)
		}
	}

	return &ForeshadowTracking{
		Planted: planted,
		PaidOff: paidOff,
		Active:  active,
	}, nil
}

// estimateChapterMetrics 估算章节指标
func (o *Orchestrator) estimateChapterMetrics(state *EvolutionState, chapter *ChapterSynopsis, scenes []*SceneDetailInstruction) (int, *WritingGuidance) {
	// 基于场景数量和类型估算字数
	baseWordCount := 3000 // 基础字数
	wordCount := baseWordCount + len(scenes)*500 // 每个场景增加500字

	guidance := &WritingGuidance{
		Techniques:       []string{"展示而非讲述", "感官细节", "节奏变化"},
		DialogueNotes:    "对话要推动情节或揭示角色",
		NarrativeDistance: "中距离",
		StyleHints:       []string{"中等节奏"},
	}

	return wordCount, guidance
}

// ============ 细纲数据结构 ============

// ChapterDetailOutline 章节细纲（给写作器使用）
type ChapterDetailOutline struct {
	Chapter           int                       `json:"chapter"`
	Title             string                    `json:"title"`
	Purpose           string                    `json:"purpose"`
	Tone              string                    `json:"tone"`
	KeyEvents         []string                  `json:"key_events"`
	EstimatedWordCount int                      `json:"estimated_word_count"`

	// 场景序列
	Scenes            []*SceneDetailInstruction  `json:"scenes"`

	// 章节级角色演化总结
	CharacterEvolution map[string]*ChapterCharacterEvolution `json:"character_evolution"`

	// 章节伏笔追踪
	ForeshadowingTracking ForeshadowTracking `json:"foreshadowing_tracking"`
}

// SceneDetailInstruction 场景详细指令
type SceneDetailInstruction struct {
	// 基础信息
	Sequence     int      `json:"sequence"`
	Purpose      string   `json:"purpose"`
	Location     string   `json:"location"`
	Time         string   `json:"time"`
	POVCharacter string   `json:"pov_character"`
	Characters    []string `json:"characters"`
	SceneType    string   `json:"scene_type"` // "对话"/"动作"/"内心"/"过渡"/"描写"

	// 核心指令
	MainAction   string `json:"main_action"`
	DialogueFocus string `json:"dialogue_focus"`

	// 角色状态变化（只记录本章的变化）
	CharacterStateChanges map[string]*CharacterStateChange `json:"character_state_changes"`

	// 关系变化
	RelationshipChanges []RelationshipDelta `json:"relationship_changes"`

	// 伏笔操作
	Foreshadowing ForeshadowInScene `json:"foreshadowing"`

	// 约束
	Constraints SceneConstraints `json:"constraints"`

	// 氛围
	Atmosphere SceneAtmosphere `json:"atmosphere"`

	// 写作指导
	WritingGuidance WritingGuidance `json:"writing_guidance"`
}

// CharacterStateChange 角色状态变化
type CharacterStateChange struct {
	EmotionalChange  string   `json:"emotional_change"`  // 情感变化描述
	NewKnowledge     []string `json:"new_knowledge"`     // 获得的新信息/新疑问
	InternalConflict string   `json:"internal_conflict"` // 内在冲突的变化
}

// RelationshipDelta 关系变化增量
type RelationshipDelta struct {
	Relationship string `json:"relationship"` // "角色A_角色B"
	Change       string `json:"change"`       // "建立"/"加深"/"恶化"/"破裂"/"转化"
	NewTension   int    `json:"new_tension"`   // 新的紧张度（0-100）
}

// ForeshadowInScene 场景中的伏笔操作
type ForeshadowInScene struct {
	Plant  []ForeshadowPlantInScene  `json:"plant"`
	Payoff []ForeshadowPayoffInScene `json:"payoff"`
}

// ForeshadowPlantInScene 种植伏笔
type ForeshadowPlantInScene struct {
	ForeshadowID string `json:"foreshadow_id"`
	Content      string `json:"content"`
	Subtlety    int    `json:"subtlety"`
	Method       string `json:"method"`
}

// ForeshadowPayoffInScene 回收伏笔
type ForeshadowPayoffInScene struct {
	ForeshadowID string `json:"foreshadow_id"`
	Reveals      string `json:"reveals"`
	Method       string `json:"method"`
}

// SceneConstraints 场景约束
type SceneConstraints struct {
	MustInclude    []string `json:"must_include"`    // 必须包含的元素
	MustNotReveal   []string `json:"must_not_reveal"` // 绝对不能透露的信息
	TransitionHint string   `json:"transition_hint"` // 场景结束时的过渡提示
}

// SceneAtmosphere 场景氛围
type SceneAtmosphere struct {
	Mood         string   `json:"mood"`         // 整体情绪基调
	Pacing       string   `json:"pacing"`       // 节奏：缓慢/中等/快速
	SensoryFocus []string `json:"sensory_focus"` // 侧重哪些感官
}

// WritingGuidance 写作指导
type WritingGuidance struct {
	Techniques        []string `json:"techniques"`         // 建议使用的写作技巧
	DialogueNotes     string   `json:"dialogue_notes"`     // 对话指导
	NarrativeDistance  string   `json:"narrative_distance"` // 叙事距离：近距离/中距离/远距离
	StyleHints        []string `json:"style_hints"`       // 风格提示
}

// ChapterCharacterEvolution 章节角色演化
type ChapterCharacterEvolution struct {
	CharacterID   string                   `json:"character_id"`
	EmotionalArc  []string                 `json:"emotional_arc"`  // 情感轨迹：["平静" → "困惑" → "决心"]
	GrowthSummary string                   `json:"growth_summary"` // 成长总结
	RelationshipChanges map[string]string   `json:"relationship_changes"` // 关系变化
}

// ForeshadowTracking 伏笔追踪
type ForeshadowTracking struct {
	Planted  []string `json:"planted"`  // 本章种植的伏笔ID
	PaidOff  []string `json:"paid_off"`  // 本章回收的伏笔ID
	Active   []string `json:"active"`    // 仍未回收的伏笔ID
}

// ============ 辅助方法：Prompt构建 ============

// buildSystemPrompt 构建系统提示词
func (o *Orchestrator) buildSystemPrompt(role string) string {
	systemPrompts := map[string]string{
		"story_architecture_analyzer": `你是一位资深的故事架构分析师，精通各种叙事理论。
你擅长分析世界设定的深层张力，确定最适合的叙事模式。
你的分析总是基于因果逻辑和故事潜力。`,

		"character_roster_planner": `你是一位专业的故事角色策划师。
你擅长根据叙事模式规划角色阵容，确保角色数量和类型适合故事规模。
你理解角色关系网络对故事张力的重要性。`,

		"conflict_architect": `你是一位故事冲突系统设计师。
你擅长识别和设计核心冲突，确保冲突具有足够的张力和演化空间。
你的冲突设计总是与主题紧密相关。`,

		"character_creator": `你是一位专业的故事角色设计师。
你擅长创造立体、复杂、有深度的角色。
你设计的角色总有明确的欲望系统、内在冲突和独特的人格。
你确保每个角色都对故事有独特贡献。`,

		"character_psychologist": `你是一位角色心理分析师。
你擅长挖掘角色的深层内在冲突、秘密和恐惧。
你确保角色的内在冲突与外在情节紧密相连。`,

		"relationship_architect": `你是一位角色关系网络设计师。
你擅长设计复杂、动态的角色关系。
你理解关系张力是故事驱动的核心力量。`,

		"relationship_evolutionist": `你是一位关系演化规划师。
你擅长规划关系如何随故事发展而变化。
你确保关系的演化有因果逻辑和情感冲击力。`,

		"foreshadow_architect": `你是一位伏笔系统设计师。
你擅长设计精巧的伏笔网络，确保伏笔既微妙又有效。
你理解伏笔的种植和回收必须满足读者的期待和惊喜。
你确保所有伏笔都能得到合理的回收。`,

		"foreshadow_validator": `你是一位伏笔系统验证专家。
你擅长检查伏笔计划的完整性和合理性。
你能识别伏笔的遗漏、冲突和时机问题。`,

		"conflict_designer": `你是一位核心冲突设计师。
你擅长设计多层次、有深度的冲突。
你确保每个冲突都有足够的赌注和演化空间。
你设计的冲突与主题紧密相连。`,

		"conflict_evolutionist": `你是一位冲突演化规划师。
你擅长规划冲突如何从建立到升级再到解决。
你确保冲突的每个阶段都有明确的情感冲击和主题深度。`,

		"conflict_hierarchist": `你是一位冲突层级分析师。
你擅长识别主要冲突、次要冲突和背景冲突。
你能理清冲突之间的关系和相互影响。`,

		"story_architect": `你是一位故事架构师。
你擅长设计故事的开篇、走向和关键转折。
你理解故事必须具有因果链和内在逻辑。`,

		"plot_designer": `你是一位情节设计师。
你擅长设计关键事件序列，确保情节紧凑有力。
你理解每个事件都必须推动故事发展或深化角色。`,

		"climax_designer": `你是一位高潮和结局设计师。
你擅长创造令人印象深刻的高潮和令人满意的结局。
你确保高潮是所有冲突的总爆发，结局是主角的真正转变。`,

		"chapter_planner": `你是一位章节规划师。
你擅长将关键事件合理分配到各个章节。
你确保每一章都有明确的目的和进展。`,

		"chapter_refiner": `你是一位章节优化专家。
你擅长优化章节之间的连接和节奏。
你能识别过渡问题并提供改进建议。`,

		"scene_sequence_designer": `你是一位场景序列设计师。
你擅长为章节规划合理的场景序列。
你理解场景类型的变化对节奏的重要性。`,

		"scene_detail_designer": `你是一位场景细节设计师。
你擅长为场景生成详细的写作指令。
你的指令包括地点、时间、角色、动作、对话、情感、氛围等所有要素。
你确保每个场景都有明确的目的和推进作用。`,

		"character_evolution_tracker": `你是一位角色演化追踪师。
你擅长分析角色在章节中的情感轨迹和成长。
你能识别关键的关系变化和内在转变。`,
	}

	if prompt, ok := systemPrompts[role]; ok {
		return prompt
	}
	return "你是一位专业的故事策划师。"
}

// buildWorldAnalysisPrompt 构建世界分析提示词
func (o *Orchestrator) buildWorldAnalysisPrompt(state *EvolutionState) string {
	world := state.WorldContext

	// 提取种族名称
	raceNames := make([]string, 0, len(world.Civilization.Races))
	for _, race := range world.Civilization.Races {
		raceNames = append(raceNames, race.Name)
	}

	// 提取社会冲突
	conflicts := make([]string, 0, len(world.Society.Conflicts))
	for _, conflict := range world.Society.Conflicts {
		conflicts = append(conflicts, conflict.Description)
	}

	return fmt.Sprintf(`分析以下世界设定，识别其核心故事张力和叙事潜力：

世界名称：%s
世界类型：%s
规模：%s
风格：%s

核心哲学问题：%s

政治结构：%s
经济结构：%s

文明特征：
- 种族/群体：%v
- 社会冲突：%v

故事潜力土壤：
%v

请以JSON格式返回：
{
  "core_tensions": ["核心张力1", "核心张力2"],
  "story_potential": ["故事潜力方向1", "故事潜力方向2"],
  "scale": "史诗/宏大/中观/微观",
  "complexity": "极复杂/复杂/中等/简单",
  "suggested_modes": ["群像剧", "个人成长", "抽象力量探索"]
}
只返回JSON，不要包含其他内容。`,
		world.Name,
		world.Type,
		world.Scale,
		world.Style,
		world.Philosophy.CoreQuestion,
		world.Society.Politics.Type,
		world.Society.Economy.Type,
		raceNames,
		conflicts,
		formatPotentialHooks(world.StorySoil.PotentialPlotHooks))
}

// buildModeDeterminationPrompt 构建模式确定提示词
func (o *Orchestrator) buildModeDeterminationPrompt(state *EvolutionState, analysis interface{}) string {
	// 简化版本，实际应该包含完整的分析结果
	return `基于前面的世界分析，从以下叙事模式中选择最合适的一个：

1. 群像剧 - 多个主要角色，复杂的角色网络，交织的故事线
2. 个人成长 - 单一主角的成长弧光，内在冲突驱动
3. 英雄之旅 - 主角踏上冒险，经历试炼和转变
4. 抽象力量探索 - 探索哲学概念或抽象力量的具象化
5. 悬疑推理 - 解谜为主线，层层递进
6. 情感关系 - 以角色关系变化为核心驱动力

请以JSON格式返回：
{
  "selected_mode": "群像剧",
  "reasoning": "选择理由",
  "considerations": ["考虑因素1", "考虑因素2"]
}
只返回JSON，不要包含其他内容。`
}

// buildRosterPlanningPrompt 构建角色阵容规划提示词
func (o *Orchestrator) buildRosterPlanningPrompt(state *EvolutionState, mode string) string {
	world := state.WorldContext

	// 提取种族名称
	raceNames := make([]string, 0, len(world.Civilization.Races))
	for _, race := range world.Civilization.Races {
		raceNames = append(raceNames, race.Name)
	}

	return fmt.Sprintf(`基于以下信息规划角色阵容：

叙事模式：%s

世界设定：
- 世界名称：%s
- 核心问题：%s
- 种族/群体：%v

请规划角色阵容，包括：
1. 总角色数量（适合该叙事模式和世界规模）
2. 主角数量（某些模式可能不需要明确主角）
3. 反派/对抗力量数量
4. 配角数量
5. 角色网络结构（网状/星状/链状等）
6. 关键关系类型（师徒/敌对/亲情/爱情/竞争等）
7. 需要的角色类型（功能型角色）

请以JSON格式返回：
{
  "total_characters": 5,
  "protagonist_count": 1,
  "antagonist_count": 1,
  "supporting_count": 3,
  "network_structure": "网状",
  "key_relationships": ["师徒关系", "敌对关系"],
  "character_types": ["导师型", "对手型"],
  "reasoning": "规划理由"
}
只返回JSON，不要包含其他内容。`,
		mode,
		world.Name,
		world.Philosophy.CoreQuestion,
		raceNames)
}

// buildConflictIdentificationPrompt 构建冲突识别提示词
func (o *Orchestrator) buildConflictIdentificationPrompt(state *EvolutionState, mode string, roster CharacterRosterSpec) string {
	world := state.WorldContext

	// 提取社会冲突
	conflicts := make([]string, 0, len(world.Society.Conflicts))
	for _, conflict := range world.Society.Conflicts {
		conflicts = append(conflicts, conflict.Description)
	}

	return fmt.Sprintf(`识别以下设定中的核心冲突方向：

叙事模式：%s
角色阵容：%d个角色，网络结构：%s

世界设定：
- 核心问题：%s
- 社会冲突：%v

请分析：
1. 主要冲突类型（人与人/与社会/与自己/与自然/与命运）
2. 次要冲突类型
3. 主题核心（冲突指向的深层问题）
4. 冲突方向描述
5. 演化潜力

请以JSON格式返回：
{
  "primary_conflicts": ["人与人：理念的冲突"],
  "secondary_conflicts": ["与自己：欲望与责任的冲突"],
  "thematic_core": "自由意志 vs 宿命",
  "conflict_direction": "多重冲突交织，以理念冲突为主线",
  "reasoning": "分析理由"
}
只返回JSON，不要包含其他内容。`,
		mode,
		roster.TotalCharacters,
		roster.NetworkStructure,
		world.Philosophy.CoreQuestion,
		conflicts)
}

// buildConflictDeepeningPrompt 构建冲突深化提示词
func (o *Orchestrator) buildConflictDeepeningPrompt(state *EvolutionState, result interface{}) string {
	return `基于前面的冲突识别，进一步深化冲突设计：

请：
1. 精炼冲突方向描述（更加具体和有力）
2. 识别冲突的层级（表层/深层/核心层）
3. 规划冲突演化路径（如何升级、转折、解决）

请以JSON格式返回：
{
  "refined_direction": "精炼后的冲突方向描述",
  "conflict_layers": ["表层：具体利益冲突", "深层：价值观冲突", "核心层：存在主义冲突"],
  "evolution_path": ["冲突建立", "冲突升级", "冲突激化", "冲突转折", "冲突解决"]
}
只返回JSON，不要包含其他内容。`
}

// formatPotentialHooks 格式化故事钩子
func formatPotentialHooks(hooks []models.PlotHook) string {
	if len(hooks) == 0 {
		return "暂无"
	}
	result := make([]string, len(hooks))
	for i, hook := range hooks {
		result[i] = fmt.Sprintf("- [%s] %s: %s", hook.Type, hook.Description, hook.StoryPotential)
	}
	return strings.Join(result, "\n")
}

// ============ 阶段2 Prompt构建方法 ============

// buildCharacterCreationPrompt 构建角色创建提示词
func (o *Orchestrator) buildCharacterCreationPrompt(state *EvolutionState, index int) string {
	roster := state.StoryArchitecture.CharacterRoster
	mode := state.StoryArchitecture.NarrativeMode
	world := state.WorldContext

	// 提取种族名称
	raceNames := make([]string, 0, len(world.Civilization.Races))
	for _, race := range world.Civilization.Races {
		raceNames = append(raceNames, race.Name)
	}

	return fmt.Sprintf(`基于以下信息创建第%d个角色：

叙事模式：%s
角色阵容规划：
- 总角色数：%d
- 主角数：%d
- 反派数：%d
- 配角数：%d
- 关键关系：%v

世界设定：
- 世界名称：%s
- 世界类型：%s
- 世界规模：%s
- 风格：%s
- 核心问题：%s
- 种族：%v

%请根据世界类型和风格创建符合时代背景的角色。
例如：
- 历史类（民国、古代）：姓名应符合时代特征，避免现代或奇幻风格
- 奇幻类：姓名可以带有魔法或神秘元素
- 科幻类：姓名应反映未来科技特征
- 现实类：姓名应贴近现实生活

已创建的角色：
%s

请创建一个独特且有深度的角色，包括：
1. 姓名和角色定位
2. 年龄和背景
3. 性格特征
4. 意识欲望（表面想要什么）
5. 潜意识需求（深层需要什么）
6. 核心特质
7. 致命弱点

请以JSON格式返回：
{
  "name": "角色名",
  "role": "主角/反派/配角/导师/对手",
  "age": 25,
  "background": "背景故事",
  "personality": ["特质1", "特质2"],
  "conscious_want": "意识欲望",
  "unconscious_need": "潜意识需求",
  "core_traits": ["核心特质"],
  "flaws": ["弱点"]
}
只返回JSON，不要包含其他内容。`,
		index+1,
		mode,
		roster.TotalCharacters,
		roster.ProtagonistCount,
		roster.AntagonistCount,
		roster.SupportingCount,
		roster.KeyRelationships,
		world.Name,
		world.Type,
		world.Scale,
		world.Style,
		world.Philosophy.CoreQuestion,
		raceNames,
		formatExistingCharacters(state.Characters))
}

// buildCharacterDeepeningPrompt 构建角色深化提示词
func (o *Orchestrator) buildCharacterDeepeningPrompt(state *EvolutionState, character *CharacterState) string {
	world := state.WorldContext
	return fmt.Sprintf(`深化角色的内在冲突和秘密：

角色信息：
- 姓名：%s
- 角色：%s
- 意识欲望：%s
- 潜意识需求：%s

世界背景：%s
故事风格：%s

⚠️ 重要要求：
1. 要让角色有深度、有缺陷、有人性复杂性
2. 恐惧要具体、深刻、能驱动角色行为
3. 秘密要有爆炸性、能改变关系
4. 冲突要尖锐、无法轻易解决
5. 不要平庸、不要俗套、不要完美

请分析并深化：
1. 内在冲突（不同欲望/价值观之间的尖锐冲突，至少2个）
2. 秘密（对他人隐瞒的爆炸性秘密，至少1个）
3. 恐惧（最害怕的具体事物，要深刻、要影响行为，至少1个）
4. 情感触发点（什么会让角色情绪失控，至少1个）
5. 伪装行为（角色如何隐藏真实自我，至少1个）
6. 欲望与需求的差距（表面想要与深层需要之间的矛盾）

请以JSON格式返回：
{
  "internal_conflicts": ["冲突1：具体描述", "冲突2：具体描述"],
  "secrets": ["秘密1：具体、有爆炸性"],
  "fears": ["恐惧1：具体、深刻、影响行为"],
  "triggers": ["触发点1：什么会让角色失控"],
  "masking_behaviors": ["伪装行为1：如何隐藏真实自我"],
  "want_vs_need_gap": "欲望与需求的具体矛盾"
}
只返回JSON，不要包含其他内容。`,
		character.Name,
		character.Role,
		character.Desires.ConsciousWant,
		character.Desires.UnconsciousNeed,
		world.Type,
		world.Style)
}

// buildRelationshipAnalysisPrompt 构建关系分析提示词
func (o *Orchestrator) buildRelationshipAnalysisPrompt(state *EvolutionState) string {
	world := state.WorldContext

	characterList := make([]string, 0, len(state.Characters))
	for charID, char := range state.Characters {
		characterList = append(characterList, fmt.Sprintf("- %s (%s): %s, 欲望:%s",
			char.Name, charID, char.Role, char.Desires.ConsciousWant))
	}

	return fmt.Sprintf(`分析以下角色之间应该建立什么样的关系：

叙事模式：%s
核心冲突方向：%s

角色列表：
%s

世界设定背景：
- 世界类型：%s
- 风格：%s
- 核心问题：%s
%s

⚠️ 重要要求：
1. 关系要有张力，不要平淡
2. 要有权力斗争、地位差异、影响力不对等
3. 要有复杂的历史和情感纠葛
4. 要有未言明的紧张感和潜在冲突
5. 避免所有关系都是和谐的，要有矛盾和冲突

请分析每对角色之间的关系，包括：
1. 关系类型（师徒/敌对/亲情/爱情/竞争/友谊/背叛等）
2. 初始紧张度（0-100，要有变化，不要都是30-50）
3. 关系描述（具体、有张力）
4. 权力动态（谁占主导、谁处于劣势、如何变化的）
5. 共同历史（一起经历过什么，形成情感基础）
6. 未言明的紧张感（有什么话没说出口、有什么秘密、有什么潜在冲突）

请以JSON格式返回：
{
  "relationships": [
    {
      "char_a": "char_0",
      "char_b": "char_1",
      "relation_type": "师徒",
      "tension": 30,
      "description": "关系描述",
      "power_dynamic": "具体的权力动态：比如A虽然地位高但被B抓住把柄，或者B表面顺从但暗中掌控",
      "shared_history": "具体的共同经历：比如一起经历过某事件，形成了情感基础或创伤",
      "unspoken_tension": "未言明的紧张感：比如有什么话没说、有什么秘密、有什么潜在冲突"
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		state.StoryArchitecture.NarrativeMode,
		state.StoryArchitecture.CoreConflictType,
		strings.Join(characterList, "\n"),
		world.Type,
		world.Style,
		world.Philosophy.CoreQuestion,
		world.Philosophy.CoreQuestion)
}

// buildRelationshipEvolutionPrompt 构建关系演化提示词
func (o *Orchestrator) buildRelationshipEvolutionPrompt(state *EvolutionState) string {
	relationships := make([]string, 0)
	for _, edge := range state.RelationshipNetwork.Edges {
		relationships = append(relationships, fmt.Sprintf("- %s -> %s: %s",
			state.Characters[edge.From].Name,
			state.Characters[edge.To].Name,
			edge.Type))
	}

	return fmt.Sprintf(`规划关系网络的演化路径：

当前关系：
%s

请分析这些关系将如何随故事发展：
1. 初始状态
2. 演化阶段（如何变化）
3. 最终状态
4. 关键转折点

请以JSON格式返回：
{
  "evolutions": [
    {
      "relation_id": "角色A_角色B",
      "initial_state": "初始状态",
      "evolution": ["阶段1", "阶段2", "阶段3"],
      "final_state": "最终状态",
      "turning_point": "转折点事件"
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		strings.Join(relationships, "\n"))
}

// formatExistingCharacters 格式化已存在的角色列表
func formatExistingCharacters(characters map[string]*CharacterState) string {
	if len(characters) == 0 {
		return "无"
	}

	result := make([]string, 0, len(characters))
	for _, char := range characters {
		result = append(result, fmt.Sprintf("- %s: %s, 欲望:%s",
			char.Name, char.Role, char.Desires.ConsciousWant))
	}
	return strings.Join(result, "\n")
}

// ============ 阶段3-7 Prompt构建方法 ============

// buildForeshadowPlanningPrompt 构建伏笔规划提示词
func (o *Orchestrator) buildForeshadowPlanningPrompt(state *EvolutionState) string {
	world := state.WorldContext

	// 阶段3时ChapterPlan还未创建，所以我们需要传入章节总数
	// 暂时使用默认值12章
	totalChapters := 12
	if state.ChapterPlan != nil && state.ChapterPlan.TotalChapters > 0 {
		totalChapters = state.ChapterPlan.TotalChapters
	}

	return fmt.Sprintf(`基于以下故事规划伏笔网络：

核心问题：%s
叙事模式：%s
主要冲突：%s
角色数量：%d
预计章节数：%d

⚠️ 重要要求：
1. 每个伏笔都必须充满悬念和张力，不要平淡无奇
2. 伏笔要能引发读者的好奇和猜测
3. 要有反转、惊喜、震撼的效果
4. 植入要巧妙自然，不要生硬
5. 回收要出人意料但情理之中

请设计5-10个伏笔，包括：
1. 伏笔类型（情节/角色/主题/象征）
2. 伏笔内容（具体、详细、有悬念）
3. 种植章节和场景
4. 种植方法（如何巧妙地植入）
5. 回收章节和场景
6. 回收方法（如何震撼地揭示）
7. 细腻程度（1-10）
8. 连接逻辑（从种植到回收的演进）

请以JSON格式返回：
{
  "foreshadows": [
    {
      "id": "foreshadow_1",
      "type": "情节",
      "content": "详细描述这个伏笔，要充满悬念和张力",
      "plant_chapter": 2,
      "plant_scene": 1,
      "plant_method": "如何巧妙植入：比如通过一段看似平常的对话、一个不起眼的物品、一个微妙的表情变化",
      "payoff_chapter": 8,
      "payoff_scene": 3,
      "payoff_method": "如何震撼揭示：比如通过一个惊人的发现、一个意外的反转、一个情感的爆发",
      "connection": "从种植到回收的逻辑演进，制造悬念和期待",
      "subtlety": 7
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		world.Philosophy.CoreQuestion,
		state.StoryArchitecture.NarrativeMode,
		state.StoryArchitecture.CoreConflictType,
		len(state.Characters),
		totalChapters)
}

// buildForeshadowValidationPrompt 构建伏笔验证提示词
func (o *Orchestrator) buildForeshadowValidationPrompt(state *EvolutionState, plan []*ForeshadowPlan) string {
	foreshadowList := make([]string, 0, len(plan))
	for _, fs := range plan {
		foreshadowList = append(foreshadowList, fmt.Sprintf("- %s: 第%d章种植, 第%d章回收",
			fs.Content, fs.PlantChapter, fs.PayoffChapter))
	}

	return fmt.Sprintf(`验证以下伏笔计划的完整性：

伏笔列表：
%s

请检查：
1. 所有伏笔都有回收吗？
2. 伏笔的时机是否合理？
3. 是否有遗漏或冲突？

请以JSON格式返回：
{
  "is_valid": true,
  "issues": ["问题1", "问题2"],
  "suggestions": ["建议1", "建议2"],
  "missing_payoffs": ["伏笔ID"]
}
只返回JSON，不要包含其他内容。`,
		strings.Join(foreshadowList, "\n"))
}

// buildConflictDesignPrompt 构建冲突设计提示词
func (o *Orchestrator) buildConflictDesignPrompt(state *EvolutionState, index int) string {
	world := state.WorldContext

	characters := make([]string, 0, len(state.Characters))
	for charID, char := range state.Characters {
		characters = append(characters, fmt.Sprintf("- %s (%s): %s",
			char.Name, charID, char.Role))
	}

	return fmt.Sprintf(`设计第%d个核心冲突：

世界核心问题：%s
已有冲突数：%d
核心冲突方向：%s

角色列表：
%s

请设计一个独特且有力的冲突，包括：
1. 冲突类型（人与人/与社会/与自己/与自然/与命运）
2. 核心问题
3. 参与者
4. 赌注（如果失败会怎样？）
5. 与主题的关联
6. 当前强度（0-100）
7. 是否为外部冲突

请以JSON格式返回：
{
  "type": "人与人",
  "core_question": "核心问题",
  "participants": ["char_0", "char_1"],
  "stakes": ["赌注1", "赌注2"],
  "thematic_relevance": "主题关联",
  "current_intensity": 60,
  "is_external": true
}
只返回JSON，不要包含其他内容。`,
		index+1,
		world.Philosophy.CoreQuestion,
		index,
		state.StoryArchitecture.CoreConflictType,
		strings.Join(characters, "\n"))
}

// buildConflictEvolutionPrompt 构建冲突演化提示词
func (o *Orchestrator) buildConflictEvolutionPrompt(state *EvolutionState, conflict *ConflictThread) string {
	return fmt.Sprintf(`设计冲突的演化路径：

冲突类型：%s
核心问题：%s
参与者：%v

请规划这个冲突将如何演化：
1. 冲突建立的阶段
2. 冲突升级的阶段
3. 冲突高潮的阶段
4. 冲突解决（或转化）的阶段

每个阶段要包括：
- 阶段描述
- 关键事件
- 情感冲击
- 主题深度（0-10）

请以JSON格式返回：
{
  "stages": [
    {
      "stage": "阶段1",
      "description": "描述",
      "events": ["事件1", "事件2"],
      "emotional_impact": "情感冲击",
      "thematic_depth": 7
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		conflict.Type,
		conflict.CoreQuestion,
		conflict.Participants)
}

// buildConflictHierarchyPrompt 构建冲突层级提示词
func (o *Orchestrator) buildConflictHierarchyPrompt(state *EvolutionState) string {
	conflicts := make([]string, 0, len(state.Conflicts))
	for _, c := range state.Conflicts {
		conflicts = append(conflicts, fmt.Sprintf("- %s: %s", c.ID, c.CoreQuestion))
	}

	return fmt.Sprintf(`分析冲突之间的层级关系：

所有冲突：
%s

请分类：
1. 主要冲突（推动主线）
2. 次要冲突（支线）
3. 三级冲突（背景冲突）
4. 冲突之间的关系

请以JSON格式返回：
{
  "primary_conflicts": ["conflict_0"],
  "secondary_conflicts": ["conflict_1"],
  "tertiary_conflicts": ["conflict_2"],
  "relationships": ["冲突0推动冲突1", "冲突2是冲突0的背景"]
}
只返回JSON，不要包含其他内容。`,
		strings.Join(conflicts, "\n"))
}

// buildStoryOpeningPrompt 构建故事开篇提示词
func (o *Orchestrator) buildStoryOpeningPrompt(state *EvolutionState) string {
	world := state.WorldContext
	protagonist := ""
	if state.RelationshipNetwork.CenterNode != "" {
		if char, ok := state.Characters[state.RelationshipNetwork.CenterNode]; ok {
			protagonist = char.Name
		}
	}

	return fmt.Sprintf(`规划故事开篇：

世界设定：%s
核心问题：%s
主角：%s
叙事模式：%s

请确定：
1. 故事如何开始（开篇情境）
2. 故事走向
3. 关键主题
4. 必须包含的元素

请以JSON格式返回：
{
  "opening": "开篇描述",
  "direction": "故事走向",
  "themes": ["主题1", "主题2"],
  "key_elements": ["元素1", "元素2"]
}
只返回JSON，不要包含其他内容。`,
		world.Name,
		world.Philosophy.CoreQuestion,
		protagonist,
		state.StoryArchitecture.NarrativeMode)
}

// buildKeyEventsPrompt 构建关键事件提示词
func (o *Orchestrator) buildKeyEventsPrompt(state *EvolutionState, opening, direction string) string {
	conflicts := make([]string, 0, len(state.Conflicts))
	for _, c := range state.Conflicts {
		conflicts = append(conflicts, fmt.Sprintf("- %s: %s", c.ID, c.CoreQuestion))
	}

	foreshadows := make([]string, 0, len(state.ForeshadowPlan))
	for _, fs := range state.ForeshadowPlan {
		foreshadows = append(foreshadows, fmt.Sprintf("- %s: 第%d章", fs.ID, fs.PlantChapter))
	}

	return fmt.Sprintf(`设计关键事件序列：

开篇：%s
方向：%s

冲突列表：
%s

伏笔列表：
%s

请设计8-15个关键事件，包括：
1. 事件ID
2. 事件名称
3. 事件类型（激励/试炼/转折/高潮/情节点等）
4. 预计章节
5. 事件描述
6. 关联的冲突
7. 参与角色
8. 关联的伏笔

请以JSON格式返回：
{
  "events": [
    {
      "id": "event_1",
      "name": "事件名",
      "type": "激励事件",
      "chapter": 1,
      "description": "事件描述",
      "conflicts": ["conflict_0"],
      "characters": ["char_0"],
      "foreshadowing": ["foreshadow_1"]
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		opening,
		direction,
		strings.Join(conflicts, "\n"),
		strings.Join(foreshadows, "\n"))
}

// buildClimaxPrompt 构建高潮结局提示词
func (o *Orchestrator) buildClimaxPrompt(state *EvolutionState, events []KeyEvent) string {
	eventSummary := make([]string, 0, len(events))
	for _, e := range events {
		eventSummary = append(eventSummary, fmt.Sprintf("- %s: %s", e.Name, e.Description))
	}

	return fmt.Sprintf(`设计高潮和结局：

关键事件概览：
%s

请设计：
1. 高潮（所有冲突的最终对决）
2. 结局（冲突解决，主角变化）
3. 余韵（世界变成怎样）

请以JSON格式返回：
{
  "climax": "高潮描述",
  "resolution": "结局描述",
  "aftermath": "余韵描述"
}
只返回JSON，不要包含其他内容。`,
		strings.Join(eventSummary, "\n"))
}

// buildChapterAssignmentPrompt 构建章节分配提示词
func (o *Orchestrator) buildChapterAssignmentPrompt(state *EvolutionState, chapterCount int) string {
	events := make([]string, 0, len(state.GlobalOutline.KeyEvents))
	for _, e := range state.GlobalOutline.KeyEvents {
		events = append(events, fmt.Sprintf("- 第%d章: %s",
			e.Sequence, e.Name))
	}

	return fmt.Sprintf(`将关键事件分配到%d个章节：

关键事件：
%s

开篇：%s
高潮：%s
结局：%s

请为每一章规划：
1. 章节编号
2. 章节标题
3. 章节目的
4. 包含的关键事件

请以JSON格式返回：
{
  "chapters": [
    {
      "chapter": 1,
      "title": "章节标题",
      "purpose": "章节目的",
      "key_events": ["event_1"]
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		chapterCount,
		strings.Join(events, "\n"),
		state.GlobalOutline.Opening,
		state.GlobalOutline.Climax,
		state.GlobalOutline.Resolution)
}

// buildChapterRefinementPrompt 构建章节优化提示词
func (o *Orchestrator) buildChapterRefinementPrompt(state *EvolutionState, sequence []ChapterSynopsis) string {
	chapterSummary := make([]string, 0, len(sequence))
	for _, ch := range sequence {
		chapterSummary = append(chapterSummary, fmt.Sprintf("- 第%d章 %s: %s",
			ch.Chapter, ch.Title, ch.Purpose))
	}

	return fmt.Sprintf(`优化章节序列和连接：

章节序列：
%s

请分析：
1. 章节之间的过渡是否流畅？
2. 节奏是否合理？
3. 有什么改进建议？

请以JSON格式返回：
{
  "transitions": ["第1-2章过渡建议"],
  "pacing": ["节奏分析"],
  "improvements": ["改进建议"]
}
只返回JSON，不要包含其他内容。`,
		strings.Join(chapterSummary, "\n"))
}

// buildSceneSequencePrompt 构建场景序列提示词
func (o *Orchestrator) buildSceneSequencePrompt(state *EvolutionState, chapter *ChapterSynopsis) string {
	return fmt.Sprintf(`设计第%d章的场景序列：

章节标题：%s
章节目的：%s
关键事件：%v

请规划3-6个场景，每个场景包括：
1. 场景序号
2. 场景类型（对话/动作/内心/过渡/描写）
3. 场景目的

请以JSON格式返回：
{
  "scenes": [
    {
      "sequence": 1,
      "type": "对话",
      "purpose": "场景目的"
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		chapter.Chapter,
		chapter.Title,
		chapter.Purpose,
		chapter.KeyEvents)
}

// buildSceneDetailPrompt 构建场景详情提示词
func (o *Orchestrator) buildSceneDetailPrompt(state *EvolutionState, chapter *ChapterSynopsis, scene interface{}, index int) string {
	return fmt.Sprintf(`生成第%d章第%d个场景的详细指令：

章节目：%s
场景类型：%v

请生成场景的详细写作指令，包括：
1. 地点
2. 时间
3. POV角色
4. 在场角色
5. 主要动作
6. 对话重点
7. 角色状态变化
8. 关系变化
9. 伏笔操作（种植/回收）
10. 场景约束
11. 氛围描写

请以JSON格式返回：
{
  "location": "地点",
  "time": "时间",
  "pov_character": "char_0",
  "characters": ["char_0", "char_1"],
  "main_action": "主要动作描述",
  "dialogue_focus": "对话重点",
  "character_changes": {
    "char_0": {
      "emotional_change": "情感变化",
      "new_knowledge": ["新信息"],
      "internal_conflict": "内在冲突"
    }
  },
  "relationship_changes": [
    {
      "relationship": "char_0_char_1",
      "change": "加深",
      "new_tension": 50
    }
  ],
  "foreshadow_plant": [
    {
      "foreshadow_id": "foreshadow_1",
      "content": "内容",
      "subtlety": 7,
      "method": "方法"
    }
  ],
  "foreshadow_payoff": [],
  "constraints": {
    "must_include": ["必须包含"],
    "must_not_reveal": ["绝不能透露"],
    "transition_hint": "过渡提示"
  },
  "atmosphere": {
    "mood": "情绪",
    "pacing": "中等",
    "sensory_focus": ["视觉", "听觉"]
  }
}
只返回JSON，不要包含其他内容。`,
		chapter.Chapter,
		index+1,
		chapter.Purpose,
		scene)
}

// buildCharacterEvolutionPrompt 构建角色演化提示词
func (o *Orchestrator) buildCharacterEvolutionPrompt(state *EvolutionState, chapterNum int, scenes []*SceneDetailInstruction) string {
	return fmt.Sprintf(`追踪第%d章的角色演化：

场景数：%d

请分析本章中主要角色的演化：
1. 情感轨迹（起始→中间→结束）
2. 成长总结
3. 关系变化

请以JSON格式返回：
{
  "evolutions": [
    {
      "character_id": "char_0",
      "emotional_arc": ["平静", "困惑", "决心"],
      "growth_summary": "成长总结",
      "relationship_changes": {
        "char_1": "关系从敌对转变为复杂"
      }
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		chapterNum,
		len(scenes))
}
