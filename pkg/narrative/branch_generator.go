package narrative

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/config"
	"github.com/xlei/xupu/pkg/llm"
)

// BranchGenerationInput 分支生成输入
type BranchGenerationInput struct {
	NodeContent     string              // 节点内容
	NodeDescription string              // 节点描述
	WorldContext    *models.WorldSetting // 世界设定
	BranchCount     int                 // 分支数量（3-5）
	BranchTypes     []models.BranchType // 指定的分支类型（可选）
	Characters      []string            // 涉及的角色
	PrevNodes       []NodeSummary       // 前序节点（用于保持连贯性）
}

// NodeSummary 节点摘要
type NodeSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Metadata string `json:"metadata"` // JSON string
}

// BranchGenerationOutput 分支生成输出
type BranchGenerationOutput struct {
	Branches []models.NodeBranch `json:"branches"`
	Count    int                 `json:"count"`
}

// BranchGenerator 分支生成器
type BranchGenerator struct {
	llmClient *llm.Client
	config    *config.Config
}

// NewBranchGenerator 创建分支生成器
func NewBranchGenerator(llmClient *llm.Client, cfg *config.Config) *BranchGenerator {
	return &BranchGenerator{
		llmClient: llmClient,
		config:    cfg,
	}
}

// GenerateBranches 生成分支选项（核心AI功能）
func (bg *BranchGenerator) GenerateBranches(input BranchGenerationInput) (*BranchGenerationOutput, error) {
	// 默认生成4个分支
	if input.BranchCount < 3 || input.BranchCount > 5 {
		input.BranchCount = 4
	}

	// 如果没有指定分支类型，使用默认类型组合
	if len(input.BranchTypes) == 0 {
		input.BranchTypes = []models.BranchType{
			models.BranchTypeContinuation,
			models.BranchTypePlotTwist,
			models.BranchTypeCharacterDevelopment,
			models.BranchTypeConflictEscalation,
		}
	}

	// 构建提示词
	prompt := bg.buildBranchPrompt(input)

	// 调用LLM生成JSON
	result, err := bg.llmClient.GenerateJSON(prompt, "")
	if err != nil {
		return nil, fmt.Errorf("LLM调用失败: %w", err)
	}

	// 解析响应
	branchesJSON, ok := result["branches"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应格式错误: 缺少branches字段")
	}

	var branches []models.NodeBranch
	branchesData, _ := json.Marshal(branchesJSON)
	if err := json.Unmarshal(branchesData, &branches); err != nil {
		return nil, fmt.Errorf("解析分支数据失败: %w", err)
	}

	// 确保生成了足够数量的分支
	for i := len(branches); i < input.BranchCount; i++ {
		branches = append(branches, models.NodeBranch{
			ID:            uuid.New().String(),
			Title:         fmt.Sprintf("选项 %d", i+1),
			Description:   "待生成",
			ContentPreview: "内容预览...",
			FullContent:   "完整内容...",
			BranchType:    models.BranchTypeContinuation,
			Rationale:     "补充选项",
			CreatedAt:     time.Now(),
		})
	}

	return &BranchGenerationOutput{
		Branches: branches,
		Count:    len(branches),
	}, nil
}

// buildBranchPrompt 构建分支生成提示词
func (bg *BranchGenerator) buildBranchPrompt(input BranchGenerationInput) string {
	// 构建世界设定摘要
	worldSummary := ""
	if input.WorldContext != nil {
		worldSummary = buildWorldSummary(input.WorldContext)
	}

	// 构建前序节点摘要
	prevContext := ""
	if len(input.PrevNodes) > 0 {
		prevContext = "\n前序情节：\n"
		for _, node := range input.PrevNodes {
			prevContext += fmt.Sprintf("- %s: %s\n", node.Title, truncateString(node.Content, 100))
		}
	}

	// 构建分支类型描述
	branchTypesDesc := ""
	branchTypeMap := map[models.BranchType]string{
		models.BranchTypeContinuation:          "延续 - 情节自然发展",
		models.BranchTypePlotTwist:             "转折 - 引入意外变化",
		models.BranchTypeCharacterDevelopment:  "角色发展 - 深化角色弧光",
		models.BranchTypeConflictEscalation:    "冲突升级 - 加剧矛盾",
		models.BranchTypeConflictResolution:    "冲突解决 - 化解矛盾",
		models.BranchTypeForeshadow:           "伏笔 - 埋下后续线索",
	}

	for _, bt := range input.BranchTypes {
		if desc, exists := branchTypeMap[bt]; exists {
			branchTypesDesc += fmt.Sprintf("- %s\n", desc)
		}
	}

	prompt := fmt.Sprintf(`你是一位专业的小说创作助手。请为以下场景节点生成 %d 个不同的发展分支选项。

【当前场景】
标题：%s
描述：%s
内容：%s

【世界设定】
%s

【涉及角色】
%s
%s

【分支类型要求】
请为以下每种类型生成一个分支选项：
%s

【生成要求】
1. 每个分支都要有明确的类型（延续/转折/角色发展/冲突升级/冲突解决/伏笔）
2. 每个分支包含：
   - title: 简短标题（5-10字）
   - description: 详细描述（50-100字）
   - content_preview: 内容预览（100-200字，展示这个分支的写作风格）
   - full_content: 完整内容（300-500字的完整场景草稿）
   - branch_type: 分支类型
   - rationale: 选择这个分支的理由（为什么这样发展）
   - expected_outcome: 预期效果
     * plot_progression: 情节如何推进
     * character_changes: 角色的变化（如有）
     * new_conflicts: 新增冲突（如有）
     * resolved_conflicts: 解决的冲突（如有）
     * foreshadow_hints: 伏笔提示（如有）

3. 保持与前序情节的连贯性
4. 符合世界设定
5. 每个分支要有不同的走向，给用户真正的选择

请以JSON格式返回，格式如下：
[
  {
    "id": "uuid",
    "title": "分支标题",
    "description": "详细描述",
    "content_preview": "内容预览...",
    "full_content": "完整内容...",
    "branch_type": "continuation|plot_twist|character_development|conflict_escalation|conflict_resolution|foreshadow",
    "rationale": "选择理由",
    "expected_outcome": {
      "plot_progression": "情节推进",
      "character_changes": ["角色变化"],
      "new_conflicts": ["新冲突"],
      "resolved_conflicts": ["解决的冲突"],
      "foreshadow_hints": ["伏笔"]
    }
  }
]

请开始生成：`,
		input.BranchCount,
		input.NodeDescription,
		input.NodeDescription,
		truncateString(input.NodeContent, 500),
		worldSummary,
		formatCharacters(input.Characters),
		prevContext,
		branchTypesDesc,
	)

	return prompt
}

// buildWorldSummary 构建世界设定摘要
func buildWorldSummary(world *models.WorldSetting) string {
	summary := ""

	if world.Philosophy.CoreQuestion != "" {
		summary += fmt.Sprintf("哲学: %s\n", world.Philosophy.CoreQuestion)
	}

	if world.Worldview.Cosmology.Origin != "" {
		summary += fmt.Sprintf("世界观: %s\n", world.Worldview.Cosmology.Origin)
	}

	if world.Laws.Physics.Gravity != "" {
		summary += fmt.Sprintf("法则: %s\n", world.Laws.Physics.Gravity)
	}

	return truncateString(summary, 500)
}

// formatCharacters 格式化角色列表
func formatCharacters(characters []string) string {
	if len(characters) == 0 {
		return "无特定角色"
	}

	result := ""
	for _, char := range characters {
		result += fmt.Sprintf("- %s\n", char)
	}
	return result
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
