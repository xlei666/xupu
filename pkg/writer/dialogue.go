// Package writer 对话生成器
// 负责生成自然、生动的角色对话
package writer

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DialogueParams 对话生成参数
type DialogueParams struct {
	Characters     []DialogueCharacter `json:"characters"`      // 参与对话的角色
	Context        string              `json:"context"`         // 对话发生的背景
	Purpose        string              `json:"purpose"`         // 对话的目的
	EmotionalTone  string              `json:"emotional_tone"`  // 情感基调
	Subtext        string              `json:"subtext"`         // 潜台词（角色真正想说但没说出口的）
	Relationship   string              `json:"relationship"`    // 角色关系
	DesiredOutcome string              `json:"desired_outcome"` // 期望的对话结果
}

// DialogueCharacter 对话角色
type DialogueCharacter struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Personality  []string `json:"personality"`  // 性格特征
	SpeechStyle  string   `json:"speech_style"` // 说话风格（正式/随意/口吃等）
	CurrentState string   `json:"current_state"` // 当前状态
	HiddenAgenda string   `json:"hidden_agenda"` // 隐藏动机
}

// DialogueLine 单句对话
type DialogueLine struct {
	Character   string `json:"character"`
	Content     string `json:"content"`
	Action      string `json:"action,omitempty"`      // 伴随动作
	InnerThought string `json:"inner_thought,omitempty"` // 内心想法
	Subtext     string `json:"subtext,omitempty"`     // 潜台词
	Tone        string `json:"tone,omitempty"`        // 语气
}

// DialogueSegment 对话片段
type DialogueSegment struct {
	Lines        []DialogueLine `json:"lines"`
	Description  string         `json:"description,omitempty"`  // 场景描述
	Pause        string         `json:"pause,omitempty"`        // 停顿/沉默
	Conflict     bool           `json:"conflict"`               // 是否有冲突
	TurningPoint bool           `json:"turning_point"`          // 是否是转折点
}

// GenerateDialogue 生成对话
func (w *Writer) GenerateDialogue(params DialogueParams) (*DialogueSegment, error) {
	prompt := w.buildDialoguePrompt(params)
	systemPrompt := w.buildDialogueSystemPrompt()

	result, err := w.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM调用失败: %w", err)
	}

	// 解析结果
	var segment DialogueSegment
	if err := json.Unmarshal([]byte(result), &segment); err != nil {
		extracted := extractJSON(result)
		if err := json.Unmarshal([]byte(extracted), &segment); err != nil {
			// 如果解析失败，尝试构建简单的对话片段
			segment = DialogueSegment{
				Lines: []DialogueLine{
					{Character: params.Characters[0].Name, Content: result},
				},
			}
		}
	}

	return &segment, nil
}

// buildDialoguePrompt 构建对话生成提示词
func (w *Writer) buildDialoguePrompt(params DialogueParams) string {
	var prompt strings.Builder

	prompt.WriteString("# 对话生成任务\n\n")

	// 角色信息
	prompt.WriteString("## 参与角色\n")
	for _, char := range params.Characters {
		prompt.WriteString(fmt.Sprintf("### %s\n", char.Name))
		prompt.WriteString(fmt.Sprintf("- 性格: %s\n", strings.Join(char.Personality, "、")))
		prompt.WriteString(fmt.Sprintf("- 说话风格: %s\n", char.SpeechStyle))
		prompt.WriteString(fmt.Sprintf("- 当前状态: %s\n", char.CurrentState))
		if char.HiddenAgenda != "" {
			prompt.WriteString(fmt.Sprintf("- 隐藏动机: %s\n", char.HiddenAgenda))
		}
	}
	prompt.WriteString("\n")

	// 对话背景
	prompt.WriteString(fmt.Sprintf("## 对话背景\n%s\n\n", params.Context))

	// 对话目的
	prompt.WriteString(fmt.Sprintf("## 对话目的\n%s\n\n", params.Purpose))

	// 情感基调
	prompt.WriteString(fmt.Sprintf("## 情感基调\n%s\n\n", params.EmotionalTone))

	// 角色关系
	if params.Relationship != "" {
		prompt.WriteString(fmt.Sprintf("## 角色关系\n%s\n\n", params.Relationship))
	}

	// 潜台词
	if params.Subtext != "" {
		prompt.WriteString(fmt.Sprintf("## 潜台词（角色真正想表达的）\n%s\n\n", params.Subtext))
	}

	// 期望结果
	if params.DesiredOutcome != "" {
		prompt.WriteString(fmt.Sprintf("## 期望结果\n%s\n\n", params.DesiredOutcome))
	}

	// 写作要求
	prompt.WriteString("## 写作要求\n")
	prompt.WriteString("1. 对话要自然流畅，符合角色性格和说话风格\n")
	prompt.WriteString("2. 避免信息倾销（info dump），让信息自然流露\n")
	prompt.WriteString("3. 每句对话都应该：推动情节 / 展现性格 / 制造张力（至少满足其一）\n")
	prompt.WriteString("4. 加入动作、表情、停顿等非语言元素\n")
	prompt.WriteString("5. 利用潜台词增加对话层次\n\n")

	// 输出格式
	prompt.WriteString("## 输出格式（JSON）\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"lines\": [\n")
	prompt.WriteString("    {\n")
	prompt.WriteString("      \"character\": \"角色名\",\n")
	prompt.WriteString("      \"content\": \"对话内容\",\n")
	prompt.WriteString("      \"action\": \"伴随动作\",\n")
	prompt.WriteString("      \"inner_thought\": \"内心想法\",\n")
	prompt.WriteString("      \"tone\": \"语气\"\n")
	prompt.WriteString("    }\n")
	prompt.WriteString("  ],\n")
	prompt.WriteString("  \"description\": \"场景描写\",\n")
	prompt.WriteString("  \"conflict\": true/false,\n")
	prompt.WriteString("  \"turning_point\": true/false\n")
	prompt.WriteString("}\n\n")
	prompt.WriteString("只返回JSON，不要包含其他内容。")

	return prompt.String()
}

