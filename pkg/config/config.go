// Package config 提供配置管理功能
// 所有提示词和LLM配置均通过此包加载，不得硬编码
package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构
type Config struct {
	LLM     LLMConfig            `yaml:"llm"`
	Prompts PromptsConfig        `yaml:"prompts"`
	System  SystemConfig         `yaml:"system"`
}

// LLMConfig LLM相关配置
type LLMConfig struct {
	DefaultProvider string                   `yaml:"default_provider"`
	Providers       map[string]ProviderConfig `yaml:"providers"`
	ModuleMapping   map[string]ModuleMapping `yaml:"module_mapping"`
}

// ProviderConfig LLM提供商配置
type ProviderConfig struct {
	BaseURL   string            `yaml:"base_url"`
	APIKey    string            `yaml:"api_key"`
	APIKeyEnv string            `yaml:"api_key_env"`
	Models    ModelsConfig      `yaml:"models"`
}

// ModelsConfig 模型配置
type ModelsConfig struct {
	Default  string          `yaml:"default"`
	Available []ModelInfo    `yaml:"available"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	Name            string  `yaml:"name"`
	MaxTokens       int     `yaml:"max_tokens"`
	CostPer1kInput  float64 `yaml:"cost_per_1k_input"`
	CostPer1kOutput float64 `yaml:"cost_per_1k_output"`
}

// ModuleMapping 模块与模型的映射
type ModuleMapping struct {
	Provider    string  `yaml:"provider"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

// PromptsConfig 提示词配置
type PromptsConfig struct {
	WorldBuilder    WorldBuilderPrompts    `yaml:"world_builder"`
	NarrativeEngine NarrativeEnginePrompts `yaml:"narrative_engine"`
	Writer          WriterPrompts          `yaml:"writer"`
	Character       CharacterPrompts       `yaml:"character"`
}

// WorldBuilderPrompts 世界构建器提示词
type WorldBuilderPrompts struct {
	System             string `yaml:"system"`
	Stage1Philosophy   string `yaml:"stage1_philosophy"`
	Stage2Worldview    string `yaml:"stage2_worldview"`
	Stage3Laws         string `yaml:"stage3_laws"`
	Stage4StorySoil    string `yaml:"stage4_story_soil"`
	Stage5Geography    string `yaml:"stage5_geography"`
	Stage6Civilization string `yaml:"stage6_civilization"`
	Stage7Consistency  string `yaml:"stage7_consistency"`
}

// NarrativeEnginePrompts 叙事引擎提示词
type NarrativeEnginePrompts struct {
	System              string `yaml:"system"`
	GenerateOutline     string `yaml:"generate_outline"`
	GenerateChapterPlans string `yaml:"generate_chapter_plans"`
	GenerateScenes      string `yaml:"generate_scenes"`
	PlanCharacterArc    string `yaml:"plan_character_arc"`
}

// WriterPrompts 写作器提示词
type WriterPrompts struct {
	System                 string `yaml:"system"`
	GenerateDialogue       string `yaml:"generate_dialogue"`
	GenerateScene          string `yaml:"generate_scene"`
	GenerateAction         string `yaml:"generate_action"`
	GenerateEnvironment    string `yaml:"generate_environment"`
	GenerateInternalMonologue string `yaml:"generate_internal_monologue"`
}

// CharacterPrompts 角色提示词
type CharacterPrompts struct {
	System         string `yaml:"system"`
	GenerateProfile string `yaml:"generate_profile"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	Project ProjectConfig `yaml:"project"`
	Logging LoggingConfig `yaml:"logging"`
	Retry   RetryConfig   `yaml:"retry"`
	Timeout TimeoutConfig `yaml:"timeout"`
}

// ProjectConfig 项目配置
type ProjectConfig struct {
	MaxParallelStories int  `yaml:"max_parallel_stories"`
	AutoSaveInterval   int  `yaml:"auto_save_interval"`
	BackupEnabled      bool `yaml:"backup_enabled"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts  int `yaml:"max_attempts"`
	InitialDelay int `yaml:"initial_delay"`
	MaxDelay     int `yaml:"max_delay"`
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	LLMRequest         int `yaml:"llm_request"`
	ChapterGeneration  int `yaml:"chapter_generation"`
}

var (
	globalConfig *Config
)

// Load 加载配置文件
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置为全局配置
	globalConfig = cfg

	return cfg, nil
}

