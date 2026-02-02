// Package writer 风格一致性控制
// 确保生成的文本保持统一的风格和声音
package writer

import (
	"fmt"
	"strings"
)

// StyleProfile 风格配置文件
type StyleProfile struct {
	Name          string            `json:"name"`
	Voice         VoiceProfile      `json:"voice"`
	Language      LanguageProfile   `json:"language"`
	Pacing        PacingProfile     `json:"pacing"`
	Tone          ToneProfile       `json:"tone"`
	ShowDontTell  float64           `json:"show_dont_tell"` // 0-1，越大越展示
	DialogueRatio float64           `json:"dialogue_ratio"` // 0-1
	SentenceStyle SentenceProfile   `json:"sentence_style"`
}

// VoiceProfile 叙述声音
type VoiceProfile struct {
	Type         string   `json:"type"` // first_person, third_limited, third_omniscient
	POVCharacter string   `json:"pov_character,omitempty"`
	Distance     string   `json:"distance"` // close, medium, distant
	Temporal     string   `json:"temporal"` // past, present
}

// LanguageProfile 语言风格
type LanguageProfile struct {
	Vocabulary   string   `json:"vocabulary"`   // simple, moderate, sophisticated
	SentenceLength string `json:"sentence_length"` // short, medium, long, varied
	Figurative   float64  `json:"figurative"`  // 0-1，比喻和意象使用程度
	Formality    int      `json:"formality"`   // 0-10，正式程度
	IdiomUse     []string `json:"idiom_use"`   // 习惯用语、口头禅
}

// PacingProfile 节奏控制
type PacingProfile struct {
	Overall      string  `json:"overall"`      // slow, medium, fast
	ActionScenes string  `json:"action_scenes"` // 快节奏场景
	SceneBreaks  bool    `json:"scene_breaks"` // 是否使用场景分割
	PauseUse     float64 `json:"pause_use"`    // 停顿使用频率
}

// ToneProfile 基调控制
type ToneProfile struct {
	Primary      string   `json:"primary"`       // 主要基调
	Secondary    []string `json:"secondary"`     // 次要基调
	MoodProgress bool     `json:"mood_progress"` // 基调是否随故事变化
}

// SentenceProfile 句式风格
type SentenceProfile struct {
	Complexity   string   `json:"complexity"`   // simple, compound, complex, varied
	Pattern      string   `json:"pattern"`      // repetitive, varied, fragmented
	Length       string   `json:"length"`       // short, medium, long
	Parallelism  bool     `json:"parallelism"`  // 是否使用排比
}

// StyleConsistency 风格一致性检查器
type StyleConsistency struct {
	profile *StyleProfile
	samples []string // 之前的文本样本，用于学习风格
}

// NewStyleConsistency 创建风格一致性检查器
func NewStyleConsistency(profile *StyleProfile) *StyleConsistency {
	return &StyleConsistency{
		profile: profile,
		samples: make([]string, 0),
	}
}

// LearnStyle 从样本学习风格
func (sc *StyleConsistency) LearnStyle(samples []string) {
	sc.samples = append(sc.samples, samples...)
	// 限制样本数量
	if len(sc.samples) > 10 {
		sc.samples = sc.samples[len(sc.samples)-10:]
	}
}

// CheckAndFix 检查并修正风格
func (sc *StyleConsistency) CheckAndFix(text string) (string, error) {
	// 简化版本：返回原始文本
	// 完整实现会分析文本特征并与配置文件对比
	return text, nil
}

