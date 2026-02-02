// Package writer 写作器 - 负责生成实际的小说文本
// 根据叙事器的场景指令，生成高质量的场景内容
package writer

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

// GenerateParams 生成参数
type GenerateParams struct {
	BlueprintID      string            // 蓝图ID
	Chapter          int               // 章节号
	Scene            int               // 场景号
	Instruction      *models.SceneInstruction // 场景指令
	PreviousSummary  string            // 前情摘要
	CharacterStates  map[string]*CharacterContext // 角色状态
	WorldContext     *models.WorldSetting // 世界设定上下文
	Style            StyleConfig       // 风格配置
}

// CharacterContext 角色上下文
type CharacterContext struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	CurrentEmotion string           `json:"current_emotion"`
	Location      string            `json:"location"`
	Knowledge     []string          `json:"knowledge"`
	Relationships map[string]string `json:"relationships"` // 与其他角色的关系
}

// StyleConfig 风格配置
type StyleConfig struct {
	Voice        string `json:"voice"`         // 叙述声音：第一人称/第三人称
	Tense        string `json:"tense"`         // 时态：过去/现在
	Tone         string `json:"tone"`          // 基调：轻松/严肃/黑暗
	Pacing       string `json:"pacing"`        // 节奏：快/中/慢
	DetailLevel  string `json:"detail_level"`  // 细节程度：简洁/适中/丰富
	DialogueRatio float64 `json:"dialogue_ratio"` // 对话占比 0-1
}

// DefaultStyle 默认风格
func DefaultStyle() StyleConfig {
	return StyleConfig{
		Voice:        "third_person_limited", // 第三人称限制视角
		Tense:        "past",                  // 过去时
		Tone:         "neutral",               // 中性基调
		Pacing:       "medium",                // 中等节奏
		DetailLevel:  "medium",                // 中等细节
		DialogueRatio: 0.3,                    // 30%对话
	}
}

// SceneGenerationResult 场景生成结果
type SceneGenerationResult struct {
	ID            string                  `json:"id"`
	Content       string                  `json:"content"`
	WordCount     int                     `json:"word_count"`
	Metadata      GenerationMetadata      `json:"metadata"`
	StateUpdates  models.StateUpdates     `json:"state_updates"`
}

// GenerationMetadata 生成元数据
type GenerationMetadata struct {
	POVCharacter   string    `json:"pov_character"`
	Tone           string    `json:"tone"`
	Style          string    `json:"style"`
	GeneratedAt    time.Time `json:"generated_at"`
	TokensUsed     int       `json:"tokens_used"`
	RetryCount     int       `json:"retry_count"`
}

// Writer 写作器
type Writer struct {
	db      db.Database
	cfg     *config.Config
	client  *llm.Client
	mapping *config.ModuleMapping
}

// New 创建写作器
func New() (*Writer, error) {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	client, mapping, err := llm.NewClientForModule("writer_scene")
	if err != nil {
		return nil, fmt.Errorf("创建LLM客户端失败: %w", err)
	}

	return &Writer{
		db:      db.Get(),
		cfg:     cfg,
		client:  client,
		mapping: mapping,
	}, nil
}

// GenerateScene 生成场景内容
func (w *Writer) GenerateScene(params GenerateParams) (*SceneGenerationResult, error) {
	startTime := time.Now()

	// 设置默认风格
	if params.Style.Voice == "" {
		params.Style = DefaultStyle()
	}

	// 构建生成提示词
	prompt := w.buildScenePrompt(params)
	systemPrompt := w.buildSystemPrompt(params.Style)

	// 调用LLM生成
	result, err := w.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM调用失败: %w", err)
	}

	// 解析结果
	generated := &GeneratedScene{}
	if err := json.Unmarshal([]byte(result), &generated); err != nil {
		// 如果JSON解析失败，尝试提取纯文本
		extracted := extractJSON(result)
		if err := json.Unmarshal([]byte(extracted), &generated); err != nil {
			// 如果还是失败，将整个结果作为内容
			generated = &GeneratedScene{
				Content:       result,
				WordCount:     len(strings.Fields(result)),
				Tone:          params.Style.Tone,
				POVCharacter:  params.Instruction.POVCharacter,
				StateChanges:  models.StateUpdates{},
			}
		}
	}

	// 创建输出结果
	output := &SceneGenerationResult{
		ID:        db.GenerateID("scene"),
		Content:   generated.Content,
		WordCount: generated.WordCount,
		Metadata: GenerationMetadata{
			POVCharacter: generated.POVCharacter,
			Tone:        generated.Tone,
			Style:       params.Style.Voice,
			GeneratedAt: startTime,
			TokensUsed:  len([]rune(result)),
			RetryCount:  0,
		},
		StateUpdates: generated.StateChanges,
	}

	// 保存到数据库
	sceneOutput := &models.SceneOutput{
		ID:          output.ID,
		BlueprintID: params.BlueprintID,
		Chapter:     params.Chapter,
		Scene:       params.Scene,
		Content:     output.Content,
		WordCount:   output.WordCount,
		CreatedAt:   startTime,
		POVCharacter: output.Metadata.POVCharacter,
		Tone:        output.Metadata.Tone,
		Style:       output.Metadata.Style,
		StateUpdates: output.StateUpdates,
	}

	if err := w.db.SaveScene(sceneOutput); err != nil {
		return nil, fmt.Errorf("保存场景失败: %w", err)
	}

	return output, nil
}

