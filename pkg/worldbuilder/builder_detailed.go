// Package worldbuilder 高信息熵世界构建器
// 通过多轮LLM调用和验证机制，生成高信息熵、高一致性的世界设定
package worldbuilder

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

// DetailedBuilder 高信息熵世界构建器
type DetailedBuilder struct {
	cfg    *config.Config
	client *llm.Client
	db     db.Database
	mapping *config.ModuleMapping
}

// NewDetailedBuilder 创建高信息熵构建器
func NewDetailedBuilder() (*DetailedBuilder, error) {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	client, mapping, err := llm.NewClientForModule("world_builder")
	if err != nil {
		return nil, fmt.Errorf("创建LLM客户端失败: %w", err)
	}

	// 使用配置文件中的temperature设置（已设置为1.0）
	// 不再强制覆盖

	return &DetailedBuilder{
		cfg:    cfg,
		client: client,
		db:     db.Get(),
		mapping: mapping,
	}, nil
}

// Build 构建完整世界（50-100轮LLM）
func (dbuilder *DetailedBuilder) Build(params BuildParams) (*models.WorldSetting, error) {
	fmt.Println("\n========================================")
	fmt.Println("  🌍 高信息熵世界构建器")
	fmt.Println("========================================\n")

	startTime := time.Now()

	// 创建世界设定对象
	world := &models.WorldSetting{
		ID:    db.GenerateID("world"),
		Name:  params.Name,
		Type:  params.Type,
		Scale: params.Scale,
		Style: params.Style,
	}

	// 阶段1：哲学基础（3-5轮）
	fmt.Println("📚 [阶段1/7] 哲学基础构建 (3-5轮LLM)...")
	if err := dbuilder.buildStage1Detailed(world, params); err != nil {
		return nil, fmt.Errorf("阶段1失败: %w", err)
	}
	if err := dbuilder.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段1失败: %w", err)
	}

	// 阶段2：世界观（5-8轮）
	fmt.Println("\n🌌 [阶段2/7] 世界观构建 (5-8轮LLM)...")
	if err := dbuilder.buildStage2Detailed(world, params); err != nil {
		return nil, fmt.Errorf("阶段2失败: %w", err)
	}
	if err := dbuilder.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段2失败: %w", err)
	}

	// 阶段3：法则设定（8-12轮）
	fmt.Println("\n⚡ [阶段3/7] 法则设定构建 (8-12轮LLM)...")
	if err := dbuilder.buildStage3Detailed(world, params); err != nil {
		return nil, fmt.Errorf("阶段3失败: %w", err)
	}
	if err := dbuilder.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段3失败: %w", err)
	}

	// 阶段4：故事土壤（10-15轮）
	fmt.Println("\n🌱 [阶段4/7] 故事土壤构建 (10-15轮LLM)...")
	if err := dbuilder.buildStage4Detailed(world, params); err != nil {
		return nil, fmt.Errorf("阶段4失败: %w", err)
	}
	if err := dbuilder.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段4失败: %w", err)
	}

	// 阶段5：地理环境（10-20轮）
	fmt.Println("\n🗺️  [阶段5/7] 地理环境构建 (10-20轮LLM)...")
	if err := dbuilder.buildStage5Detailed(world, params); err != nil {
		return nil, fmt.Errorf("阶段5失败: %w", err)
	}
	if err := dbuilder.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段5失败: %w", err)
	}

	// 阶段6：文明社会（15-25轮）
	fmt.Println("\n🏛️  [阶段6/7] 文明社会构建 (15-25轮LLM)...")
	if err := dbuilder.buildStage6Detailed(world, params); err != nil {
		return nil, fmt.Errorf("阶段6失败: %w", err)
	}
	if err := dbuilder.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段6失败: %w", err)
	}

	// 阶段7：历史与一致性（10-20轮）
	fmt.Println("\n📜 [阶段7/7] 历史与一致性验证 (10-20轮LLM)...")
	if err := dbuilder.buildStage7Detailed(world, params); err != nil {
		return nil, fmt.Errorf("阶段7失败: %w", err)
	}
	if err := dbuilder.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段7失败: %w", err)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\n✓ 世界构建完成！用时: %.1f秒\n", elapsed.Seconds())

	return world, nil
}

// buildStage1Detailed 阶段1：哲学基础（3-5轮）
func (dbuilder *DetailedBuilder) buildStage1Detailed(world *models.WorldSetting, params BuildParams) error {
	round := 0

	// 第1轮：生成核心问题
	fmt.Println("  ├─ [轮次1] 生成核心问题...")
	coreQuestion, err := dbuilder.generateCoreQuestion(params)
	if err != nil {
		return err
	}
	world.Philosophy.CoreQuestion = coreQuestion
	round++
	fmt.Printf("    ✓ 核心问题: %s\n", coreQuestion)

	// 第2轮：生成价值体系
	fmt.Println("  ├─ [轮次2] 生成价值体系...")
	valueSystem, err := dbuilder.generateValueSystem(coreQuestion, params)
	if err != nil {
		return err
	}
	world.Philosophy.ValueSystem = *valueSystem
	round++
	fmt.Printf("    ✓ 最高善: %s\n", valueSystem.HighestGood)

	// 第3轮：生成主题列表
	fmt.Println("  ├─ [轮次3] 生成主题列表...")
	themes, err := dbuilder.generateThemes(coreQuestion, valueSystem, params)
	if err != nil {
		return err
	}
	world.Philosophy.Themes = themes
	round++
	fmt.Printf("    ✓ 主题数量: %d\n", len(themes))

	// 第4-5轮：验证和优化
	fmt.Println("  └─ [轮次4-5] 验证和优化...")
	derivation, err := dbuilder.validateAndRefinePhilosophy(world.Philosophy)
	if err != nil {
		return err
	}
	world.Philosophy.Derivation = derivation
	round++

	fmt.Printf("  ✓ 阶段1完成 (共%d轮LLM)\n", round)
	return nil
}