// buildDialogueSystemPrompt 构建对话系统提示词
func (w *Writer) buildDialogueSystemPrompt() string {
	return `你是一位专业小说家，擅长创作生动、自然的角色对话。

# 对话写作原则

## 核心原则
1. **每句对话都有目的** - 推动情节 / 展现性格 / 制造张力 / 传递信息
2. **角色声音独特** - 每个角色的说话方式应该可区分
3. **避免直接陈述** - 让角色通过对话展现信息，而非直接告诉读者

## 对话技巧

### 言语多样化
- 不同角色有不同的词汇选择、句式结构、说话习惯
- 考虑角色的教育背景、地域、性格、情绪状态

### 动作与对话配合
- 用动作替代部分对话标签
- 动作可以揭示角色真实想法（与话语矛盾时更有张力）

### 潜台词
- 角色说的 ≠ 角色想的
- 通过犹豫、转移话题、顾左右而言他展现内心冲突

### 节奏控制
- 短句：紧张、快速、冲突激烈
- 长句：沉思、解释、情感流动
- 沉默：有时比语言更有力量

## 避免的陷阱
- ❌ 过度使用"他说/她说"
- ❌ 所有角色说话方式一样
- ❌ 对话纯粹用于解释设定
- ❌ 缺乏冲突或张力
- ❌ 过于正式或过于口语化（不符合场景）

请根据用户提供的角色和场景生成高质量对话。`
}

// EnhanceDialogue 增强现有对话
func (w *Writer) EnhanceDialogue(existingLines []DialogueLine, params EnhancementParams) ([]DialogueLine, error) {
	prompt := w.buildEnhancementPrompt(existingLines, params)
	systemPrompt := `你是一位小说编辑，擅长改进对话质量。

请分析提供的对话，并进行以下改进：
1. 增加角色声音的独特性
2. 加入动作和描写
3. 增添潜台词层次
4. 移除冗余或无效对话
5. 保持对话的节奏和张力

返回改进后的对话数组。`

	result, err := w.callWithRetry(prompt, systemPrompt)
	if err != nil {
		return nil, err
	}

	var enhanced []DialogueLine
	if err := json.Unmarshal([]byte(result), &enhanced); err != nil {
		extracted := extractJSON(result)
		json.Unmarshal([]byte(extracted), &enhanced)
	}

	return enhanced, nil
}

// EnhancementParams 增强参数
type EnhancementParams struct {
	Focus          string `json:"focus"`           // 关注点：conflict, character, tension
	AddSubtext     bool   `json:"add_subtext"`     // 是否添加潜台词
	ShowDontTell   bool   `json:"show_dont_tell"`  // 是否应用展示而非讲述
	IncreaseTension bool  `json:"increase_tension"` // 是否增加张力
}

func (w *Writer) buildEnhancementPrompt(lines []DialogueLine, params EnhancementParams) string {
	var prompt strings.Builder

	prompt.WriteString("# 对话增强任务\n\n")

	prompt.WriteString("## 原始对话\n")
	for _, line := range lines {
		prompt.WriteString(fmt.Sprintf("- %s: %s\n", line.Character, line.Content))
	}
	prompt.WriteString("\n")

	prompt.WriteString("## 增强要求\n")
	if params.Focus != "" {
		prompt.WriteString(fmt.Sprintf("- 关注点: %s\n", params.Focus))
	}
	if params.AddSubtext {
		prompt.WriteString("- 添加潜台词层次\n")
	}
	if params.ShowDontTell {
		prompt.WriteString("- 应用「展示而非讲述」原则\n")
	}
	if params.IncreaseTension {
		prompt.WriteString("- 增加对话张力\n")
	}
	prompt.WriteString("\n")

	prompt.WriteString("## 输出格式\n")
	prompt.WriteString("返回改进后的对话数组（JSON格式）：\n")
	prompt.WriteString("[\n")
	prompt.WriteString("  {\"character\": \"名字\", \"content\": \"对话\", \"action\": \"动作\"},\n")
	prompt.WriteString("  ...\n")
	prompt.WriteString("]\n")

	return prompt.String()
}
