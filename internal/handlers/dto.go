// Package handlers HTTP处理器和DTO
package handlers

import "github.com/xlei/xupu/internal/models"

// ============================================
// 请求 DTO
// ============================================

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Mode        string          `json:"mode" binding:"required,oneof=planning intervention random story_core short script assisted workflow"`
	Params      *CreationParams `json:"params"` // 可选：AI创作参数
}

// CreationParams 创作参数
type CreationParams struct {
	// 世界参数
	WorldName  string `json:"world_name"`
	WorldType  string `json:"world_type" binding:"required,oneof=fantasy scifi historical urban wuxia xianxia mixed"`
	WorldTheme string `json:"world_theme"`
	WorldScale string `json:"world_scale" binding:"required,oneof=village city nation continent planet universe"`
	WorldStyle string `json:"world_style"`

	// 故事参数
	StoryType    string `json:"story_type" binding:"required"`
	Theme        string `json:"theme" binding:"required"`
	Protagonist  string `json:"protagonist" binding:"required"`
	Length       string `json:"length" binding:"required,oneof=short medium long"`
	ChapterCount int    `json:"chapter_count" binding:"min=1,max=100"`
	Structure    string `json:"structure" binding:"oneof=three_act heros_journey save_the_cat kishotenketsu freytag_pyramid"`

	// 生成选项
	Options GenerationOptions `json:"options"`
}

// GenerationOptions 生成选项
type GenerationOptions struct {
	SkipWorldBuild      bool   `json:"skip_world_build"`
	ExistingWorldID     string `json:"existing_world_id"`
	SkipNarrative       bool   `json:"skip_narrative"`
	ExistingBlueprintID string `json:"existing_blueprint_id"`
	GenerateContent     bool   `json:"generate_content"`
	StartChapter        int    `json:"start_chapter" binding:"min=1"`
	EndChapter          int    `json:"end_chapter" binding:"min=1"`
	Style               string `json:"style"`
}

// GenerateChapterRequest 生成章节请求
type GenerateChapterRequest struct {
	Regenerate bool `json:"regenerate"`
}

// InterveneRequest 干预请求
type InterveneRequest struct {
	Type    string `json:"type" binding:"required,oneof=world narrative chapter"`
	Content string `json:"content" binding:"required"`
}

// CreateWorldRequest 创建世界请求
type CreateWorldRequest struct {
	Name  string `json:"name" binding:"required"`
	Type  string `json:"type" binding:"required,oneof=fantasy scifi historical urban wuxia xianxia mixed"`
	Scale string `json:"scale" binding:"required,oneof=village city nation continent planet universe"`
	Theme string `json:"theme" binding:"required"`
	Style string `json:"style"`
}

// CreateBlueprintRequest 创建蓝图请求
type CreateBlueprintRequest struct {
	ProjectID    string `json:"project_id"` // Optional, to link immediately
	WorldID      string `json:"world_id" binding:"required"`
	StoryType    string `json:"story_type" binding:"required"`
	Theme        string `json:"theme" binding:"required"`
	Protagonist  string `json:"protagonist" binding:"required"`
	Length       string `json:"length" binding:"required,oneof=short medium long"`
	ChapterCount int    `json:"chapter_count" binding:"min=1,max=100"`
	Structure    string `json:"structure" binding:"oneof=three_act heros_journey save_the_cat kishotenketsu freytag_pyramid"`
}

// ============================================
// 响应 DTO
// ============================================

// APIResponse 通用API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ProjectResponse 项目响应
type ProjectResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Mode        string  `json:"mode"`
	Status      string  `json:"status"`
	Progress    float64 `json:"progress"`
	WorldID     string  `json:"world_id"`
	NarrativeID string  `json:"narrative_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// WorldResponse 世界响应
type WorldResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Scale           string `json:"scale"`
	Style           string `json:"style"`
	CoreQuestion    string `json:"core_question"`
	HighestGood     string `json:"highest_good"`
	UltimateEvil    string `json:"ultimate_evil"`
	SocialConflicts int    `json:"social_conflicts_count"`
	RegionCount     int    `json:"region_count"`
	RaceCount       int    `json:"race_count"`
	CreatedAt       string `json:"created_at"`
}

// BlueprintResponse 蓝图响应
type BlueprintResponse struct {
	ID            string               `json:"id"`
	WorldID       string               `json:"world_id"`
	StructureType string               `json:"structure_type"`
	ChapterCount  int                  `json:"chapter_count"`
	SceneCount    int                  `json:"scene_count"`
	CoreTheme     string               `json:"core_theme"`
	CharacterArcs int                  `json:"character_arcs_count"`
	StoryOutline  models.StoryOutline  `json:"story_outline"`
	ChapterPlans  []models.ChapterPlan `json:"chapter_plans"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`
}

// CreateChapterRequest 创建章节请求
type CreateChapterRequest struct {
	Title      string `json:"title" binding:"required"`
	ChapterNum int    `json:"chapter_num"`
}

// UpdateChapterRequest 更新章节请求
type UpdateChapterRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status" binding:"omitempty,oneof=draft completed"`
}

// ChapterResponse 章节响应
type ChapterResponse struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	ChapterNum  int    `json:"chapter_num"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	WordCount   int    `json:"word_count"`
	AIWordCount int    `json:"ai_generated_word_count"`
	Status      string `json:"status"`
	GeneratedAt string `json:"generated_at,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ReorderChaptersRequest 重新排序章节请求
type ReorderChaptersRequest struct {
	ChapterIDs []string `json:"chapter_ids" binding:"required"`
}