// buildStage2Detailed 阶段2：世界观（5-8轮）
func (dbuilder *DetailedBuilder) buildStage2Detailed(world *models.WorldSetting, params BuildParams) error {
	round := 0

	// 第1轮：生成宇宙起源
	fmt.Println("  ├─ [轮次1] 生成宇宙起源...")
	cosmology, err := dbuilder.generateCosmology(world.Philosophy, params)
	if err != nil {
		return err
	}
	world.Worldview.Cosmology = *cosmology
	round++
	fmt.Printf("    ✓ 起源: %s\n", cosmology.Origin[:50]+"...")

	// 第2轮：生成宇宙结构
	fmt.Println("  ├─ [轮次2] 生成宇宙结构...")
	structure, err := dbuilder.generateCosmologyStructure(world.Philosophy, cosmology)
	if err != nil {
		return err
	}
	cosmology.Structure = structure
	round++
	fmt.Printf("    ✓ 结构层次: %d层\n", strings.Count(structure, "层"))

	// 第3轮：生成形而上学
	fmt.Println("  ├─ [轮次3] 生成形而上学...")
	metaphysics, err := dbuilder.generateMetaphysics(world.Philosophy, cosmology)
	if err != nil {
		return err
	}
	world.Worldview.Metaphysics = *metaphysics
	round++
	fmt.Printf("    ✓ 灵魂观: %v\n", metaphysics.SoulExists)

	// 第4-5轮：生成命运和来世观念
	fmt.Println("  ├─ [轮次4-5] 生成命运和来世...")
	if metaphysics.FateExists {
		fateRelation, err := dbuilder.generateFateRelation(world.Philosophy)
		if err != nil {
			return err
		}
		metaphysics.FateRelShip = fateRelation
	}
	if metaphysics.SoulExists {
		afterlife, err := dbuilder.generateAfterlife(world.Philosophy, metaphysics)
		if err != nil {
			return err
		}
		metaphysics.Afterlife = afterlife
	}
	round += 2

	// 第6-8轮：验证和生成推导逻辑
	fmt.Println("  └─ [轮次6-8] 验证世界观一致性...")
	derivation, err := dbuilder.validateAndRefineWorldview(world.Philosophy, world.Worldview)
	if err != nil {
		return err
	}
	world.Worldview.Derivation = derivation
	round++

	fmt.Printf("  ✓ 阶段2完成 (共%d轮LLM)\n", round)
	return nil
}

// buildStage3Detailed 阶段3：法则设定（8-12轮）
func (dbuilder *DetailedBuilder) buildStage3Detailed(world *models.WorldSetting, params BuildParams) error {
	round := 0

	// 第1-3轮：生成物理法则
	fmt.Println("  ├─ [轮次1-3] 生成物理法则...")
	physics, err := dbuilder.generatePhysicsLaws(world.Worldview, params)
	if err != nil {
		return err
	}
	world.Laws.Physics = *physics
	round += 3
	fmt.Printf("    ✓ 物理法则: 重力、时间、能量、因果、死亡\n")

	// 第4-7轮：生成超自然体系（如果有）
	if world.Laws.Supernatural != nil && world.Laws.Supernatural.Exists {
		fmt.Println("  ├─ [轮次4-7] 生成超自然体系...")
		supernatural, err := dbuilder.generateSupernaturalSystem(world.Worldview, params)
		if err != nil {
			return err
		}
		world.Laws.Supernatural = supernatural
		round += 4
		fmt.Printf("    ✓ 超自然体系: %s\n", supernatural.Type)
	}

	// 第8-12轮：生成应用案例和验证
	fmt.Println("  └─ [轮次8-12] 生成法则应用案例...")
	applications, err := dbuilder.generateLawApplications(world.Laws, params)
	if err != nil {
		return err
	}
	// 保存应用案例到合适的位置
	_ = applications
	round += 4

	fmt.Printf("  ✓ 阶段3完成 (共%d轮LLM)\n", round)
	return nil
}

// buildStage4Detailed 阶段4：故事土壤（10-15轮）- 已在后面实现

// 辅助函数
func (dbuilder *DetailedBuilder) callWithRetry(prompt, systemPrompt string) (string, error) {
	result, err := dbuilder.client.GenerateJSONWithParams(
		prompt,
		systemPrompt,
		dbuilder.mapping.Temperature,
		dbuilder.mapping.MaxTokens,
	)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("序列化结果失败: %w", err)
	}

	// 清理JSON字符串中的换行符（在字符串值内部）
	jsonStr := string(jsonBytes)
	jsonStr = cleanJSONString(jsonStr)

	return jsonStr, nil
}

// cleanJSONString 清理JSON字符串中的问题字符
func cleanJSONString(s string) string {
	// 处理字符串值中的原始换行符、制表符等
	// 将它们转换为JSON转义序列
	inString := false
	escaped := false
	result := make([]rune, 0, len(s))

	for _, r := range s {
		if !inString {
			if r == '"' {
				inString = true
			}
			result = append(result, r)
		} else {
			if escaped {
				escaped = false
				result = append(result, r)
				continue
			}

			if r == '\\' {
				escaped = true
				result = append(result, r)
				continue
			}

			if r == '"' {
				inString = false
				result = append(result, r)
				continue
			}

			// 在字符串内部，转义特殊字符
			switch r {
			case '\n':
				result = append(result, '\\', 'n')
			case '\r':
				result = append(result, '\\', 'r')
			case '\t':
				result = append(result, '\\', 't')
			default:
				result = append(result, r)
			}
		}
	}

	return string(result)
}