// LoadDefault 加载默认配置文件
func LoadDefault() (*Config, error) {
	configPaths := []string{
		"config/config.yaml",
		"./config.yaml",
		"/etc/xupu/config.yaml",
	}

	for _, path := range configPaths {
		if cfg, err := Load(path); err == nil {
			return cfg, nil
		}
	}

	return nil, fmt.Errorf("未找到配置文件，尝试的路径: %v", configPaths)
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("配置未初始化，请先调用 Load() 或 LoadDefault()")
	}
	return globalConfig
}

// GetAPIKey 获取API Key（优先从配置文件读取，失败则从环境变量读取）
func (c *ProviderConfig) GetAPIKey() (string, error) {
	// 优先使用配置文件中的 api_key
	if c.APIKey != "" {
		return c.APIKey, nil
	}

	// 备用：从环境变量读取
	if c.APIKeyEnv == "" {
		return "", fmt.Errorf("未配置 api_key（配置文件）或 api_key_env（环境变量）")
	}
	apiKey := os.Getenv(c.APIKeyEnv)
	if apiKey == "" {
		return "", fmt.Errorf("配置文件中未设置 api_key，且环境变量 %s 未设置", c.APIKeyEnv)
	}
	return apiKey, nil
}

// GetModuleConfig 获取模块的LLM配置
func (c *LLMConfig) GetModuleConfig(moduleName string) (*ModuleMapping, *ProviderConfig, error) {
	mapping, ok := c.ModuleMapping[moduleName]
	if !ok {
		return nil, nil, fmt.Errorf("未找到模块 %s 的配置", moduleName)
	}

	provider, ok := c.Providers[mapping.Provider]
	if !ok {
		return nil, nil, fmt.Errorf("未找到提供商 %s 的配置", mapping.Provider)
	}

	return &mapping, &provider, nil
}

// RenderPrompt 渲染提示词模板
func RenderPrompt(template string, data map[string]interface{}) (string, error) {
	var buf bytes.Buffer

	// 将map转换为适合模板的数据结构
	tmplData := make(map[string]interface{})
	for k, v := range data {
		tmplData[k] = v
	}

	// 简单的模板替换，替换 {{.Key}} 格式
	result := template
	for key, value := range tmplData {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = v
		case fmt.Stringer:
			valueStr = v.String()
		default:
			valueStr = fmt.Sprintf("%v", v)
		}
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}

	// 处理换行
	result = strings.ReplaceAll(result, "\\n", "\n")

	buf.WriteString(result)
	return buf.String(), nil
}

// GetWorldBuilderSystem 获取世界构建器系统提示词
func (c *Config) GetWorldBuilderSystem() string {
	return c.Prompts.WorldBuilder.System
}

// GetWorldBuilderStage1 获取世界构建器第一阶段提示词
func (c *Config) GetWorldBuilderStage1(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.WorldBuilder.Stage1Philosophy, data)
}

// GetWorldBuilderStage2 获取世界构建器第二阶段提示词
func (c *Config) GetWorldBuilderStage2(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.WorldBuilder.Stage2Worldview, data)
}

// GetWorldBuilderStage3 获取世界构建器第三阶段提示词
func (c *Config) GetWorldBuilderStage3(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.WorldBuilder.Stage3Laws, data)
}

// GetWorldBuilderStage4 获取世界构建器第四阶段提示词
func (c *Config) GetWorldBuilderStage4(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.WorldBuilder.Stage4StorySoil, data)
}

// GetWorldBuilderStage5 获取世界构建器第五阶段提示词
func (c *Config) GetWorldBuilderStage5(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.WorldBuilder.Stage5Geography, data)
}

// GetWorldBuilderStage6 获取世界构建器第六阶段提示词
func (c *Config) GetWorldBuilderStage6(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.WorldBuilder.Stage6Civilization, data)
}

// GetWorldBuilderStage7 获取世界构建器第七阶段提示词
func (c *Config) GetWorldBuilderStage7(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.WorldBuilder.Stage7Consistency, data)
}