// ProgressResponse 进度响应
type ProgressResponse struct {
	ProjectID          string  `json:"project_id"`
	ProjectName        string  `json:"project_name"`
	Status             string  `json:"status"`
	Progress           float64 `json:"progress"`
	CurrentStage       string  `json:"current_stage"`
	WorldCompleted     bool    `json:"world_completed"`
	NarrativeCompleted bool    `json:"narrative_completed"`
	TotalChapters      int     `json:"total_chapters"`
	TotalScenes        int     `json:"total_scenes"`
	GeneratedScenes    int     `json:"generated_scenes"`
	WordCount          int     `json:"word_count"`
	CompletionPercent  float64 `json:"completion_percent"`
}

// ============================================
// 转换函数
// ============================================

// toProjectResponse 转换项目响应
func toProjectResponse(p *models.Project) ProjectResponse {
	return ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Mode:        string(p.Mode),
		Status:      string(p.Status),
		Progress:    p.Progress,
		WorldID:     p.WorldID,
		NarrativeID: p.NarrativeID,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// toBlueprintResponse 转换蓝图响应
func toBlueprintResponse(b *models.NarrativeBlueprint) BlueprintResponse {
	characterArcsCount := 0
	if b.CharacterArcs != nil {
		characterArcsCount = len(b.CharacterArcs)
	}

	return BlueprintResponse{
		ID:            b.ID,
		WorldID:       b.WorldID,
		StructureType: b.StoryOutline.StructureType,
		ChapterCount:  len(b.ChapterPlans),
		SceneCount:    len(b.Scenes),
		CoreTheme:     b.ThemePlan.CoreTheme,
		CharacterArcs: characterArcsCount,
		StoryOutline:  b.StoryOutline,
		ChapterPlans:  b.ChapterPlans,
		CreatedAt:     b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// successResponse 成功响应
func successResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

// errorResponse 错误响应
func errorResponse(code, message string, details string) APIResponse {
	return APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// toChapterResponse 转换章节响应
func toChapterResponse(c *models.Chapter) ChapterResponse {
	resp := ChapterResponse{
		ID:          c.ID,
		ProjectID:   c.ProjectID,
		ChapterNum:  c.ChapterNum,
		Title:       c.Title,
		Content:     c.Content,
		WordCount:   c.WordCount,
		AIWordCount: c.AIWordCount,
		Status:      string(c.Status),
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if c.GeneratedAt != nil {
		resp.GeneratedAt = c.GeneratedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return resp
}

// ============================================
// 叙事节点相关
// ============================================

// NodeTreeItem 节点树项
type NodeTreeItem struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	NodeType    models.NodeType   `json:"node_type"`
	NodeStatus  models.NodeStatus `json:"node_status"`
	NodeLevel   int               `json:"node_level"`
	NodeOrder   int               `json:"node_order"`
	Description string            `json:"description"`
	Children    []NodeTreeItem    `json:"children"`
}

// NodeResponse 节点响应
type NodeResponse struct {
	ID               string              `json:"id"`
	ProjectID        string              `json:"project_id"`
	WorldID          string              `json:"world_id"`
	ChapterID        *string             `json:"chapter_id,omitempty"`
	ParentID         *string             `json:"parent_id,omitempty"`
	NodeLevel        int                 `json:"node_level"`
	NodeOrder        int                 `json:"node_order"`
	NodeType         models.NodeType     `json:"node_type"`
	NodeStatus       models.NodeStatus   `json:"node_status"`
	Title            string              `json:"title"`
	Description      string              `json:"description"`
	Content          string              `json:"content"`
	Metadata         models.NodeMetadata `json:"metadata"`
	Branches         []models.NodeBranch `json:"branches"`
	SelectedBranchID *string             `json:"selected_branch_id,omitempty"`
	CreatedAt        string              `json:"created_at"`
	UpdatedAt        string              `json:"updated_at"`
}

// MappingResponse 映射关系响应
type MappingResponse struct {
	ID            string               `json:"id"`
	ProjectID     string               `json:"project_id"`
	ChapterID     string               `json:"chapter_id"`
	NodeID        string               `json:"node_id"`
	MappingType   models.MappingType   `json:"mapping_type"`
	Sequence      int                  `json:"sequence"`
	MergeStrategy models.MergeStrategy `json:"merge_strategy"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`
}

// toNodeResponse 转换节点响应
func toNodeResponse(node *models.NarrativeNode) NodeResponse {
	return NodeResponse{
		ID:               node.ID,
		ProjectID:        node.ProjectID,
		WorldID:          node.WorldID,
		ChapterID:        node.ChapterID,
		ParentID:         node.ParentID,
		NodeLevel:        node.NodeLevel,
		NodeOrder:        node.NodeOrder,
		NodeType:         node.NodeType,
		NodeStatus:       node.NodeStatus,
		Title:            node.Title,
		Description:      node.Description,
		Content:          node.Content,
		Metadata:         node.Metadata,
		Branches:         node.Branches,
		SelectedBranchID: node.SelectedBranchID,
		CreatedAt:        node.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        node.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// toMappingResponse 转换映射关系响应
func toMappingResponse(mapping *models.NodeChapterMapping) MappingResponse {
	return MappingResponse{
		ID:            mapping.ID,
		ProjectID:     mapping.ProjectID,
		ChapterID:     mapping.ChapterID,
		NodeID:        mapping.NodeID,
		MappingType:   mapping.MappingType,
		Sequence:      mapping.Sequence,
		MergeStrategy: mapping.MergeStrategy,
		CreatedAt:     mapping.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     mapping.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