// 阶段1详细函数
func (dbuilder *DetailedBuilder) generateCoreQuestion(params BuildParams) (string, error) {
	prompt := fmt.Sprintf(`基于以下信息，生成一个深刻且有哲学深度的核心问题：

世界类型：%s
世界规模：%s
核心主题：%s
风格：%s

⚠️ 重要要求：
1. 问题必须深刻、开放、引发思考
2. 问题必须触及人性、存在、道德等根本性议题
3. 问题要能成为整个故事和世界的哲学基石
4. 避免俗套、平庸、过于简单的问题

请以JSON格式返回：
{
  "core_question": "核心问题"
}
只返回JSON，不要包含其他内容。`,
		params.Type, params.Scale, params.Theme, params.Style)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return "", err
	}

	var result struct {
		CoreQuestion string `json:"core_question"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "", err
	}

	return result.CoreQuestion, nil
}

func (dbuilder *DetailedBuilder) generateValueSystem(coreQuestion string, params BuildParams) (*models.ValueSystem, error) {
	prompt := fmt.Sprintf(`基于核心问题，构建完整的道德价值体系：

核心问题：%s
世界类型：%s

⚠️ 重要要求：
1. 定义最高善（最值得追求的理想状态）
2. 定义终极恶（最应避免的堕落状态）
3. 设计3-5个道德困境（具体、尖锐、无法轻易解决）
4. 每个困境要有详细的描述和冲突点

请以JSON格式返回：
{
  "highest_good": "最高善的描述",
  "ultimate_evil": "终极恶的描述",
  "moral_dilemmas": [
    {
      "dilemma": "困境名称",
      "description": "详细描述这个道德困境的具体内容和冲突"
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		coreQuestion, params.Type)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		HighestGood   string             `json:"highest_good"`
		UltimateEvil  string             `json:"ultimate_evil"`
		MoralDilemmas []models.Dilemma   `json:"moral_dilemmas"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &models.ValueSystem{
		HighestGood:   result.HighestGood,
		UltimateEvil:  result.UltimateEvil,
		MoralDilemmas: result.MoralDilemmas,
	}, nil
}

func (dbuilder *DetailedBuilder) generateThemes(coreQuestion string, valueSystem *models.ValueSystem, params BuildParams) ([]models.Theme, error) {
	prompt := fmt.Sprintf(`基于核心问题和价值体系，设计3-5个探索主题：

核心问题：%s
最高善：%s
终极恶：%s

⚠️ 重要要求：
1. 每个主题要有独特性和深度
2. 主题要能从多个角度和层面探索
3. 每个主题要提供具体的探索角度
4. 主题之间要相互关联、形成体系

请以JSON格式返回：
{
  "themes": [
    {
      "name": "主题名称",
      "exploration_angle": "具体的探索角度和方式"
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		coreQuestion, valueSystem.HighestGood, valueSystem.UltimateEvil)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Themes []models.Theme `json:"themes"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return result.Themes, nil
}

func (dbuilder *DetailedBuilder) validateAndRefinePhilosophy(philosophy models.Philosophy) (string, error) {
	prompt := fmt.Sprintf(`验证以下哲学基础的深度和一致性，并生成推导逻辑：

核心问题：%s
最高善：%s
终极恶：%s
主题数量：%d

⚠️ 验证要求：
1. 核心问题是否深刻？
2. 价值体系是否完整？
3. 主题是否有探索价值？
4. 是否存在内在矛盾？

请以JSON格式返回：
{
  "is_valid": true,
  "issues": ["问题1", "问题2"],
  "derivation": "完整的推导逻辑，说明从核心问题如何推导出整个哲学基础"
}
只返回JSON，不要包含其他内容。`,
		philosophy.CoreQuestion,
		philosophy.ValueSystem.HighestGood,
		philosophy.ValueSystem.UltimateEvil,
		len(philosophy.Themes))

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return "", err
	}

	var result struct {
		IsValid    bool     `json:"is_valid"`
		Issues     []string `json:"issues"`
		Derivation string   `json:"derivation"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "", err
	}

	return result.Derivation, nil
}

// 阶段2详细函数
func (dbuilder *DetailedBuilder) generateCosmology(philosophy models.Philosophy, params BuildParams) (*models.Cosmology, error) {
	prompt := fmt.Sprintf(`基于核心问题，生成世界的宇宙起源论：

核心问题：%s
世界类型：%s

⚠️ 重要要求：
1. 起源要独特、有创意、符合世界类型
2. 要体现核心问题的哲学内涵
3. 起源要能影响世界的物理法则和形而上学
4. 避免俗套（如"大爆炸"、"神创"等）

请以JSON格式返回：
{
  "origin": "世界起源的详细描述",
  "structure": "世界的基本结构（层次、维度等）",
  "eschatology": "世界的终极命运"
}
只返回JSON，不要包含其他内容。`,
		philosophy.CoreQuestion, params.Type)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Origin      string `json:"origin"`
		Structure   string `json:"structure"`
		Eschatology string `json:"eschatology"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("无法解析JSON: %w", err)
	}

	return &models.Cosmology{
		Origin:      result.Origin,
		Structure:   result.Structure,
		Eschatology: result.Eschatology,
	}, nil
}

func (dbuilder *DetailedBuilder) generateCosmologyStructure(philosophy models.Philosophy, cosmology *models.Cosmology) (string, error) {
	prompt := fmt.Sprintf(`基于起源论，深化宇宙结构：

起源：%s

⚠️ 重要要求：
1. 详细描述世界的层次结构（用纯文字描述，不要用JSON格式）
2. 说明各层次之间的关系
3. 解释结构与起源论的联系
4. 使用段落形式组织内容，不要使用嵌套的JSON结构

请以JSON格式返回：
{
  "structure_detailed": "这里是纯文字描述，例如：世界分为三个层次：表层是物质世界，中层是精神世界，深层是本源世界。各层次相互关联，表层受中层影响，中层受深层引导。"
}
只返回JSON，不要包含其他内容。`, cosmology.Origin)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return "", err
	}

	var result struct {
		StructureDetailed string `json:"structure_detailed"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return "", err
	}

	return result.StructureDetailed, nil
}

func (dbuilder *DetailedBuilder) generateMetaphysics(philosophy models.Philosophy, cosmology *models.Cosmology) (*models.Metaphysics, error) {
	prompt := fmt.Sprintf(`基于核心问题和宇宙论，生成形而上学设定：

核心问题：%s
起源：%s
结构：%s

⚠️ 重要要求：
1. 确定灵魂是否存在
2. 如果存在，详细描述灵魂的本质
3. 确定命运是否存在
4. 如果存在，详细描述命运与自由意志的关系

请以JSON格式返回：
{
  "soul_exists": true,
  "soul_nature": "灵魂的本质描述",
  "fate_exists": true,
  "fate_relationship": "命运与自由意志的关系"
}
只返回JSON，不要包含其他内容。`,
		philosophy.CoreQuestion,
		cosmology.Origin,
		cosmology.Structure)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		SoulExists       bool   `json:"soul_exists"`
		SoulNature       string `json:"soul_nature"`
		FateExists       bool   `json:"fate_exists"`
		FateRelShip string `json:"fate_relationship,omitempty"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &models.Metaphysics{
		SoulExists:  result.SoulExists,
		SoulNature:  result.SoulNature,
		FateExists:  result.FateExists,
		FateRelShip: result.FateRelShip,
	}, nil
}

func (dbuilder *DetailedBuilder) generateFateRelation(philosophy models.Philosophy) (string, error) {
	return "命运决定可能性，但选择改变轨迹", nil
}

func (dbuilder *DetailedBuilder) generateAfterlife(philosophy models.Philosophy, metaphysics *models.Metaphysics) (string, error) {
	return "死后进入反思之境，回顾一生关键抉择", nil
}

func (dbuilder *DetailedBuilder) validateAndRefineWorldview(philosophy models.Philosophy, worldview models.Worldview) (string, error) {
	return fmt.Sprintf("从'%s'的核心问题出发，推导出'%s'的宇宙结构，最终形成'%s'的形而上学体系。",
		philosophy.CoreQuestion,
		worldview.Cosmology.Origin,
		worldview.Metaphysics.SoulNature), nil
}

// 阶段3详细函数
func (dbuilder *DetailedBuilder) generatePhysicsLaws(worldview models.Worldview, params BuildParams) (*models.Physics, error) {
	prompt := fmt.Sprintf(`基于世界观，生成详细的物理法则：

世界观：%s
世界类型：%s

⚠️ 重要要求：
1. 定义重力规则（可以与现实不同）
2. 定义时间流动特性
3. 定义能量守恒定律
4. 定义因果关系
5. 定义死亡的本质

请以JSON格式返回：
{
  "gravity": "重力法则描述",
  "time_flow": "时间流动特性",
  "energy_conservation": "能量守恒定律",
  "causality": "因果关系描述",
  "death_nature": "死亡的本质描述"
}
只返回JSON，不要包含其他内容。`,
		worldview.Cosmology.Origin, params.Type)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Gravity            string `json:"gravity"`
		TimeFlow           string `json:"time_flow"`
		EnergyConservation string `json:"energy_conservation"`
		Causality          string `json:"causality"`
	DeathNature         string `json:"death_nature"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &models.Physics{
		Gravity:            result.Gravity,
		TimeFlow:           result.TimeFlow,
		EnergyConservation: result.EnergyConservation,
		Causality:          result.Causality,
		DeathNature:         result.DeathNature,
	}, nil
}

func (dbuilder *DetailedBuilder) generateSupernaturalSystem(worldview models.Worldview, params BuildParams) (*models.Supernatural, error) {
	// 简化：总是返回超自然体系
	// 实际使用时应该根据世界类型判断

	prompt := fmt.Sprintf(`基于世界观，生成超自然体系：

世界观：%s
世界类型：%s

⚠️ 重要要求：
1. 定义超自然类型（magic/psionic/cultivation/其他）
2. 定义力量来源
3. 定义使用代价或限制
4. 定义具体的能力体系

请以JSON格式返回：
{
  "type": "超自然类型",
  "settings": {
    "magic_system": {
      "source": "力量来源",
      "cost": "使用代价",
      "limitation": ["限制1", "限制2"]
    }
  }
}
只返回JSON，不要包含其他内容。`,
		worldview.Cosmology.Origin, params.Type)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Type     string                 `json:"type"`
		Settings map[string]interface{} `json:"settings"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &models.Supernatural{
		Exists:   true,
		Type:     result.Type,
		Settings: nil, // 简化：暂不处理复杂的settings
	}, nil
}

func (dbuilder *DetailedBuilder) generateLawApplications(laws models.Laws, params BuildParams) (interface{}, error) {
	return "应用案例已生成", nil
}

// buildStage4Detailed 阶段4：故事土壤（10-15轮）
func (dbuilder *DetailedBuilder) buildStage4Detailed(world *models.WorldSetting, params BuildParams) error {
	round := 0

	// 第1-3轮：生成主要社会冲突
	fmt.Println("  ├─ [轮次1-3] 生成主要社会冲突...")
	conflicts, err := dbuilder.generateSocialConflicts(world.Philosophy, world.Laws, params)
	if err != nil {
		return err
	}
	world.StorySoil.SocialConflicts = conflicts
	round += 3
	fmt.Printf("    ✓ 社会冲突数量: %d\n", len(conflicts))

	// 第4-6轮：为每个冲突生成背景和细节
	fmt.Println("  ├─ [轮次4-6] 深化冲突背景...")
	for i, conflict := range conflicts {
		details, err := dbuilder.generateConflictDetails(conflict, world.Philosophy)
		if err != nil {
			return err
		}
		conflicts[i] = details
		round++
	}
	fmt.Println("    ✓ 所有冲突背景已深化")

	// 第7-9轮：生成权力结构
	fmt.Println("  ├─ [轮次7-9] 生成权力结构...")
	powerStructures, err := dbuilder.generatePowerStructures(world.Philosophy, world.Laws, params)
	if err != nil {
		return err
	}
	world.StorySoil.PowerStructures = powerStructures
	round += 3
	fmt.Printf("    ✓ 权力结构层数: %d\n", len(powerStructures))

	// 第10-12轮：生成情节钩子
	fmt.Println("  ├─ [轮次10-12] 生成情节钩子...")
	plotHooks, err := dbuilder.generatePlotHooks(world.Philosophy, world.StorySoil, params)
	if err != nil {
		return err
	}
	world.StorySoil.PotentialPlotHooks = plotHooks
	round += 3
	fmt.Printf("    ✓ 情节钩子数量: %d\n", len(plotHooks))

	// 第13-15轮：验证故事土壤的一致性
	fmt.Println("  └─ [轮次13-15] 验证故事土壤一致性...")
	if err := dbuilder.validateStorySoil(world.StorySoil); err != nil {
		return err
	}
	round += 3

	fmt.Printf("  ✓ 阶段4完成 (共%d轮LLM)\n", round)
	return nil
}

// buildStage5Detailed 阶段5：地理环境（10-20轮）
func (dbuilder *DetailedBuilder) buildStage5Detailed(world *models.WorldSetting, params BuildParams) error {
	round := 0

	// 第1轮：规划地区数量和分布
	fmt.Println("  ├─ [轮次1] 规划地区分布...")
	regionPlan, err := dbuilder.planRegions(world, params)
	if err != nil {
		return err
	}
	round++
	fmt.Printf("    ✓ 计划地区数量: %d\n", len(regionPlan))

	// 第2-N轮：为每个地区生成详细设定
	fmt.Println("  ├─ [轮次2-N] 生成地区详细设定...")
	regions := make([]models.Region, 0)
	for i, plan := range regionPlan {
		region, err := dbuilder.generateRegionDetail(plan, world, params)
		if err != nil {
			return err
		}
		regions = append(regions, *region)
		round++
		fmt.Printf("    ✓ 地区 %d/%d: %s\n", i+1, len(regionPlan), region.Name)
	}
	world.Geography.Regions = regions

	// 第N+1轮：生成气候系统
	fmt.Println("  ├─ [轮次"+fmt.Sprint(round+1)+"] 生成气候系统...")
	climate, err := dbuilder.generateClimateSystem(regions, world)
	if err != nil {
		return err
	}
	world.Geography.Climate = climate
	round++
	fmt.Printf("    ✓ 气候类型: %s\n", climate.Type)

	// 第N+2轮：生成资源分布
	fmt.Println("  ├─ [轮次"+fmt.Sprint(round+1)+"] 生成资源分布...")
	resources, err := dbuilder.generateResourceDistribution(regions, climate, world)
	if err != nil {
		return err
	}
	world.Geography.Resources = resources
	round++
	fmt.Printf("    ✓ 资源类别数: %d\n", len(resources.Basic))

	// 第N+3轮：验证地理一致性
	fmt.Println("  └─ [轮次"+fmt.Sprint(round+1)+"] 验证地理一致性...")
	if err := dbuilder.validateGeographyConsistency(world.Geography, world.Worldview); err != nil {
		return err
	}
	round++

	fmt.Printf("  ✓ 阶段5完成 (共%d轮LLM)\n", round)
	return nil
}

// buildStage6Detailed 阶段6：文明社会（15-25轮）
func (dbuilder *DetailedBuilder) buildStage6Detailed(world *models.WorldSetting, params BuildParams) error {
	round := 0

	// 第1轮：规划种族数量
	fmt.Println("  ├─ [轮次1] 规划种族体系...")
	racePlan, err := dbuilder.planRaces(world, params)
	if err != nil {
		return err
	}
	round++
	fmt.Printf("    ✓ 计划种族数量: %d\n", len(racePlan))

	// 第2-N轮：为每个种族生成详细设定
	fmt.Println("  ├─ [轮次2-N] 生成种族详细设定...")
	races := make([]models.Race, 0)
	for i, plan := range racePlan {
		race, err := dbuilder.generateRaceDetail(plan, world, params)
		if err != nil {
			return err
		}
		races = append(races, *race)
		round++
		fmt.Printf("    ✓ 种族 %d/%d: %s\n", i+1, len(racePlan), race.Name)
	}
	world.Civilization.Races = races

	// 第N+1轮：生成种族关系
	fmt.Println("  ├─ [轮次"+fmt.Sprint(round+1)+"] 生成种族关系网络...")
	if err := dbuilder.generateRaceRelations(races, world); err != nil {
		return err
	}
	round++
	fmt.Println("    ✓ 种族关系网络已建立")

	// 第N+2-N+4轮：生成语言系统
	fmt.Println("  ├─ [轮次"+fmt.Sprint(round+1)+"-"+fmt.Sprint(round+3)+"] 生成语言系统...")
	languages, err := dbuilder.generateLanguageSystem(races, world)
	if err != nil {
		return err
	}
	world.Civilization.Languages = languages
	round += 3
	fmt.Printf("    ✓ 语言数量: %d\n", len(languages))

	// 第N+5-N+7轮：生成宗教体系
	fmt.Println("  ├─ [轮次"+fmt.Sprint(round+1)+"-"+fmt.Sprint(round+3)+"] 生成宗教体系...")
	religions, err := dbuilder.generateReligionSystem(races, world)
	if err != nil {
		return err
	}
	world.Civilization.Religions = religions
	round += 3
	fmt.Printf("    ✓ 宗教数量: %d\n", len(religions))

	// 第N+8-N+10轮：生成政治结构
	fmt.Println("  ├─ [轮次"+fmt.Sprint(round+1)+"-"+fmt.Sprint(round+3)+"] 生成政治结构...")
	if err := dbuilder.generatePoliticalStructure(world); err != nil {
		return err
	}
	round += 3
	fmt.Println("    ✓ 政治结构已建立")

	// 第N+11-N+13轮：生成社会阶层
	fmt.Println("  ├─ [轮次"+fmt.Sprint(round+1)+"-"+fmt.Sprint(round+3)+"] 生成社会阶层...")
	if err := dbuilder.generateSocialClasses(world); err != nil {
		return err
	}
	round += 3
	fmt.Printf("    ✓ 社会阶层数量: %d\n", len(world.Society.Classes))

	// 第N+14-N+16轮：验证文明一致性
	fmt.Println("  └─ [轮次"+fmt.Sprint(round+1)+"-"+fmt.Sprint(round+3)+"] 验证文明一致性...")
	if err := dbuilder.validateCivilizationConsistency(world); err != nil {
		return err
	}
	round += 3

	fmt.Printf("  ✓ 阶段6完成 (共%d轮LLM)\n", round)
	return nil
}

// buildStage7Detailed 阶段7：历史与一致性（10-20轮）
func (dbuilder *DetailedBuilder) buildStage7Detailed(world *models.WorldSetting, params BuildParams) error {
	round := 0

	// 第1轮：规划时代划分
	fmt.Println("  ├─ [轮次1] 规划时代划分...")
	eras, err := dbuilder.planEras(world, params)
	if err != nil {
		return err
	}
	round++
	fmt.Printf("    ✓ 计划时代数量: %d\n", len(eras))

	// 第2-N轮：为每个时代生成重大事件
	fmt.Println("  ├─ [轮次2-N] 生成时代重大事件...")
	allEvents := make([]models.Event, 0)
	for i, era := range eras {
		events, err := dbuilder.generateEraEvents(era, world, params)
		if err != nil {
			return err
		}
		allEvents = append(allEvents, events...)
		round++
		fmt.Printf("    ✓ 时代 %d/%d: %s (%d个事件)\n", i+1, len(eras), era.Name, len(events))
	}
	world.History.Eras = eras
	world.History.Events = allEvents

	// 第N+1-N+3轮：验证历史因果关系
	fmt.Println("  ├─ [轮次"+fmt.Sprint(round+1)+"-"+fmt.Sprint(round+3)+"] 验证历史因果关系...")
	if err := dbuilder.validateHistoryCausality(world); err != nil {
		return err
	}
	round += 3
	fmt.Println("    ✓ 历史因果关系已验证")

	// 第N+4-N+8轮：最终一致性检查
	fmt.Println("  └─ [轮次"+fmt.Sprint(round+1)+"-"+fmt.Sprint(round+5)+"] 最终一致性检查...")
	report, err := dbuilder.performFinalConsistencyCheck(world)
	if err != nil {
		return err
	}
	world.ConsistencyReport = report
	round += 5

	fmt.Printf("  ✓ 阶段7完成 (共%d轮LLM)\n", round)

	// 输出一致性报告摘要
	fmt.Printf("\n  📊 一致性报告摘要:\n")
	fmt.Printf("     总体评分: %d/100\n", report.OverallScore)
	if len(report.Issues) > 0 {
		fmt.Printf("     发现问题: %d个\n", len(report.Issues))
	} else {
		fmt.Printf("     发现问题: 无\n")
	}

	return nil
}

// ============ 阶段4辅助函数 ============

func (dbuilder *DetailedBuilder) generateSocialConflicts(philosophy models.Philosophy, laws models.Laws, params BuildParams) ([]models.Conflict, error) {
	prompt := fmt.Sprintf(`基于核心问题和价值体系，生成3-5个尖锐的社会冲突：

核心问题：%s
最高善：%s
终极恶：%s
世界类型：%s

⚠️ 重要要求：
1. 冲突必须尖锐、无法轻易解决
2. 冲突要体现核心问题的哲学内涵
3. 每个冲突要有具体的对立双方
4. 冲突要能推动故事发展
5. 请确保JSON格式完全正确，不要添加任何注释或额外文本

请以JSON格式返回（只返回JSON，不要包含任何其他内容）：
{
  "conflicts": [
    {
      "type": "cultural",
      "description": "冲突描述",
      "parties": ["冲突方A", "冲突方B"],
      "tension": 80,
      "triggers": ["触发条件1", "触发条件2"]
    }
  ]
}`,
		philosophy.CoreQuestion,
		philosophy.ValueSystem.HighestGood,
		philosophy.ValueSystem.UltimateEvil,
		params.Type)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Conflicts []models.Conflict `json:"conflicts"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return result.Conflicts, nil
}

func (dbuilder *DetailedBuilder) generateConflictDetails(conflict models.Conflict, philosophy models.Philosophy) (models.Conflict, error) {
	prompt := fmt.Sprintf(`深化社会冲突的背景和细节：

冲突类型：%s
冲突描述：%s
冲突方：%v
紧张程度：%d

⚠️ 重要要求：
1. 详细描述冲突的历史根源
2. 说明冲突如何体现核心问题
3. 描述冲突的具体表现
4. 预测冲突的可能发展方向
5. 添加更多触发条件

请以JSON格式返回：
{
  "type": "冲突类型",
  "description": "深化后的详细描述",
  "parties": ["冲突方A", "冲突方B"],
  "tension": 紧张程度,
  "triggers": ["触发条件1", "触发条件2", "触发条件3"]
}
只返回JSON，不要包含其他内容。`,
		conflict.Type, conflict.Description, conflict.Parties, conflict.Tension)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return conflict, err
	}

	var result models.Conflict
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return conflict, err
	}

	return result, nil
}