// GetNarrativeEngineSystem 获取叙事引擎系统提示词
func (c *Config) GetNarrativeEngineSystem() string {
	return c.Prompts.NarrativeEngine.System
}

// GetNarrativeOutline 获取叙事引擎大纲生成提示词
func (c *Config) GetNarrativeOutline(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.NarrativeEngine.GenerateOutline, data)
}

// GetNarrativeChapterPlans 获取叙事引擎章节规划提示词
func (c *Config) GetNarrativeChapterPlans(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.NarrativeEngine.GenerateChapterPlans, data)
}

// GetNarrativeScenes 获取叙事引擎场景生成提示词
func (c *Config) GetNarrativeScenes(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.NarrativeEngine.GenerateScenes, data)
}

// GetNarrativeCharacterArc 获取叙事引擎角色弧光提示词
func (c *Config) GetNarrativeCharacterArc(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.NarrativeEngine.PlanCharacterArc, data)
}

// GetWriterSystem 获取写作器系统提示词
func (c *Config) GetWriterSystem() string {
	return c.Prompts.Writer.System
}

// GetWriterDialogue 获取写作器对话生成提示词
func (c *Config) GetWriterDialogue(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.Writer.GenerateDialogue, data)
}

// GetWriterScene 获取写作器场景生成提示词
func (c *Config) GetWriterScene(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.Writer.GenerateScene, data)
}

// GetWriterAction 获取写作器动作生成提示词
func (c *Config) GetWriterAction(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.Writer.GenerateAction, data)
}

// GetWriterEnvironment 获取写作器环境生成提示词
func (c *Config) GetWriterEnvironment(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.Writer.GenerateEnvironment, data)
}

// GetWriterInternalMonologue 获取写作器内心独白生成提示词
func (c *Config) GetWriterInternalMonologue(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.Writer.GenerateInternalMonologue, data)
}

// GetCharacterSystem 获取角色系统提示词
func (c *Config) GetCharacterSystem() string {
	return c.Prompts.Character.System
}

// GetCharacterProfile 获取角色档案生成提示词
func (c *Config) GetCharacterProfile(data map[string]interface{}) (string, error) {
	return RenderPrompt(c.Prompts.Character.GenerateProfile, data)
}

// GetAllPrompts 获取所有提示词（用于调试）
func (c *Config) GetAllPrompts() map[string]string {
	return map[string]string{
		"world_builder.system":              c.Prompts.WorldBuilder.System,
		"world_builder.stage1_philosophy":   c.Prompts.WorldBuilder.Stage1Philosophy,
		"world_builder.stage2_worldview":    c.Prompts.WorldBuilder.Stage2Worldview,
		"world_builder.stage3_laws":         c.Prompts.WorldBuilder.Stage3Laws,
		"world_builder.stage4_story_soil":   c.Prompts.WorldBuilder.Stage4StorySoil,
		"world_builder.stage5_geography":    c.Prompts.WorldBuilder.Stage5Geography,
		"world_builder.stage6_civilization": c.Prompts.WorldBuilder.Stage6Civilization,
		"world_builder.stage7_consistency":  c.Prompts.WorldBuilder.Stage7Consistency,

		"narrative_engine.system":                   c.Prompts.NarrativeEngine.System,
		"narrative_engine.generate_outline":         c.Prompts.NarrativeEngine.GenerateOutline,
		"narrative_engine.generate_chapter_plans":   c.Prompts.NarrativeEngine.GenerateChapterPlans,
		"narrative_engine.generate_scenes":          c.Prompts.NarrativeEngine.GenerateScenes,
		"narrative_engine.plan_character_arc":       c.Prompts.NarrativeEngine.PlanCharacterArc,

		"writer.system":                      c.Prompts.Writer.System,
		"writer.generate_dialogue":           c.Prompts.Writer.GenerateDialogue,
		"writer.generate_scene":              c.Prompts.Writer.GenerateScene,
		"writer.generate_action":             c.Prompts.Writer.GenerateAction,
		"writer.generate_environment":        c.Prompts.Writer.GenerateEnvironment,
		"writer.generate_internal_monologue":  c.Prompts.Writer.GenerateInternalMonologue,

		"character.system":         c.Prompts.Character.System,
		"character.generate_profile": c.Prompts.Character.GenerateProfile,
	}
}