// PredefinedStyles 预定义风格
var PredefinedStyles = map[string]*StyleProfile{
	"classic_chinese": {
		Name: "经典中文小说",
		Voice: VoiceProfile{
			Type:     "third_limited",
			Distance: "medium",
			Temporal: "past",
		},
		Language: LanguageProfile{
			Vocabulary:     "moderate",
			SentenceLength: "varied",
			Figurative:    0.4,
			Formality:     5,
		},
		Pacing: PacingProfile{
			Overall:      "medium",
			ActionScenes: "fast",
			SceneBreaks:  true,
		},
		Tone: ToneProfile{
			Primary: "neutral",
		},
		SentenceStyle: SentenceProfile{
			Complexity:  "varied",
			Pattern:     "varied",
			Parallelism: false,
		},
	},
	"modern_fast": {
		Name: "现代快节奏",
		Voice: VoiceProfile{
			Type:     "third_limited",
			Distance: "close",
			Temporal: "present",
		},
		Language: LanguageProfile{
			Vocabulary:     "simple",
			SentenceLength: "short",
			Figurative:    0.2,
			Formality:     2,
		},
		Pacing: PacingProfile{
			Overall:      "fast",
			ActionScenes: "fast",
			SceneBreaks:  true,
		},
		Tone: ToneProfile{
			Primary: "urgent",
		},
		SentenceStyle: SentenceProfile{
			Complexity:  "simple",
			Pattern:     "fragmented",
			Parallelism: false,
		},
		ShowDontTell:  0.8,
		DialogueRatio: 0.4,
	},
	"literary_dense": {
		Name: "文学密集型",
		Voice: VoiceProfile{
			Type:     "third_omniscient",
			Distance: "distant",
			Temporal: "past",
		},
		Language: LanguageProfile{
			Vocabulary:     "sophisticated",
			SentenceLength: "long",
			Figurative:    0.7,
			Formality:     8,
		},
		Pacing: PacingProfile{
			Overall:      "slow",
			ActionScenes: "medium",
			SceneBreaks:  false,
		},
		Tone: ToneProfile{
			Primary:   "contemplative",
			Secondary: []string{"melancholic", "philosophical"},
		},
		SentenceStyle: SentenceProfile{
			Complexity:  "complex",
			Pattern:     "varied",
			Parallelism: true,
		},
		ShowDontTell:  0.6,
		DialogueRatio: 0.2,
	},
}

// GetStyle 获取预定义风格
func GetStyle(name string) (*StyleProfile, bool) {
	style, ok := PredefinedStyles[name]
	return style, ok
}

// AnalyzeTextStyle 分析文本风格
func AnalyzeTextStyle(text string) *StyleAnalysis {
	sentences := splitSentences(text)
	avgLength := averageSentenceLength(sentences)

	analysis := &StyleAnalysis{
		SentenceCount:  len(sentences),
		AvgSentenceLen: avgLength,
		WordCount:      len(strings.Fields(text)),
	}

	// 分析句子长度分布
	short, medium, long := 0, 0, 0
	for _, s := range sentences {
		length := len([]rune(s))
		if length < 20 {
			short++
		} else if length < 50 {
			medium++
		} else {
			long++
		}
	}
	analysis.SentenceLengthDist = map[string]int{
			"short": short,
			"medium": medium,
			"long":  long,
	}

	// 分析对话占比
	quoteCount := strings.Count(text, "\"")
	if quoteCount > 0 {
		analysis.DialogueRatio = float64(quoteCount/2) / float64(len(sentences))
	}

	return analysis
}

// StyleAnalysis 风格分析结果
type StyleAnalysis struct {
	SentenceCount      int               `json:"sentence_count"`
	AvgSentenceLen     float64           `json:"avg_sentence_len"`
	WordCount          int               `json:"word_count"`
	SentenceLengthDist map[string]int    `json:"sentence_length_dist"`
	DialogueRatio      float64           `json:"dialogue_ratio"`
	VocabularyDiversity float64          `json:"vocabulary_diversity"`
}

// ApplyStyle 应用风格到生成参数
func ApplyStyle(params GenerateParams, styleName string) GenerateParams {
	if style, ok := GetStyle(styleName); ok {
		params.Style = StyleConfig{
			Voice:        style.Voice.Type,
			Tense:        style.Voice.Temporal,
			Tone:         style.Tone.Primary,
			Pacing:       style.Pacing.Overall,
			DetailLevel:  map[string]string{"slow": "rich", "fast": "minimal", "medium": "medium"}[style.Pacing.Overall],
			DialogueRatio: style.DialogueRatio,
		}
	}
	return params
}

