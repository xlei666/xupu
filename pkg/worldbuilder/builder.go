// Package worldbuilder 世界设定器
// 负责7个阶段的世界构建：哲学、世界观、法则、地理、文明、历史、一致性检查
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

// BuildParams 世界构建参数
type BuildParams struct {
	// 基本参数
	Name  string            `json:"name"`  // 世界名称
	Type  models.WorldType  `json:"type"`  // 世界类型
	Scale models.WorldScale `json:"scale"` // 世界规模
	Style string            `json:"style"` // 风格倾向

	// 哲学参数（阶段1）
	Theme string `json:"theme"` // 核心主题
}

// Stage1Input 阶段1输入
type Stage1Input struct {
	WorldType string `json:"world_type"`
	Theme     string `json:"theme"`
	Style     string `json:"style"`
}

// Stage2Input 阶段2输入
type Stage2Input struct {
	CoreQuestion string `json:"core_question"`
	HighestGood  string `json:"highest_good"`
	UltimateEvil string `json:"ultimate_evil"`
}

// Stage3Input 阶段3输入
type Stage3Input struct {
	WorldType string `json:"world_type"`
	Worldview string `json:"worldview"`
}

// Stage4Input 阶段4输入
type Stage4Input struct {
	CoreQuestion  string `json:"core_question"`
	MainConflicts string `json:"main_conflicts"`
	WorldType     string `json:"world_type"`
}

// Stage5Input 阶段5输入
type Stage5Input struct {
	WorldType         string `json:"world_type"`
	WorldScale        string `json:"world_scale"`
	LawsSummary       string `json:"laws_summary"`
	CivilizationNeeds string `json:"civilization_needs"`
}

// Stage6Input 阶段6输入
type Stage6Input struct {
	WorldType        string `json:"world_type"`
	GeographySummary string `json:"geography_summary"`
	ValueSystem      string `json:"value_system"`
}

// Stage7Input 阶段7输入
type Stage7Input struct {
	WorldSettingSummary string `json:"world_setting_summary"`
}

// Stage5Output 阶段5输出（匹配LLM输出格式，带geography包装）
type Stage5Output struct {
	Geography struct {
		Regions []struct {
			ID          string   `json:"id"`
			Name        string   `json:"name"`
			Type        string   `json:"type"`
			Description string   `json:"description"`
			Resources   []string `json:"resources"`
			Risks       []string `json:"risks"`
		} `json:"regions"`
		Resources *struct {
			Basic     []string `json:"basic"`
			Strategic []string `json:"strategic"`
			Rare      []string `json:"rare"`
		} `json:"resources"`
		Climate *struct {
			Type     string   `json:"type"`
			Seasons  bool     `json:"seasons"`
			Features []string `json:"features"`
		} `json:"climate"`
	} `json:"geography"`
}