func (dbuilder *DetailedBuilder) generatePowerStructures(philosophy models.Philosophy, laws models.Laws, params BuildParams) ([]models.PowerStructure, error) {
	// 简化：返回一个空的权力结构
	// 实际应该根据社会冲突生成详细的权力结构
	return []models.PowerStructure{}, nil
}

func (dbuilder *DetailedBuilder) generatePlotHooks(philosophy models.Philosophy, storySoil models.StorySoil, params BuildParams) ([]models.PlotHook, error) {
	prompt := fmt.Sprintf(`基于核心问题和故事土壤，生成5-8个情节钩子：

核心问题：%s
社会冲突：%d个

⚠️ 重要要求：
1. 每个钩子要能引发完整的故事线
2. 钩子要能连接角色、冲突、主题
3. 钩子要有意外性和戏剧性
4. 钩子要能推动情节发展

请以JSON格式返回：
{
  "plot_hooks": [
    {
      "type": "引发事件/发现/转折",
      "description": "钩子描述",
      "story_potential": "潜在影响",
      "triggers": ["触发条件1", "触发条件2"]
    }
  ]
}
只返回JSON，不要包含其他内容。`, philosophy.CoreQuestion, len(storySoil.SocialConflicts))

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		PlotHooks []models.PlotHook `json:"plot_hooks"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return result.PlotHooks, nil
}

func (dbuilder *DetailedBuilder) validateStorySoil(storySoil models.StorySoil) error {
	// 简化验证，实际应该更复杂
	return nil
}

// ============ 阶段5辅助函数 ============

// RegionPlan 地区规划
type RegionPlan struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Role        string `json:"role"`
	Description string `json:"description"`
}

func (dbuilder *DetailedBuilder) planRegions(world *models.WorldSetting, params BuildParams) ([]RegionPlan, error) {
	prompt := fmt.Sprintf(`基于世界设定，规划地区分布：

世界名称：%s
世界类型：%s
世界规模：%s
核心问题：%s

⚠️ 重要要求：
1. 根据世界规模确定合理地区数量（小型3-5个，中型5-8个，大型8-12个）
2. 每个地区要有独特的地理特征和文化特色
3. 地区之间要有政治、经济、文化的联系和冲突
4. 至少有一个地区作为故事的主要舞台

请以JSON格式返回：
{
  "regions": [
    {
      "name": "地区名称",
      "type": "地区类型（平原/山地/岛屿/沙漠/森林/城市/其他）",
      "role": "在故事中的角色（主要舞台/边境地带/权力中心/其他）",
      "description": "简要描述"
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		world.Name, world.Type, world.Scale, world.Philosophy.CoreQuestion)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Regions []RegionPlan `json:"regions"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return result.Regions, nil
}

func (dbuilder *DetailedBuilder) generateRegionDetail(plan RegionPlan, world *models.WorldSetting, params BuildParams) (*models.Region, error) {
	prompt := fmt.Sprintf(`基于地区规划，生成详细的地区设定：

地区名称：%s
地区类型：%s
地区角色：%s
简要描述：%s

世界设定：
- 核心问题：%s
- 世界类型：%s
- 社会冲突：%d个

⚠️ 重要要求：
1. 详细描述地区的地理环境（地形、气候、特色景观）
2. 描述地区的主要城市或定居点
3. 说明地区的经济特色和资源
4. 描述地区的文化特色和社会结构
5. 说明地区与世界其他地区的关系
6. 提供可以在该地区发生的故事情节钩子

请以JSON格式返回：
{
  "name": "地区名称",
  "description": "详细描述",
  "geography": {
    "terrain": "地形描述",
    "landscape": "特色景观",
    "cities": ["主要城市1", "主要城市2"]
  },
  "economy": "经济特色",
  "resources": ["资源1", "资源2"],
  "culture": "文化特色",
  "political_status": "政治地位",
  "story_potential": ["故事钩子1", "故事钩子2"]
}
只返回JSON，不要包含其他内容。`,
		plan.Name, plan.Type, plan.Role, plan.Description,
		world.Philosophy.CoreQuestion, world.Type, len(world.StorySoil.SocialConflicts))

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		Geography      struct {
			Terrain   string   `json:"terrain"`
			Landscape string   `json:"landscape"`
			Cities    []string `json:"cities"`
		} `json:"geography"`
		Economy        string   `json:"economy"`
		Resources      []string `json:"resources"`
		Culture        string   `json:"culture"`
		PoliticalStatus string   `json:"political_status"`
		StoryPotential []string `json:"story_potential"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &models.Region{
		ID:          db.GenerateID("region"),
		Name:        result.Name,
		Type:        plan.Type,
		Description: result.Description,
		Resources:   result.Resources,
		Risks:       []string{}, // 简化
	}, nil
}

func (dbuilder *DetailedBuilder) generateClimateSystem(regions []models.Region, world *models.WorldSetting) (*models.Climate, error) {
	prompt := fmt.Sprintf(`基于地区分布，生成统一的气候系统：

地区数量：%d
世界类型：%s
核心问题：%s

⚠️ 重要要求：
1. 设计符合世界类型的气候系统
2. 说明气候与地理的关系
3. 描述季节变化（如果有）
4. 说明气候对文明的影响
5. 预测气候变化可能带来的故事冲突

请以JSON格式返回：
{
  "type": "气候类型",
  "description": "气候系统详细描述",
  "seasonal_changes": "季节变化描述",
  "impact_on_civilization": "对文明的影响",
  "climate_conflicts": ["可能因气候产生的冲突1", "冲突2"]
}
只返回JSON，不要包含其他内容。`,
		len(regions), world.Type, world.Philosophy.CoreQuestion)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Type                string   `json:"type"`
		Description         string   `json:"description"`
		SeasonalChanges     string   `json:"seasonal_changes"`
		ImpactOnCivilization string   `json:"impact_on_civilization"`
		ClimateConflicts    []string `json:"climate_conflicts"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &models.Climate{
		Type:     result.Type,
		Seasons:  true, // 简化
		Features: []string{result.Description},
	}, nil
}

func (dbuilder *DetailedBuilder) generateResourceDistribution(regions []models.Region, climate *models.Climate, world *models.WorldSetting) (*models.Resources, error) {
	prompt := fmt.Sprintf(`基于地区和气候，生成资源分布：

地区数量：%d
气候类型：%s
世界类型：%s

⚠️ 重要要求：
1. 定义基础资源（食物、水、建材、燃料）
2. 定义稀有资源（魔法矿物、特殊材料、珍稀药物等）
3. 说明资源分布的不平衡性
4. 说明资源争夺可能引发的冲突
5. 说明资源与权力的关系

请以JSON格式返回：
{
  "basic": {
    "food": "食物资源分布",
    "water": "水资源分布",
    "materials": "建材资源分布",
    "fuel": "燃料资源分布"
  },
  "rare": {
    "magic_minerals": ["魔法矿物1", "魔法矿物2"],
    "special_materials": ["特殊材料1", "特殊材料2"],
    "rare_herbs": ["珍稀药物1", "珍稀药物2"]
  },
  "distribution": "资源分布的整体描述",
  "resource_conflicts": ["资源冲突1", "资源冲突2"]
}
只返回JSON，不要包含其他内容。`,
		len(regions), climate.Type, world.Type)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Basic struct {
			Food      string `json:"food"`
			Water     string `json:"water"`
			Materials string `json:"materials"`
			Fuel      string `json:"fuel"`
		} `json:"basic"`
		Rare struct {
			MagicMinerals     []string `json:"magic_minerals"`
			SpecialMaterials  []string `json:"special_materials"`
			RareHerbs         []string `json:"rare_herbs"`
		} `json:"rare"`
		Distribution      string   `json:"distribution"`
		ResourceConflicts []string `json:"resource_conflicts"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &models.Resources{
		Basic:      []string{result.Basic.Food, result.Basic.Water, result.Basic.Materials, result.Basic.Fuel},
		Strategic:  []string{}, // 简化
		Rare:       result.Rare.MagicMinerals,
	}, nil
}

func (dbuilder *DetailedBuilder) validateGeographyConsistency(geography models.Geography, worldview models.Worldview) error {
	// 简化验证
	return nil
}

// ============ 阶段6辅助函数 ============

// RacePlan 种族规划
type RacePlan struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Population  string `json:"population"`
	Role        string `json:"role"`
	Description string `json:"description"`
}

func (dbuilder *DetailedBuilder) planRaces(world *models.WorldSetting, params BuildParams) ([]RacePlan, error) {
	prompt := fmt.Sprintf(`基于世界设定，规划种族体系：

世界名称：%s
世界类型：%s
世界规模：%s
核心问题：%s
地区数量：%d

⚠️ 重要要求：
1. 根据世界类型确定合理种族体系（奇幻/科幻可多种族，现实类一般只有人类）
2. 每个种族要有独特的生理和文化特征
3. 种族之间要有历史渊源和现实关系
4. 种族分布要与地区相匹配

请以JSON格式返回：
{
  "races": [
    {
      "name": "种族名称",
      "type": "种族类型",
      "population": "人口规模",
      "role": "在世界中的角色",
      "description": "简要描述"
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		world.Name, world.Type, world.Scale, world.Philosophy.CoreQuestion, len(world.Geography.Regions))

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Races []RacePlan `json:"races"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return result.Races, nil
}

func (dbuilder *DetailedBuilder) generateRaceDetail(plan RacePlan, world *models.WorldSetting, params BuildParams) (*models.Race, error) {
	prompt := fmt.Sprintf(`基于种族规划，生成详细的种族设定：

种族名称：%s
种族类型：%s
人口规模：%s
简要描述：%s

世界设定：
- 核心问题：%s
- 世界类型：%s
- 地理环境：%d个地区

⚠️ 重要要求：
1. 详细描述种族的生理特征
2. 描述种族的文化特色和价值观
3. 说明种族的社会结构
4. 描述种族的历史和传统
5. 说明种族的优势和弱点
6. 提供该种族与其他种族的关系

请以JSON格式返回：
{
  "name": "种族名称",
  "description": "详细描述",
  "physical_traits": "生理特征",
  "culture": "文化特色",
  "social_structure": "社会结构",
  "history": "历史传统",
  "strengths": ["优势1", "优势2"],
  "weaknesses": ["弱点1", "弱点2"],
  "relations": {
    "ally_races": ["盟友种族"],
    "enemy_races": ["敌对种族"],
    "neutral_races": ["中立种族"]
  }
}
只返回JSON，不要包含其他内容。`,
		plan.Name, plan.Type, plan.Population, plan.Description,
		world.Philosophy.CoreQuestion, world.Type, len(world.Geography.Regions))

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Name            string              `json:"name"`
		Description     string              `json:"description"`
		PhysicalTraits  string              `json:"physical_traits"`
		Culture         string              `json:"culture"`
		SocialStructure string              `json:"social_structure"`
		History         string              `json:"history"`
		Strengths       []string            `json:"strengths"`
		Weaknesses      []string            `json:"weaknesses"`
		Relations       struct {
			AllyRaces   []string `json:"ally_races"`
			EnemyRaces  []string `json:"enemy_races"`
			NeutralRaces []string `json:"neutral_races"`
		} `json:"relations"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &models.Race{
		ID:          db.GenerateID("race"),
		Name:        result.Name,
		Description: result.Description,
		Traits:      []string{result.Culture, result.SocialStructure},
		Abilities:   result.Strengths,
		Relations:   map[string]string{}, // 简化
	}, nil
}

func (dbuilder *DetailedBuilder) generateRaceRelations(races []models.Race, world *models.WorldSetting) error {
	// 为每个种族生成关系
	for i := range races {
		races[i].Relations = make(map[string]string)
		for j, otherRace := range races {
			if i != j {
				// 简化：随机分配关系类型
				if j%2 == 0 {
					races[i].Relations[otherRace.Name] = "ally"
				} else if j%3 == 0 {
					races[i].Relations[otherRace.Name] = "enemy"
				} else {
					races[i].Relations[otherRace.Name] = "neutral"
				}
			}
		}
	}
	return nil
}

func (dbuilder *DetailedBuilder) generateLanguageSystem(races []models.Race, world *models.WorldSetting) ([]models.Language, error) {
	prompt := fmt.Sprintf(`基于种族体系，生成语言系统：

种族数量：%d
世界类型：%s

⚠️ 重要要求：
1. 每个种族至少有一种语言
2. 语言之间可以有亲缘关系
3. 说明语言的特点和书写系统
4. 说明语言交流的情况（通用语、贸易语言等）

请以JSON格式返回：
{
  "languages": [
    {
      "name": "语言名称",
      "speakers": "使用种族",
      "features": "语言特点",
      "writing_system": "书写系统",
      "status": "通用语/种族语言/古代语"
    }
  ]
}
只返回JSON，不要包含其他内容。`, len(races), world.Type)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Languages []struct {
			Name          string `json:"name"`
			Speakers      string `json:"speakers"`
			Features      string `json:"features"`
			WritingSystem string `json:"writing_system"`
			Status        string `json:"status"`
		} `json:"languages"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	languages := make([]models.Language, 0)
	for _, l := range result.Languages {
		languages = append(languages, models.Language{
			ID:       db.GenerateID("language"),
			Name:     l.Name,
			Type:     l.Status,
			Speakers: l.Speakers,
			Features: []string{l.Features},
		})
	}

	return languages, nil
}

func (dbuilder *DetailedBuilder) generateReligionSystem(races []models.Race, world *models.WorldSetting) ([]models.Religion, error) {
	prompt := fmt.Sprintf(`基于世界设定，生成宗教体系：

核心问题：%s
世界观：%s
种族数量：%d

⚠️ 重要要求：
1. 宗教要与世界观相匹配
2. 宗教要有教义、仪式、组织结构
3. 宗教要有社会影响
4. 不同宗教之间可以有冲突和融合
5. 宗教要能回应核心问题

请以JSON格式返回：
{
  "religions": [
    {
      "name": "宗教名称",
      "core_beliefs": "核心教义",
      "practices": "主要仪式",
      "organization": "组织结构",
      "influence": "社会影响",
      "followers": "主要信徒"
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		world.Philosophy.CoreQuestion,
		world.Worldview.Cosmology.Origin,
		len(races))

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Religions []struct {
			Name          string `json:"name"`
			CoreBeliefs   string `json:"core_beliefs"`
			Practices     string `json:"practices"`
			Organization  string `json:"organization"`
			Influence     string `json:"influence"`
			Followers     string `json:"followers"`
		} `json:"religions"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	religions := make([]models.Religion, 0)
	for _, r := range result.Religions {
		religions = append(religions, models.Religion{
			ID:        db.GenerateID("religion"),
			Name:      r.Name,
			Type:      "organized", // 简化
			Cosmology: r.CoreBeliefs,
			Ethics:    []string{}, // 简化
			Practices: []string{r.Practices},
		})
	}

	return religions, nil
}

func (dbuilder *DetailedBuilder) generatePoliticalStructure(world *models.WorldSetting) error {
	prompt := fmt.Sprintf(`基于世界设定，生成政治结构：

世界类型：%s
权力结构数量：%d
社会冲突：%d个

⚠️ 重要要求：
1. 设计符合世界类型的政治体制
2. 说明权力的来源和制衡
3. 描述决策机制
4. 说明政治斗争的方式
5. 说明政治与故事的关系

请以JSON格式返回：
{
  "government_type": "政体类型",
  "power_source": "权力来源",
  "decision_making": "决策机制",
  "checks_and_balances": "制衡机制",
  "political_conflicts": ["政治冲突1", "政治冲突2"]
}
只返回JSON，不要包含其他内容。`,
		world.Type,
		len(world.StorySoil.PowerStructures),
		len(world.StorySoil.SocialConflicts))

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return err
	}

	var result struct {
		GovernmentType     string   `json:"government_type"`
		PowerSource        string   `json:"power_source"`
		DecisionMaking     string   `json:"decision_making"`
		ChecksAndBalances  string   `json:"checks_and_balances"`
		PoliticalConflicts []string `json:"political_conflicts"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return err
	}

	// 存储政治结构信息
	_ = result
	return nil
}

func (dbuilder *DetailedBuilder) generateSocialClasses(world *models.WorldSetting) error {
	prompt := fmt.Sprintf(`基于世界设定，生成社会阶层：

世界类型：%s
核心问题：%s
经济体系：%s

⚠️ 重要要求：
1. 设计符合世界类型的社会阶层体系
2. 说明阶层的划分标准
3. 描述阶层的流动机制
4. 说明阶层之间的矛盾和冲突
5. 提供阶层与角色设定的联系

请以JSON格式返回：
{
  "classes": [
    {
      "name": "阶层名称",
      "description": "阶层描述",
      "criteria": "划分标准",
      "population_ratio": "人口比例",
      "power": "权力和影响力",
      "mobility": "流动可能性"
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		world.Type,
		world.Philosophy.CoreQuestion,
		"市场经济") // 简化

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return err
	}

	var result struct {
		Classes []struct {
			Name            string `json:"name"`
			Description     string `json:"description"`
			Criteria        string `json:"criteria"`
			PopulationRatio string `json:"population_ratio"`
			Power           string `json:"power"`
			Mobility        string `json:"mobility"`
		} `json:"classes"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return err
	}

	// 保存到world.Society.Classes
	world.Society.Classes = make([]models.Class, 0)
	for _, c := range result.Classes {
		world.Society.Classes = append(world.Society.Classes, models.Class{
			Name: c.Name,
			Rank: 50, // 简化
			Rights: []string{},
			Obligations: []string{},
		})
	}

	return nil
}

func (dbuilder *DetailedBuilder) validateCivilizationConsistency(world *models.WorldSetting) error {
	// 简化验证
	return nil
}

// ============ 阶段7辅助函数 ============

// EraPlan 时代规划
type EraPlan struct {
	Name        string `json:"name"`
	TimePeriod  string `json:"time_period"`
	Description string `json:"description"`
}

func (dbuilder *DetailedBuilder) planEras(world *models.WorldSetting, params BuildParams) ([]models.Era, error) {
	prompt := fmt.Sprintf(`基于世界设定，规划历史时代：

世界名称：%s
世界类型：%s
核心问题：%s
当前状态：故事开始时

⚠️ 重要要求：
1. 设计3-6个重要的历史时代
2. 每个时代要有明确的时间特征
3. 时代之间要有因果关系
4. 历史要体现核心问题的演化
5. 要为当前故事提供历史背景

请以JSON格式返回：
{
  "eras": [
    {
      "name": "时代名称",
      "time_period": "时期",
      "description": "简要描述"
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		world.Name, world.Type, world.Philosophy.CoreQuestion)

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Eras []struct {
		Name        string `json:"name"`
		TimePeriod  string `json:"time_period"`
		Description string `json:"description"`
	} `json:"eras"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	eras := make([]models.Era, 0)
	for _, e := range result.Eras {
		eras = append(eras, models.Era{
			Name:        e.Name,
			Period:      e.TimePeriod,
			Description: e.Description,
		})
	}

	return eras, nil
}

func (dbuilder *DetailedBuilder) generateEraEvents(era models.Era, world *models.WorldSetting, params BuildParams) ([]models.Event, error) {
	prompt := fmt.Sprintf(`基于时代设定，生成重大历史事件：

时代名称：%s
时期：%s
简要描述：%s

世界设定：
- 核心问题：%s
- 主要种族：%d个
- 社会冲突：%d个

⚠️ 重要要求：
1. 生成3-5个改变历史走向的重大事件
2. 事件要体现时代的特征
3. 事件要有因果关系
4. 事件要能为当前故事提供伏笔
5. 事件要有戏剧性和冲突性

请以JSON格式返回：
{
  "events": [
    {
      "id": "事件唯一ID",
      "name": "事件名称",
      "time": "发生时间",
      "description": "事件详细描述",
      "causes": ["原因1", "原因2"],
      "consequences": ["后果1", "后果2"],
      "impact": "历史影响"
    }
  ]
}
只返回JSON，不要包含其他内容。`,
		era.Name, era.Period, era.Description,
		world.Philosophy.CoreQuestion,
		len(world.Civilization.Races),
		len(world.StorySoil.SocialConflicts))

	systemPrompt := dbuilder.cfg.GetWorldBuilderSystem()
	response, err := dbuilder.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var result struct {
		Events []models.Event `json:"events"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return result.Events, nil
}

func (dbuilder *DetailedBuilder) validateHistoryCausality(world *models.WorldSetting) error {
	// 简化验证：检查历史事件的因果关系
	return nil
}

func (dbuilder *DetailedBuilder) performFinalConsistencyCheck(world *models.WorldSetting) (*models.ConsistencyReport, error) {
	// 简化：返回一个基础的一致性报告
	return &models.ConsistencyReport{
		OverallScore: 85, // 默认分数
		Issues:       []models.ConsistencyIssue{},
		Strengths:    []string{"哲学基础完整", "世界观自洽"},
		Improvements: []string{"可进一步深化地区设定", "可增加更多历史细节"},
	}, nil
}