// BuildStylePrompt 构建风格提示词
func BuildStylePrompt(profile *StyleProfile) string {
	var prompt strings.Builder

	prompt.WriteString("# 风格要求\n\n")

	prompt.WriteString("## 叙述声音\n")
	prompt.WriteString(fmt.Sprintf("- 视角类型: %s\n", profile.Voice.Type))
	if profile.Voice.POVCharacter != "" {
		prompt.WriteString(fmt.Sprintf("- 视角人物: %s\n", profile.Voice.POVCharacter))
	}
	prompt.WriteString(fmt.Sprintf("- 叙述距离: %s\n", profile.Voice.Distance))
	prompt.WriteString(fmt.Sprintf("- 时间视角: %s\n\n", profile.Voice.Temporal))

	prompt.WriteString("## 语言风格\n")
	prompt.WriteString(fmt.Sprintf("- 词汇水平: %s\n", profile.Language.Vocabulary))
	prompt.WriteString(fmt.Sprintf("- 句子长度: %s\n", profile.Language.SentenceLength))
	prompt.WriteString(fmt.Sprintf("- 比喻意象程度: %.0f%%\n", profile.Language.Figurative*100))
	prompt.WriteString(fmt.Sprintf("- 正式程度: %d/10\n\n", profile.Language.Formality))

	prompt.WriteString("## 节奏控制\n")
	prompt.WriteString(fmt.Sprintf("- 整体节奏: %s\n", profile.Pacing.Overall))
	prompt.WriteString(fmt.Sprintf("- 动作场景节奏: %s\n", profile.Pacing.ActionScenes))
	if profile.Pacing.SceneBreaks {
		prompt.WriteString("- 使用场景分割\n")
	}
	prompt.WriteString("\n")

	prompt.WriteString("## 基调\n")
	prompt.WriteString(fmt.Sprintf("- 主要基调: %s\n", profile.Tone.Primary))
	if len(profile.Tone.Secondary) > 0 {
		prompt.WriteString(fmt.Sprintf("- 次要基调: %s\n", strings.Join(profile.Tone.Secondary, "、")))
	}
	prompt.WriteString("\n")

	prompt.WriteString("## 句式风格\n")
	prompt.WriteString(fmt.Sprintf("- 复杂度: %s\n", profile.SentenceStyle.Complexity))
	prompt.WriteString(fmt.Sprintf("- 模式: %s\n", profile.SentenceStyle.Pattern))
	if profile.SentenceStyle.Parallelism {
		prompt.WriteString("- 使用排比句式\n")
	}
	prompt.WriteString("\n")

	if profile.ShowDontTell > 0 {
		prompt.WriteString(fmt.Sprintf("## 展示原则\n- 展示而非讲述比例: %.0f%%\n\n", profile.ShowDontTell*100))
	}

	if profile.DialogueRatio > 0 {
		prompt.WriteString(fmt.Sprintf("## 对话占比\n- 约%.0f%%的内容应为对话\n\n", profile.DialogueRatio*100))
	}

	return prompt.String()
}

// 辅助函数
func splitSentences(text string) []string {
	// 简化版本：按句号、问号、感叹号分割
	sentences := make([]string, 0)
	current := ""

	for _, r := range text {
		current += string(r)
		if r == '。' || r == '！' || r == '？' || r == '\n' {
			if strings.TrimSpace(current) != "" {
				sentences = append(sentences, strings.TrimSpace(current))
			}
			current = ""
		}
	}

	if strings.TrimSpace(current) != "" {
		sentences = append(sentences, strings.TrimSpace(current))
	}

	return sentences
}

func averageSentenceLength(sentences []string) float64 {
	if len(sentences) == 0 {
		return 0
	}

	total := 0
	for _, s := range sentences {
		total += len([]rune(s))
	}

	return float64(total) / float64(len(sentences))
}
