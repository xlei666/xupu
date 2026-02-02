// Package writer 描写生成器
// 负责生成各种类型的场景描写
package writer

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DescriptionType 描写类型
type DescriptionType string

const (
	DescEnvironment DescriptionType = "environment" // 环境描写
	DescCharacter  DescriptionType = "character"   // 人物描写
	DescAction     DescriptionType = "action"      // 动作描写
	DescEmotion    DescriptionType = "emotion"     // 情绪描写
	DescSensory    DescriptionType = "sensory"     // 感官描写
	DescAtmosphere DescriptionType = "atmosphere" // 氛围描写
)

// DescriptionParams 描写生成参数
type DescriptionParams struct {
	Type        DescriptionType `json:"type"`
	Subject     string          `json:"subject"`      // 描写对象
	Purpose     string          `json:"purpose"`      // 描写目的
	Mood        string          `json:"mood"`         // 情绪基调
	POV         string          `json:"pov"`          // 视角人物
	Context     string          `json:"context"`      // 上下文信息
	Sensory     SensoryFocus    `json:"sensory"`      // 感官侧重
	DetailLevel string          `json:"detail_level"` // 细节程度
	Length      int             `json:"length"`       // 期望字数
}

// SensoryFocus 感官侧重
type SensoryFocus struct {
	Visual bool `json:"visual"` // 视觉
	Audio  bool `json:"audio"`  // 听觉
	Olfactory bool `json:"olfactory"` // 嗅觉
	Gustatory bool `json:"gustatory"` // 味觉
	Tactile bool `json:"tactile"` // 触觉
	Proprioceptive bool `json:"proprioceptive"` // 本体觉
}

// DescriptionOutput 描写输出
type DescriptionOutput struct {
	Type        DescriptionType `json:"type"`
	Content     string          `json:"content"`
	WordCount   int             `json:"word_count"`
	Techniques  []string        `json:"techniques,omitempty"` // 使用的技巧
	SensoryUsed []string        `json:"sensory_used,omitempty"` // 使用的感官
}

// GenerateDescription 生成描写
func (w *Writer) GenerateDescription(params DescriptionParams) (*DescriptionOutput, error) {
	prompt := w.buildDescriptionPrompt(params)
	systemPrompt := w.buildDescriptionSystemPrompt(params.Type)

	result, err := w.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM调用失败: %w", err)
	}

	// 解析结果
	var output DescriptionOutput
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		extracted := extractJSON(result)
		if err := json.Unmarshal([]byte(extracted), &output); err != nil {
			// 如果解析失败，使用原始内容
			output = DescriptionOutput{
				Type:      params.Type,
				Content:   result,
				WordCount: len([]rune(result)),
			}
		}
	}

	return &output, nil
}

// buildDescriptionPrompt 构建描写提示词
func (w *Writer) buildDescriptionPrompt(params DescriptionParams) string {
	var prompt strings.Builder

	// 描写类型标题
	titles := map[DescriptionType]string{
		DescEnvironment: "环境描写",
		DescCharacter:  "人物描写",
		DescAction:     "动作描写",
		DescEmotion:    "情绪描写",
		DescSensory:    "感官描写",
		DescAtmosphere: "氛围描写",
	}

	prompt.WriteString(fmt.Sprintf("# %s生成任务\n\n", titles[params.Type]))

	// 基本信息
	prompt.WriteString("## 描写对象\n")
	prompt.WriteString(fmt.Sprintf("%s\n\n", params.Subject))

	// 描写目的
	prompt.WriteString("## 描写目的\n")
	prompt.WriteString(fmt.Sprintf("%s\n\n", params.Purpose))

	// 情绪基调
	if params.Mood != "" {
		prompt.WriteString("## 情绪基调\n")
		prompt.WriteString(fmt.Sprintf("%s\n\n", params.Mood))
	}

	// 视角
	if params.POV != "" {
		prompt.WriteString("## 视角人物\n")
		prompt.WriteString(fmt.Sprintf("从%s的视角描写\n\n", params.POV))
	}

	// 上下文
	if params.Context != "" {
		prompt.WriteString("## 上下文\n")
		prompt.WriteString(fmt.Sprintf("%s\n\n", params.Context))
	}

	// 感官侧重
	if anySensory(params.Sensory) {
		prompt.WriteString("## 感官侧重\n")
		var senses []string
		if params.Sensory.Visual {
			senses = append(senses, "视觉")
		}
		if params.Sensory.Audio {
			senses = append(senses, "听觉")
		}
		if params.Sensory.Olfactory {
			senses = append(senses, "嗅觉")
		}
		if params.Sensory.Gustatory {
			senses = append(senses, "味觉")
		}
		if params.Sensory.Tactile {
			senses = append(senses, "触觉")
		}
		if params.Sensory.Proprioceptive {
			senses = append(senses, "本体觉")
		}
		prompt.WriteString(fmt.Sprintf("侧重感官: %s\n\n", strings.Join(senses, "、")))
	}

	// 细节程度
	detailDesc := map[string]string{
		"minimal": "精简 - 只保留关键细节",
		"medium":  "适中 - 重要细节充分描绘",
		"rich":    "丰富 - 多层次细节呈现",
	}
	if params.DetailLevel != "" {
		prompt.WriteString("## 细节程度\n")
		prompt.WriteString(fmt.Sprintf("%s\n\n", detailDesc[params.DetailLevel]))
	}

	// 期望长度
	if params.Length > 0 {
		prompt.WriteString(fmt.Sprintf("## 期望长度\n约%d字\n\n", params.Length))
	}

	// 输出格式
	prompt.WriteString("## 输出格式（JSON）\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"type\": \"" + string(params.Type) + "\",\n")
	prompt.WriteString("  \"content\": \"生成的描写内容\",\n")
	prompt.WriteString("  \"word_count\": 实际字数,\n")
	prompt.WriteString("  \"techniques\": [\"使用的技巧1\", \"技巧2\"],\n")
	prompt.WriteString("  \"sensory_used\": [\"视觉\", \"听觉\", ...]\n")
	prompt.WriteString("}\n\n")
	prompt.WriteString("只返回JSON，不要包含其他内容。")

	return prompt.String()
}