// Stage6Output 阶段6输出（匹配LLM输出格式，带civilization和society包装）
type Stage6Output struct {
	Civilization struct {
		Races []struct {
			ID          string            `json:"id"`
			Name        string            `json:"name"`
			Description string            `json:"description"`
			Traits      []string          `json:"traits"`
			Abilities   []string          `json:"abilities"`
			Relations   map[string]string `json:"relations"`
		} `json:"races"`
		Languages []struct {
			ID       string   `json:"id"`
			Name     string   `json:"name"`
			Type     string   `json:"type"`
			Speakers string   `json:"speakers"`
			Features []string `json:"features"`
		} `json:"languages"`
		Religions []struct {
			ID           string                       `json:"id"`
			Name         string                       `json:"name"`
			Type         string                       `json:"type"`
			Cosmology    string                       `json:"cosmology"`
			Ethics       []string                     `json:"ethics"`
			Practices    []string                     `json:"practices"`
			Organization *models.ReligionOrganization `json:"organization,omitempty"`
		} `json:"religions"`
	} `json:"civilization"`
	Society struct {
		Politics struct {
			Type             string `json:"type"`
			LegitimacySource string `json:"legitimacy_source"`
			PowerStructure   struct {
				Formal []struct {
					Level  string   `json:"level"`
					Name   string   `json:"name"`
					Powers []string `json:"powers"`
				} `json:"formal"`
				Actual []struct {
					Entity       string `json:"entity"`
					PowerSource  string `json:"power_source"`
					Relationship string `json:"relationship"`
				} `json:"actual"`
			} `json:"power_structure"`
		} `json:"politics"`
		Classes []struct {
			Name        string   `json:"name"`
			Rank        int      `json:"rank"`
			Rights      []string `json:"rights"`
			Obligations []string `json:"obligations"`
		} `json:"classes"`
		Economy struct {
			Type         string   `json:"type"`
			TradeNetwork string   `json:"trade_network"`
			Currency     []string `json:"currency"`
		} `json:"economy"`
		Laws []struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"laws"`
	} `json:"society"`
}

// Stage6Result 阶段6结果（包含文明和社会）
type Stage6Result struct {
	Civilization *models.Civilization
	Society      *models.Society
}

// Stage7Output 阶段7输出（匹配LLM输出格式）
type Stage7Output struct {
	ConsistencyCheck struct {
		OverallScore int `json:"overall_score"`
		Issues       []struct {
			Aspect     string `json:"aspect"`
			Issue      string `json:"issue"`
			Severity   string `json:"severity"`
			Suggestion string `json:"suggestion"`
		} `json:"issues"`
		Strengths      []string `json:"strengths"`
		Improvements   []string `json:"improvements"`
		StoryPotential struct {
			Score                 int      `json:"score"`
			HighPotentialElements []string `json:"high_potential_elements"`
			UnderutilizedElements []string `json:"underutilized_elements"`
		} `json:"story_potential"`
	} `json:"consistency_check"`
}

// Stage1Output 阶段1输出
type Stage1Output struct {
	CoreQuestion string             `json:"core_question"`
	Derivation   string             `json:"derivation"`
	ValueSystem  models.ValueSystem `json:"value_system"`
	Themes       []models.Theme     `json:"themes"`
}

// Stage2Output 阶段2输出
type Stage2Output struct {
	DerivationLogic string             `json:"derivation_logic"`
	Cosmology       models.Cosmology   `json:"cosmology"`
	Metaphysics     models.Metaphysics `json:"metaphysics"`
}

// Stage3Output 阶段3输出（匹配LLM输出格式）
type Stage3Output struct {
	Physics struct {
		Gravity            string `json:"gravity"`
		TimeFlow           string `json:"time_flow"`
		EnergyConservation string `json:"energy_conservation"`
		Causality          string `json:"causality"`
		DeathNature        string `json:"death_nature"`
	} `json:"physics"`
	Supernatural *struct {
		Exists     bool     `json:"exists"`
		Type       string   `json:"type"`
		Source     string   `json:"source"`
		Cost       string   `json:"cost"`
		Limitation []string `json:"limitation"`
	} `json:"supernatural"`
}

// WorldBuilder 世界设定器
type WorldBuilder struct {
	db      db.Database
	cfg     *config.Config
	client  *llm.Client
	mapping *config.ModuleMapping
}

// New 创建世界设定器
func New() (*WorldBuilder, error) {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 创建LLM客户端
	client, mapping, err := llm.NewClientForModule("world_builder")
	if err != nil {
		return nil, fmt.Errorf("创建LLM客户端失败: %w", err)
	}

	return &WorldBuilder{
		db:      db.Get(),
		cfg:     cfg,
		client:  client,
		mapping: mapping,
	}, nil
}

// Build 完整构建世界（执行所有7个阶段）
func (wb *WorldBuilder) Build(params BuildParams) (*models.WorldSetting, error) {
	// 创建世界设定对象
	world := &models.WorldSetting{
		ID:    db.GenerateID("world"),
		Name:  params.Name,
		Type:  params.Type,
		Scale: params.Scale,
		Style: params.Style,
	}

	// 阶段1: 哲学基础
	philosophy, _, err := wb.GenerateStage1(Stage1Input{
		WorldType: string(params.Type),
		Theme:     params.Theme,
		Style:     params.Style,
	})
	if err != nil {
		return nil, fmt.Errorf("阶段1失败: %w", err)
	}
	world.Philosophy = *philosophy
	if err := wb.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段1失败: %w", err)
	}

	// 阶段2: 世界观
	worldview, _, err := wb.GenerateStage2(Stage2Input{
		CoreQuestion: philosophy.CoreQuestion,
		HighestGood:  philosophy.ValueSystem.HighestGood,
		UltimateEvil: philosophy.ValueSystem.UltimateEvil,
	})
	if err != nil {
		return nil, fmt.Errorf("阶段2失败: %w", err)
	}
	world.Worldview = *worldview
	if err := wb.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段2失败: %w", err)
	}

	// 阶段3: 法则设定
	worldviewSummary := fmt.Sprintf("起源:%s 结构:%s", worldview.Cosmology.Origin, worldview.Cosmology.Structure)
	laws, _, err := wb.GenerateStage3(Stage3Input{
		WorldType: string(params.Type),
		Worldview: worldviewSummary,
	})
	if err != nil {
		return nil, fmt.Errorf("阶段3失败: %w", err)
	}
	world.Laws = *laws
	if err := wb.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段3失败: %w", err)
	}

	// 阶段4: 故事土壤
	// 准备主要矛盾摘要
	mainConflicts := ""
	if len(philosophy.ValueSystem.MoralDilemmas) > 0 {
		mainConflicts = philosophy.ValueSystem.MoralDilemmas[0].Dilemma
	}

	storySoil, _, err := wb.GenerateStage4(Stage4Input{
		CoreQuestion:  philosophy.CoreQuestion,
		MainConflicts: mainConflicts,
		WorldType:     string(params.Type),
	})
	if err != nil {
		return nil, fmt.Errorf("阶段4失败: %w", err)
	}
	world.StorySoil = *storySoil
	if err := wb.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段4失败: %w", err)
	}

	// 阶段5: 地理环境
	// 准备法则摘要
	lawsSummary := fmt.Sprintf("物理:%s 超自然:%v", laws.Physics.Gravity, laws.Supernatural != nil && laws.Supernatural.Exists)
	civilizationNeeds := fmt.Sprintf("资源需求基于%s类型的世界", params.Type)

	geography, _, err := wb.GenerateStage5(Stage5Input{
		WorldType:         string(params.Type),
		WorldScale:        string(params.Scale),
		LawsSummary:       lawsSummary,
		CivilizationNeeds: civilizationNeeds,
	})
	if err != nil {
		return nil, fmt.Errorf("阶段5失败: %w", err)
	}
	world.Geography = *geography
	if err := wb.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段5失败: %w", err)
	}

	// 阶段6: 文明社会
	// 准备地理摘要
	geographySummary := fmt.Sprintf("%d个区域, 气候:%s", len(geography.Regions),
		func() string {
			if geography.Climate != nil {
				return geography.Climate.Type
			}
			return "未知"
		}())
	valueSystem := fmt.Sprintf("最高善:%s", philosophy.ValueSystem.HighestGood)

	civResult, _, err := wb.GenerateStage6(Stage6Input{
		WorldType:        string(params.Type),
		GeographySummary: geographySummary,
		ValueSystem:      valueSystem,
	})
	if err != nil {
		return nil, fmt.Errorf("阶段6失败: %w", err)
	}
	world.Civilization = *civResult.Civilization
	world.Society = *civResult.Society
	if err := wb.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段6失败: %w", err)
	}

	// 阶段7: 一致性检查
	// 构建世界设定摘要
	worldSummary := wb.buildWorldSummary(world)
	report, _, err := wb.GenerateStage7(Stage7Input{
		WorldSettingSummary: worldSummary,
	})
	if err != nil {
		return nil, fmt.Errorf("阶段7失败: %w", err)
	}
	world.ConsistencyReport = report
	if err := wb.db.SaveWorld(world); err != nil {
		return nil, fmt.Errorf("保存阶段7失败: %w", err)
	}

	return world, nil
}

// GenerateStage1 生成阶段1：哲学基础
func (wb *WorldBuilder) GenerateStage1(input Stage1Input) (*models.Philosophy, string, error) {
	// 准备模板数据
	data := map[string]interface{}{
		"WorldType": input.WorldType,
		"Theme":     input.Theme,
		"Style":     input.Style,
	}

	// 渲染提示词
	prompt, err := wb.cfg.GetWorldBuilderStage1(data)
	if err != nil {
		return nil, "", fmt.Errorf("渲染提示词失败: %w", err)
	}

	// 获取系统提示词
	systemPrompt := wb.cfg.GetWorldBuilderSystem()

	// 调用LLM（带重试）
	result, err := wb.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, "", err
	}

	// 解析输出
	var output Stage1Output
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		// 尝试提取JSON
		extracted := extractJSON(result)
		if err := json.Unmarshal([]byte(extracted), &output); err != nil {
			return nil, "", fmt.Errorf("解析LLM输出失败: %w, 原始内容: %s", err, result[:min(500, len(result))])
		}
	}

	// 构建哲学对象
	philosophy := &models.Philosophy{
		CoreQuestion: output.CoreQuestion,
		Derivation:   output.Derivation,
		ValueSystem:  output.ValueSystem,
		Themes:       output.Themes,
	}

	return philosophy, prompt, nil
}

// GenerateStage1ForWorld 为已有世界生成/重新生成阶段1
func (wb *WorldBuilder) GenerateStage1ForWorld(worldID string, input Stage1Input) error {
	philosophy, _, err := wb.GenerateStage1(input)
	if err != nil {
		return err
	}

	return wb.db.UpdateWorldStage(worldID, "philosophy", philosophy)
}

// GenerateStage2 生成阶段2：世界观
func (wb *WorldBuilder) GenerateStage2(input Stage2Input) (*models.Worldview, string, error) {
	// 准备模板数据
	data := map[string]interface{}{
		"CoreQuestion": input.CoreQuestion,
		"HighestGood":  input.HighestGood,
		"UltimateEvil": input.UltimateEvil,
	}

	// 渲染提示词
	prompt, err := wb.cfg.GetWorldBuilderStage2(data)
	if err != nil {
		return nil, "", fmt.Errorf("渲染提示词失败: %w", err)
	}

	// 获取系统提示词
	systemPrompt := wb.cfg.GetWorldBuilderSystem()

	// 调用LLM（带重试）
	result, err := wb.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, "", err
	}

	// 解析输出
	var output Stage2Output
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		// 尝试提取JSON
		extracted := extractJSON(result)
		if err := json.Unmarshal([]byte(extracted), &output); err != nil {
			return nil, "", fmt.Errorf("解析LLM输出失败: %w, 原始内容: %s", err, result[:min(500, len(result))])
		}
	}

	// 构建世界观对象
	worldview := &models.Worldview{
		Derivation:  output.DerivationLogic,
		Cosmology:   output.Cosmology,
		Metaphysics: output.Metaphysics,
	}

	return worldview, prompt, nil
}

// GenerateStage2ForWorld 为已有世界生成/重新生成阶段2
func (wb *WorldBuilder) GenerateStage2ForWorld(worldID string, input Stage2Input) error {
	worldview, _, err := wb.GenerateStage2(input)
	if err != nil {
		return err
	}

	return wb.db.UpdateWorldStage(worldID, "worldview", worldview)
}

// GenerateStage3 生成阶段3：法则设定
func (wb *WorldBuilder) GenerateStage3(input Stage3Input) (*models.Laws, string, error) {
	// 准备模板数据
	data := map[string]interface{}{
		"WorldType": input.WorldType,
		"Worldview": input.Worldview,
	}

	// 渲染提示词
	prompt, err := wb.cfg.GetWorldBuilderStage3(data)
	if err != nil {
		return nil, "", fmt.Errorf("渲染提示词失败: %w", err)
	}

	// 获取系统提示词
	systemPrompt := wb.cfg.GetWorldBuilderSystem()

	// 调用LLM（带重试）
	result, err := wb.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, "", err
	}

	// 解析输出
	var output Stage3Output
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		// 尝试提取JSON
		extracted := extractJSON(result)
		if err := json.Unmarshal([]byte(extracted), &output); err != nil {
			return nil, "", fmt.Errorf("解析LLM输出失败: %w, 原始内容: %s", err, result[:min(500, len(result))])
		}
	}

	// 构建法则对象（需要将LLM输出映射到模型结构）
	laws := &models.Laws{
		Physics: models.Physics{
			Gravity:            output.Physics.Gravity,
			TimeFlow:           output.Physics.TimeFlow,
			EnergyConservation: output.Physics.EnergyConservation,
			Causality:          output.Physics.Causality,
			DeathNature:        output.Physics.DeathNature,
		},
	}

	// 处理超自然体系
	if output.Supernatural != nil && output.Supernatural.Exists {
		supernatural := &models.Supernatural{
			Exists: output.Supernatural.Exists,
			Type:   output.Supernatural.Type,
		}

		// 根据类型创建详细设定
		settings := &models.SupernaturalSettings{}

		switch output.Supernatural.Type {
		case "magic":
			settings.MagicSystem = &models.MagicSystem{
				Source:     output.Supernatural.Source,
				Cost:       output.Supernatural.Cost,
				Limitation: output.Supernatural.Limitation,
			}
		case "cultivation":
			settings.CultivationSystem = &models.CultivationSystem{
				// 简化的修真体系设定，实际由LLM生成
				ResourceSystem: output.Supernatural.Source,
			}
			if len(output.Supernatural.Limitation) > 0 {
				settings.CultivationSystem.Bottleneck = output.Supernatural.Limitation[0]
			}
		case "superpower":
			settings.SuperpowerSystem = &models.SuperpowerSystem{
				Origin: output.Supernatural.Source,
				Type:   output.Supernatural.Type,
				Limit:  output.Supernatural.Limitation,
			}
		}

		if settings.MagicSystem != nil || settings.CultivationSystem != nil || settings.SuperpowerSystem != nil {
			supernatural.Settings = settings
		}

		laws.Supernatural = supernatural
	}

	return laws, prompt, nil
}

// GenerateStage3ForWorld 为已有世界生成/重新生成阶段3
func (wb *WorldBuilder) GenerateStage3ForWorld(worldID string, input Stage3Input) error {
	laws, _, err := wb.GenerateStage3(input)
	if err != nil {
		return err
	}

	return wb.db.UpdateWorldStage(worldID, "laws", laws)
}

// GenerateStage4 生成阶段4：故事土壤
func (wb *WorldBuilder) GenerateStage4(input Stage4Input) (*models.StorySoil, string, error) {
	// 准备模板数据
	data := map[string]interface{}{
		"CoreQuestion":  input.CoreQuestion,
		"MainConflicts": input.MainConflicts,
		"WorldType":     input.WorldType,
	}

	// 渲染提示词
	prompt, err := wb.cfg.GetWorldBuilderStage4(data)
	if err != nil {
		return nil, "", fmt.Errorf("渲染提示词失败: %w", err)
	}

	// 获取系统提示词
	systemPrompt := wb.cfg.GetWorldBuilderSystem()

	// 调用LLM（带重试）
	result, err := wb.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, "", err
	}

	// 解析输出
	var output models.StorySoil
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		// 尝试提取JSON
		extracted := extractJSON(result)
		if err := json.Unmarshal([]byte(extracted), &output); err != nil {
			return nil, "", fmt.Errorf("解析LLM输出失败: %w, 原始内容: %s", err, result[:min(500, len(result))])
		}
	}

	return &output, prompt, nil
}

// GenerateStage4ForWorld 为已有世界生成/重新生成阶段4
func (wb *WorldBuilder) GenerateStage4ForWorld(worldID string, input Stage4Input) error {
	storySoil, _, err := wb.GenerateStage4(input)
	if err != nil {
		return err
	}

	return wb.db.UpdateWorldStage(worldID, "story_soil", storySoil)
}

// GenerateStage5 生成阶段5：地理环境
func (wb *WorldBuilder) GenerateStage5(input Stage5Input) (*models.Geography, string, error) {
	// 准备模板数据
	data := map[string]interface{}{
		"WorldType":         input.WorldType,
		"WorldScale":        input.WorldScale,
		"LawsSummary":       input.LawsSummary,
		"CivilizationNeeds": input.CivilizationNeeds,
	}

	var prompt string
	var err error

	// 历史世界使用专门的提示词，生成真实地理名称
	if input.WorldType == string(models.WorldHistorical) {
		prompt = wb.buildHistoricalGeographyPrompt(data)
	} else {
		// 渲染提示词
		prompt, err = wb.cfg.GetWorldBuilderStage5(data)
		if err != nil {
			return nil, "", fmt.Errorf("渲染提示词失败: %w", err)
		}
	}

	// 获取系统提示词
	systemPrompt := wb.cfg.GetWorldBuilderSystem()

	// 调用LLM（带重试）
	result, err := wb.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, "", err
	}

	// 解析输出
	var output Stage5Output
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		// 尝试提取JSON
		extracted := extractJSON(result)
		if err := json.Unmarshal([]byte(extracted), &output); err != nil {
			return nil, "", fmt.Errorf("解析LLM输出失败: %w, 原始内容: %s", err, result[:min(500, len(result))])
		}
	}

	// 构建地理对象
	geography := &models.Geography{
		Regions: make([]models.Region, len(output.Geography.Regions)),
	}

	// 映射区域
	for i, r := range output.Geography.Regions {
		geography.Regions[i] = models.Region{
			ID:          r.ID,
			Name:        r.Name,
			Type:        r.Type,
			Description: r.Description,
			Resources:   r.Resources,
			Risks:       r.Risks,
		}
	}

	// 映射资源
	if output.Geography.Resources != nil {
		geography.Resources = &models.Resources{
			Basic:     output.Geography.Resources.Basic,
			Strategic: output.Geography.Resources.Strategic,
			Rare:      output.Geography.Resources.Rare,
		}
	}

	// 映射气候
	if output.Geography.Climate != nil {
		geography.Climate = &models.Climate{
			Type:     output.Geography.Climate.Type,
			Seasons:  output.Geography.Climate.Seasons,
			Features: output.Geography.Climate.Features,
		}
	}

	return geography, prompt, nil
}

// GenerateStage5ForWorld 为已有世界生成/重新生成阶段5
func (wb *WorldBuilder) GenerateStage5ForWorld(worldID string, input Stage5Input) error {
	geography, _, err := wb.GenerateStage5(input)
	if err != nil {
		return err
	}

	return wb.db.UpdateWorldStage(worldID, "geography", geography)
}

// buildHistoricalGeographyPrompt 构建历史世界地理提示词
func (wb *WorldBuilder) buildHistoricalGeographyPrompt(data map[string]interface{}) string {
	worldScale := data["WorldScale"].(string)
	lawsSummary := ""
	if v, ok := data["LawsSummary"].(string); ok {
		lawsSummary = v
	}
	civilizationNeeds := ""
	if v, ok := data["CivilizationNeeds"].(string); ok {
		civilizationNeeds = v
	}

	var prompt strings.Builder

	prompt.WriteString("# 历史世界地理环境生成\n\n")
	prompt.WriteString("世界类型：历史 (historical)\n")
	prompt.WriteString(fmt.Sprintf("世界规模：%s\n", worldScale))
	if lawsSummary != "" {
		prompt.WriteString(fmt.Sprintf("法则设定：%s\n", lawsSummary))
	}
	if civilizationNeeds != "" {
		prompt.WriteString(fmt.Sprintf("文明需求：%s\n", civilizationNeeds))
	}

	prompt.WriteString("\n# 重要要求\n")
	prompt.WriteString("这是【历史世界】，必须使用真实的历史地理名称！\n")
	prompt.WriteString("禁止使用奇幻风格的名称（如「意识平原」「平衡之河」等）。\n\n")

	prompt.WriteString("# 参考示例\n")
	prompt.WriteString("## 中国历史地名参考\n")
	prompt.WriteString("### 城市区域类\n")
	prompt.WriteString("- 城市中心/商业区：外滩、南京路、王府井、西单、春熙路、解放碑\n")
	prompt.WriteString("- 居住区：静安寺、徐汇、虹口、闸北、宣武、崇文\n")
	prompt.WriteString("- 租界区：法租界、英租界、公共租界、日租界\n")
	prompt.WriteString("- 老城区：城隍庙、豫园、前门、大栅栏\n\n")

	prompt.WriteString("### 自然地理类\n")
	prompt.WriteString("- 河流：黄浦江、苏州河、长江、珠江、渭水\n")
	prompt.WriteString("- 湖泊：西湖、太湖、洞庭湖、鄱阳湖\n")
	prompt.WriteString("- 山脉：黄山、庐山、泰山、华山、衡山\n")
	prompt.WriteString("- 平原：长江三角洲、珠江三角洲、关中平原\n")
	prompt.WriteString("- 丘陵：江南丘陵、浙闽丘陵\n\n")

	prompt.WriteString("### 郊区乡村类\n")
	prompt.WriteString("- 郊县：松江、青浦、奉贤、南汇、川沙\n")
	prompt.WriteString("- 乡村：水乡、山村、渔村、古镇\n\n")

	prompt.WriteString("# 生成任务\n")
	prompt.WriteString("基于以上信息生成历史世界的【地理环境】，请生成以下内容并以JSON格式返回：\n")
	prompt.WriteString(`{
  "geography": {
    "regions": [
      {
        "id": "region_001",
        "name": "真实的历史地名（如：外滩）",
        "type": "urban/plain/river/lake/mountain/hill/forest",
        "description": "符合历史背景的详细描述",
        "resources": ["符合时代的资源1", "资源2"],
        "risks": ["符合时代背景的风险1", "风险2"]
      }
    ],
    "resources": {
      "basic": ["符合时代的基础资源"],
      "strategic": ["符合时代的战略资源"],
      "rare": ["符合时代的稀有资源"]
    },
    "climate": {
      "type": "符合该时代该地区的气候类型",
      "seasons": true,
      "features": ["气候特征1", "特征2"]
    }
  }
}

只返回JSON，不要包含其他内容。`)

	return prompt.String()
}

// GenerateStage6 生成阶段6：文明社会
func (wb *WorldBuilder) GenerateStage6(input Stage6Input) (*Stage6Result, string, error) {
	// 准备模板数据
	data := map[string]interface{}{
		"WorldType":        input.WorldType,
		"GeographySummary": input.GeographySummary,
		"ValueSystem":      input.ValueSystem,
	}

	// 渲染提示词
	prompt, err := wb.cfg.GetWorldBuilderStage6(data)
	if err != nil {
		return nil, "", fmt.Errorf("渲染提示词失败: %w", err)
	}

	// 获取系统提示词
	systemPrompt := wb.cfg.GetWorldBuilderSystem()

	// 调用LLM（带重试）
	result, err := wb.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, "", err
	}

	// 解析输出
	var output Stage6Output
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		// 尝试提取JSON
		extracted := extractJSON(result)
		if err := json.Unmarshal([]byte(extracted), &output); err != nil {
			return nil, "", fmt.Errorf("解析LLM输出失败: %w, 原始内容: %s", err, result[:min(500, len(result))])
		}
	}

	// 构建文明对象
	civilization := &models.Civilization{
		Races:     make([]models.Race, len(output.Civilization.Races)),
		Languages: make([]models.Language, len(output.Civilization.Languages)),
		Religions: make([]models.Religion, len(output.Civilization.Religions)),
	}

	// 映射种族
	for i, r := range output.Civilization.Races {
		civilization.Races[i] = models.Race{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Traits:      r.Traits,
			Abilities:   r.Abilities,
			Relations:   r.Relations,
		}
	}

	// 映射语言
	for i, l := range output.Civilization.Languages {
		civilization.Languages[i] = models.Language{
			ID:       l.ID,
			Name:     l.Name,
			Type:     l.Type,
			Speakers: l.Speakers,
			Features: l.Features,
		}
	}

	// 映射宗教
	for i, r := range output.Civilization.Religions {
		civilization.Religions[i] = models.Religion{
			ID:           r.ID,
			Name:         r.Name,
			Type:         r.Type,
			Cosmology:    r.Cosmology,
			Ethics:       r.Ethics,
			Practices:    r.Practices,
			Organization: r.Organization,
		}
	}

	// 构建社会对象
	powerStructure := &models.PowerStructure{
		Formal:            make([]models.PowerLevel, len(output.Society.Politics.PowerStructure.Formal)),
		Actual:            make([]models.PowerHolder, len(output.Society.Politics.PowerStructure.Actual)),
		ChecksAndBalances: "",
	}

	// 映射权力层级
	for i, p := range output.Society.Politics.PowerStructure.Formal {
		powerStructure.Formal[i] = models.PowerLevel{
			Level:  p.Level,
			Name:   p.Name,
			Powers: p.Powers,
		}
	}

	// 映射实际掌权者
	for i, p := range output.Society.Politics.PowerStructure.Actual {
		powerStructure.Actual[i] = models.PowerHolder{
			Entity:       p.Entity,
			PowerSource:  p.PowerSource,
			Relationship: p.Relationship,
		}
	}

	society := &models.Society{
		Politics: models.Politics{
			Type:             output.Society.Politics.Type,
			LegitimacySource: output.Society.Politics.LegitimacySource,
			PowerStructure:   powerStructure,
		},
		Classes: make([]models.Class, len(output.Society.Classes)),
		Economy: models.Economy{
			Type:         output.Society.Economy.Type,
			TradeNetwork: output.Society.Economy.TradeNetwork,
			Currency:     output.Society.Economy.Currency,
		},
		Laws: make([]models.Law, len(output.Society.Laws)),
	}

	// 映射社会阶级
	for i, c := range output.Society.Classes {
		society.Classes[i] = models.Class{
			Name:        c.Name,
			Rank:        c.Rank,
			Rights:      c.Rights,
			Obligations: c.Obligations,
		}
	}

	// 映射法律
	for i, l := range output.Society.Laws {
		society.Laws[i] = models.Law{
			Name:        l.Name,
			Type:        l.Type,
			Description: l.Description,
			Penalty:     "", // 阶段6不生成惩罚细节
		}
	}

	return &Stage6Result{
		Civilization: civilization,
		Society:      society,
	}, prompt, nil
}

// GenerateStage6ForWorld 为已有世界生成/重新生成阶段6
func (wb *WorldBuilder) GenerateStage6ForWorld(worldID string, input Stage6Input) error {
	result, _, err := wb.GenerateStage6(input)
	if err != nil {
		return err
	}

	// 更新文明
	if err := wb.db.UpdateWorldStage(worldID, "civilization", result.Civilization); err != nil {
		return err
	}

	// 更新社会
	return wb.db.UpdateWorldStage(worldID, "society", result.Society)
}

// GenerateStage7 生成阶段7：一致性检查
func (wb *WorldBuilder) GenerateStage7(input Stage7Input) (*models.ConsistencyReport, string, error) {
	// 准备模板数据
	data := map[string]interface{}{
		"WorldSettingSummary": input.WorldSettingSummary,
	}

	// 渲染提示词
	prompt, err := wb.cfg.GetWorldBuilderStage7(data)
	if err != nil {
		return nil, "", fmt.Errorf("渲染提示词失败: %w", err)
	}

	// 获取系统提示词
	systemPrompt := wb.cfg.GetWorldBuilderSystem()

	// 调用LLM（带重试）
	result, err := wb.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, "", err
	}

	// 解析输出
	var output Stage7Output
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		// 尝试提取JSON
		extracted := extractJSON(result)
		if err := json.Unmarshal([]byte(extracted), &output); err != nil {
			return nil, "", fmt.Errorf("解析LLM输出失败: %w, 原始内容: %s", err, result[:min(500, len(result))])
		}
	}

	// 构建一致性报告
	report := &models.ConsistencyReport{
		OverallScore: output.ConsistencyCheck.OverallScore,
		Issues:       make([]models.ConsistencyIssue, len(output.ConsistencyCheck.Issues)),
		Strengths:    output.ConsistencyCheck.Strengths,
		Improvements: output.ConsistencyCheck.Improvements,
		StoryPotential: models.StoryPotential{
			Score:                 output.ConsistencyCheck.StoryPotential.Score,
			HighPotentialElements: output.ConsistencyCheck.StoryPotential.HighPotentialElements,
			UnderutilizedElements: output.ConsistencyCheck.StoryPotential.UnderutilizedElements,
		},
	}

	// 映射问题
	for i, issue := range output.ConsistencyCheck.Issues {
		report.Issues[i] = models.ConsistencyIssue{
			Aspect:     issue.Aspect,
			Issue:      issue.Issue,
			Severity:   issue.Severity,
			Suggestion: issue.Suggestion,
		}
	}

	return report, prompt, nil
}

// GenerateStage7ForWorld 为已有世界生成/重新生成阶段7
// 注意：阶段7是检查阶段，返回报告但不修改世界设定
func (wb *WorldBuilder) GenerateStage7ForWorld(worldID string) (*models.ConsistencyReport, string, error) {
	world, err := wb.db.GetWorld(worldID)
	if err != nil {
		return nil, "", fmt.Errorf("获取世界失败: %w", err)
	}

	// 构建世界设定摘要
	summary := wb.buildWorldSummary(world)

	return wb.GenerateStage7(Stage7Input{
		WorldSettingSummary: summary,
	})
}

// buildWorldSummary 构建世界设定摘要（用于阶段7）
func (wb *WorldBuilder) buildWorldSummary(world *models.WorldSetting) string {
	summary := fmt.Sprintf("世界名称: %s\n类型: %s\n规模: %s\n\n",
		world.Name, world.Type, world.Scale)

	summary += fmt.Sprintf("【哲学】核心问题: %s\n", world.Philosophy.CoreQuestion)

	origin := world.Worldview.Cosmology.Origin
	if len(origin) > 100 {
		origin = origin[:100] + "..."
	}
	summary += fmt.Sprintf("【世界观】起源: %s\n", origin)

	summary += fmt.Sprintf("【法则】重力: %s\n", world.Laws.Physics.Gravity)

	if len(world.Geography.Regions) > 0 {
		summary += fmt.Sprintf("【地理】区域数量: %d\n", len(world.Geography.Regions))
	}

	if len(world.Civilization.Races) > 0 {
		summary += fmt.Sprintf("【文明】种族数量: %d\n", len(world.Civilization.Races))
	}

	if len(world.Society.Classes) > 0 {
		summary += fmt.Sprintf("【社会】阶级数量: %d\n", len(world.Society.Classes))
	}

	return summary
}

// callWithRetry 调用LLM并自动重试
func (wb *WorldBuilder) callWithRetry(prompt, systemPrompt string) (string, error) {
	retryConfig := wb.cfg.System.Retry
	maxAttempts := retryConfig.MaxAttempts
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// 调用LLM
		result, err := wb.client.GenerateJSONWithParams(
			prompt,
			systemPrompt,
			wb.mapping.Temperature,
			wb.mapping.MaxTokens,
		)

		if err == nil {
			// 成功，转换为JSON字符串返回
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("序列化结果失败: %w", err)
			}
			return string(jsonBytes), nil
		}

		lastErr = err

		// 如果不是最后一次尝试，等待后重试
		if attempt < maxAttempts {
			delay := time.Duration(retryConfig.InitialDelay*attempt) * time.Second
			if delay > time.Duration(retryConfig.MaxDelay)*time.Second {
				delay = time.Duration(retryConfig.MaxDelay) * time.Second
			}
			time.Sleep(delay)
		}
	}

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

// indexOf 查找字节切片位置
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

// lastIndexOf 从后查找字节切片位置
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

// GetProgress 获取构建进度（0-100）
func (wb *WorldBuilder) GetProgress(worldID string) (int, error) {
	world, err := wb.db.GetWorld(worldID)
	if err != nil {
		return 0, err
	}

	progress := 0
	stageWeight := 100 / 7 // 每阶段约14.3%

	// 检查每个阶段
	if world.Philosophy.CoreQuestion != "" {
		progress += stageWeight // 阶段1完成
	}
	if world.Worldview.Derivation != "" {
		progress += stageWeight // 阶段2完成
	}
	if world.Laws.Physics.Gravity != "" {
		progress += stageWeight // 阶段3完成
	}
	if len(world.StorySoil.SocialConflicts) > 0 {
		progress += stageWeight // 阶段4完成
	}
	if len(world.Geography.Regions) > 0 {
		progress += stageWeight // 阶段5完成
	}
	if len(world.Civilization.Races) > 0 {
		progress += stageWeight // 阶段6完成
	}
	if world.ConsistencyReport != nil {
		progress += stageWeight // 阶段7完成
	}

	return progress, nil
}