// GeneratedScene LLM生成的场景结构
type GeneratedScene struct {
	Content       string              `json:"content"`
	WordCount     int                 `json:"word_count"`
	Tone          string              `json:"tone"`
	POVCharacter  string              `json:"pov_character"`
	StateChanges  models.StateUpdates `json:"state_changes"`
	Segments      []TextSegment       `json:"segments,omitempty"`
}

// TextSegment 文本段落
type TextSegment struct {
	Type       string `json:"type"`        // dialogue, action, description, inner_thought
	Content    string `json:"content"`
	Character  string `json:"character,omitempty"` // 对话时的说话者
	Emotion    string `json:"emotion,omitempty"`
}

// buildScenePrompt 构建场景生成提示词
func (w *Writer) buildScenePrompt(params GenerateParams) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("# 场景生成任务\n\n"))
	prompt.WriteString(fmt.Sprintf("## 场景信息\n"))
	prompt.WriteString(fmt.Sprintf("- 章节: 第%d章\n", params.Chapter))
	prompt.WriteString(fmt.Sprintf("- 场景: 第%d个场景\n", params.Scene))
	prompt.WriteString(fmt.Sprintf("- 目的: %s\n", params.Instruction.Purpose))
	prompt.WriteString(fmt.Sprintf("- 地点: %s\n", params.Instruction.Location))
	prompt.WriteString(fmt.Sprintf("- 氛围: %s\n", params.Instruction.Mood))
	prompt.WriteString(fmt.Sprintf("- 预期长度: %d 字\n\n", params.Instruction.ExpectedLength))

	// 前情摘要
	if params.PreviousSummary != "" {
		prompt.WriteString(fmt.Sprintf("## 前情摘要\n%s\n\n", params.PreviousSummary))
	}

	// 角色信息
	prompt.WriteString(fmt.Sprintf("## 出场角色\n"))
	if len(params.Instruction.Characters) > 0 {
		for _, charID := range params.Instruction.Characters {
			if charCtx, exists := params.CharacterStates[charID]; exists {
				prompt.WriteString(fmt.Sprintf("- %s: 当前情绪=%s\n", charCtx.Name, charCtx.CurrentEmotion))
			}
		}
	}
	prompt.WriteString("\n")

	// 场景动作
	if params.Instruction.Action != "" {
		prompt.WriteString(fmt.Sprintf("## 场景动作\n%s\n\n", params.Instruction.Action))
	}

	// 对话焦点
	if params.Instruction.DialogueFocus != "" {
		prompt.WriteString(fmt.Sprintf("## 对话焦点\n%s\n\n", params.Instruction.DialogueFocus))
	}

	// 风格要求
	prompt.WriteString(fmt.Sprintf("## 风格要求\n"))
	prompt.WriteString(fmt.Sprintf("- 叙述视角: %s\n", voiceDescription(params.Style.Voice)))
	prompt.WriteString(fmt.Sprintf("- 基调: %s\n", params.Style.Tone))
	prompt.WriteString(fmt.Sprintf("- 节奏: %s\n", pacingDescription(params.Style.Pacing)))
	prompt.WriteString(fmt.Sprintf("- 对话占比: %.0f%%\n\n", params.Style.DialogueRatio*100))

	// 世界背景信息
	if params.WorldContext != nil {
		prompt.WriteString(fmt.Sprintf("## 世界背景\n"))
		prompt.WriteString(fmt.Sprintf("- 世界类型: %s\n", params.WorldContext.Type))
		if len(params.WorldContext.Geography.Regions) > 0 {
			prompt.WriteString(fmt.Sprintf("- 主要区域: %s\n", getRegionNames(params.WorldContext.Geography.Regions)))
		}
		prompt.WriteString("\n")
	}

	// 输出格式要求
	prompt.WriteString(fmt.Sprintf("# 输出要求\n\n"))
	prompt.WriteString(fmt.Sprintf("请根据以上信息生成场景内容，要求：\n"))
	prompt.WriteString(fmt.Sprintf("1. 符合场景目的和动作要求\n"))
	prompt.WriteString(fmt.Sprintf("2. 保持角色性格和情绪一致\n"))
	prompt.WriteString(fmt.Sprintf("3. 对话自然生动，符合角色身份\n"))
	prompt.WriteString(fmt.Sprintf("4. 描写细致但不冗余\n"))
	prompt.WriteString(fmt.Sprintf("5. 控制在预期字数附近\n\n"))

	prompt.WriteString(fmt.Sprintf("# 输出格式（JSON）\n"))
	prompt.WriteString(fmt.Sprintf("{\n"))
	prompt.WriteString(fmt.Sprintf("  \"content\": \"生成的场景文本...\",\n"))
	prompt.WriteString(fmt.Sprintf("  \"word_count\": 实际字数,\n"))
	prompt.WriteString(fmt.Sprintf("  \"tone\": \"%s\",\n", params.Style.Tone))
	prompt.WriteString(fmt.Sprintf("  \"pov_character\": \"%s\",\n", params.Instruction.POVCharacter))
	prompt.WriteString(fmt.Sprintf("  \"state_changes\": {\n"))
	prompt.WriteString(fmt.Sprintf("    \"characters\": [],\n"))
	prompt.WriteString(fmt.Sprintf("    \"plot_progress\": \"情节进展描述\"\n"))
	prompt.WriteString(fmt.Sprintf("  }\n"))
	prompt.WriteString(fmt.Sprintf("}\n\n"))
	prompt.WriteString(fmt.Sprintf("只返回JSON，不要包含其他内容。"))

	return prompt.String()
}