// buildDescriptionSystemPrompt 构建描写系统提示词
func (w *Writer) buildDescriptionSystemPrompt(descType DescriptionType) string {
	basePrompt := `你是一位专业小说家，擅长创作细腻生动的描写。

# 描写写作原则

## 核心原则
1. **展示而非讲述** - 通过具体细节让读者体验，而非直接告诉读者
2. **调动感官** - 使用五感描写创造沉浸感
3. **选择性强** - 只描写与情节/情绪相关的细节
4. **动态呈现** - 让描写参与叙事，而非静止的说明书

`

	techniqueTips := map[DescriptionType]string{
		DescEnvironment: `## 环境描写技巧
- 从宏观到微观或从微观到宏观
- 通过环境反映人物情绪（情景交融）
- 注意光线、色彩、温度等氛围元素
- 环境细节应与情节相关`,

		DescCharacter: `## 人物描写技巧
- 避免特征清单，通过动作和对话展现性格
- 注意标志性细节（而非面面俱到）
- 外貌描写揭示内在特质
- 考虑视角人物的观察角度`,

		DescAction: `## 动作描写技巧
- 分解动作为清晰序列
- 使用强有力的动词
- 注意节奏和停顿
- 动作应反映角色性格和情绪状态`,

		DescEmotion: `## 情绪描写技巧
- 优先通过身体反应展现情绪
- 避免直接命名情绪（如"他很愤怒"）
- 使用环境、动作、对话暗示情绪
- 注意情绪的层次和变化`,

		DescSensory: `## 感官描写技巧
- 超越视觉，运用所有感官
- 通感：混合感官体验
- 具体的感官细节比抽象形容词有力
- 感官细节应与场景情绪一致`,

		DescAtmosphere: `## 氛围描写技巧
- 通过感官细节创造整体感觉
- 利用环境和人物互动营造氛围
- 注意节奏和语言风格的配合
- 氛围应为场景目的服务`,
	}

	return basePrompt + techniqueTips[descType]
}

// GenerateActionSequence 生成动作序列
func (w *Writer) GenerateActionSequence(params ActionSequenceParams) ([]string, error) {
	prompt := w.buildActionSequencePrompt(params)
	systemPrompt := `你是一位动作描写专家。

请将复杂的动作分解为清晰、连贯的序列。
- 每个动作应该具体可执行
- 注意动作的节奏和强度变化
- 动作序列应该符合人物能力和物理规律
- 加入关键细节使动作生动

返回动作描述数组。`

	result, err := w.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var actions []string
	if err := json.Unmarshal([]byte(result), &actions); err != nil {
		extracted := extractJSON(result)
		json.Unmarshal([]byte(extracted), &actions)
	}

	return actions, nil
}

// ActionSequenceParams 动作序列参数
type ActionSequenceParams struct {
	Character      string   `json:"character"`
	Objective      string   `json:"objective"`      // 动作目标
	Environment    string   `json:"environment"`    // 环境条件
	Constraints    []string `json:"constraints"`    // 限制因素
	EmotionalState string   `json:"emotional_state"` // 情绪状态
	StepCount      int      `json:"step_count"`     // 期望步骤数
}

func (w *Writer) buildActionSequencePrompt(params ActionSequenceParams) string {
	var prompt strings.Builder

	prompt.WriteString("# 动作序列生成任务\n\n")

	prompt.WriteString(fmt.Sprintf("## 角色\n%s\n\n", params.Character))
	prompt.WriteString(fmt.Sprintf("## 目标\n%s\n\n", params.Objective))

	if params.Environment != "" {
		prompt.WriteString(fmt.Sprintf("## 环境条件\n%s\n\n", params.Environment))
	}

	if len(params.Constraints) > 0 {
		prompt.WriteString("## 限制因素\n")
		for _, c := range params.Constraints {
			prompt.WriteString(fmt.Sprintf("- %s\n", c))
		}
		prompt.WriteString("\n")
	}

	if params.EmotionalState != "" {
		prompt.WriteString(fmt.Sprintf("## 情绪状态\n%s\n\n", params.EmotionalState))
	}

	if params.StepCount > 0 {
		prompt.WriteString(fmt.Sprintf("## 要求\n分解为约%d个清晰步骤\n\n", params.StepCount))
	}

	prompt.WriteString("## 输出格式\n")
	prompt.WriteString("返回动作描述数组（JSON）：\n")
	prompt.WriteString("[\"步骤1描述\", \"步骤2描述\", ...]\n")

	return prompt.String()
}

// anySensory 检查是否有任何感官选项
func anySensory(s SensoryFocus) bool {
	return s.Visual || s.Audio || s.Olfactory || s.Gustatory || s.Tactile || s.Proprioceptive
}