// buildSystemPrompt 构建系统提示词
func (w *Writer) buildSystemPrompt(style StyleConfig) string {
	return fmt.Sprintf(`你是一位专业小说作家，擅长创作引人入胜的叙事内容。

# 写作原则
1. 展示而非讲述（Show, Don't Tell）
2. 通过动作和对话展现角色性格
3. 保持叙事连贯性和节奏感
4. 注意感官描写的层次感
5. 对话要推动情节或揭示角色

# 风格要求
- 叙述视角: %s
- 基调: %s
- 节奏: %s

请根据用户提供的场景指令生成高质量的小说文本。`,
		voiceDescription(style.Voice),
		style.Tone,
		pacingDescription(style.Pacing),
	)
}

// voiceDescription 视角描述
func voiceDescription(voice string) string {
	descriptions := map[string]string{
		"first_person":         "第一人称（我）- 亲身经历，情感直接",
		"third_person_limited": "第三人称限制视角 - 跟随单一角色视角",
		"third_person_omniscient": "第三人称全知视角 - 可展现多人视角",
		"second_person":        "第二人称（你）- 沉浸式体验",
	}
	if desc, ok := descriptions[voice]; ok {
		return desc
	}
	return "第三人称限制视角"
}

// pacingDescription 节奏描述
func pacingDescription(pacing string) string {
	descriptions := map[string]string{
		"slow":    "慢节奏 - 细致描写，心理刻画丰富",
		"medium":  "中等节奏 - 叙述与动作平衡",
		"fast":    "快节奏 - 紧凑有力，动作密集",
	}
	if desc, ok := descriptions[pacing]; ok {
		return desc
	}
	return "中等节奏"
}

// getRegionNames 获取区域名称列表
func getRegionNames(regions []models.Region) string {
	names := make([]string, 0, len(regions))
	for _, r := range regions {
		names = append(names, r.Name)
	}
	return strings.Join(names, "、")
}

// callWithRetry 调用LLM并重试
func (w *Writer) callWithRetry(prompt, systemPrompt string) (string, error) {
	retryConfig := w.cfg.System.Retry
	maxAttempts := retryConfig.MaxAttempts
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := w.client.GenerateJSONWithParams(
			prompt,
			systemPrompt,
			w.mapping.Temperature,
			w.mapping.MaxTokens,
		)

		if err == nil {
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("序列化结果失败: %w", err)
			}
			return string(jsonBytes), nil
		}

		lastErr = err

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

// extractJSON 从文本中提取JSON
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
