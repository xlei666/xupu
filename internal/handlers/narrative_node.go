// Package handlers HTTP处理器
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/config"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/llm"
	"github.com/xlei/xupu/pkg/narrative"
)

// NarrativeNodeHandler 叙事节点处理器
type NarrativeNodeHandler struct {
	db              db.Database
	llmClient       *llm.Client
	config          *config.Config
	branchGenerator *narrative.BranchGenerator
}

// NewNarrativeNodeHandler 创建叙事节点处理器
func NewNarrativeNodeHandler(database db.Database, llmClient *llm.Client, cfg *config.Config) *NarrativeNodeHandler {
	return &NarrativeNodeHandler{
		db:              database,
		llmClient:       llmClient,
		config:          cfg,
		branchGenerator: narrative.NewBranchGenerator(llmClient, cfg),
	}
}

// ============================================
// 请求/响应 DTO
// ============================================

// CreateNodeRequest 创建节点请求
type CreateNodeRequest struct {
	ParentID    *string  `json:"parent_id"`
	NodeLevel   int      `json:"node_level"`
	NodeOrder   int      `json:"node_order"`
	NodeType    string   `json:"node_type" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
}

// UpdateNodeRequest 更新节点请求
type UpdateNodeRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Content     *string `json:"content"`
	Metadata    *models.NodeMetadata `json:"metadata"`
}

// GenerateBranchesRequest 生成分支请求
type GenerateBranchesRequest struct {
	BranchCount int      `json:"branch_count"` // 分支数量，3-5
	BranchTypes []string `json:"branch_types"` // 可选：指定分支类型
}

// SelectBranchRequest 选择分支请求
type SelectBranchRequest struct {
	BranchID string `json:"branch_id" binding:"required"`
}

// MergeToChapterRequest 合并到章节请求
type MergeToChapterRequest struct {
	ChapterID     *string              `json:"chapter_id"`     // 可选：为空则创建新章节
	ChapterNum    int                  `json:"chapter_num"`    // 章节号
	ChapterTitle  string              `json:"chapter_title"` // 章节标题
	MappingType   models.MappingType   `json:"mapping_type"`   // 映射类型
	MergeStrategy models.MergeStrategy `json:"merge_strategy"` // 合并策略
	Sequence      int                  `json:"sequence"`       // 顺序
	Position      *int                `json:"position"`       // 插入位置（仅insert策略）
}

// ============================================
// API Handlers
// ============================================

// CreateNode 创建叙事节点
// @Summary 创建叙事节点
// @Description 为项目创建一个新的叙事节点
// @Tags narrative-nodes
// @Accept json
// @Produce json
// @Param id path string true "项目ID"
// @Param request body CreateNodeRequest true "节点信息"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/narrative-nodes [post]
func (h *NarrativeNodeHandler) CreateNode(c *gin.Context) {
	projectID := c.Param("projectId")

	// 验证项目存在
	project, err := h.db.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	var req CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 获取世界设定快照
	var worldStages models.WorldStagesSnapshot
	if project.WorldID != "" {
		world, err := h.db.GetWorld(project.WorldID)
		if err == nil {
			worldStages = models.WorldStagesSnapshot{
				Philosophy:   &world.Philosophy,
				Worldview:    &world.Worldview,
				Laws:         &world.Laws,
				Geography:    &world.Geography,
				Civilization: &world.Civilization,
				Society:      &world.Society,
				History:      &world.History,
				StorySoil:    &world.StorySoil,
			}
		}
	}

	// 创建节点
	node := &models.NarrativeNode{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		WorldID:     project.WorldID,
		ParentID:    req.ParentID,
		NodeLevel:   req.NodeLevel,
		NodeOrder:   req.NodeOrder,
		NodeType:    models.NodeType(req.NodeType),
		NodeStatus:  models.NodeStatusDraft,
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		WorldStages: worldStages,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存
	if err := h.db.SaveNarrativeNode(node); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("SAVE_FAILED", "保存失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(toNodeResponse(node)))
}

// GetNodeTree 获取节点树（层级结构）
// @Summary 获取项目的叙事节点树
// @Description 获取指定项目的所有叙事节点，以树形结构返回
// @Tags narrative-nodes
// @Produce json
// @Param id path string true "项目ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/narrative-nodes [get]
func (h *NarrativeNodeHandler) GetNodeTree(c *gin.Context) {
	projectID := c.Param("projectId")

	// 验证项目存在
	if _, err := h.db.GetProject(projectID); err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	nodes := h.db.ListNarrativeNodesByProject(projectID)

	// 构建树形结构
	tree := h.buildNodeTree(nodes)

	c.JSON(http.StatusOK, successResponse(gin.H{
		"nodes": tree,
		"total": len(nodes),
	}))
}

// buildNodeTree 构建节点树
func (h *NarrativeNodeHandler) buildNodeTree(nodes []*models.NarrativeNode) []NodeTreeItem {
	// 创建ID到节点的映射
	nodeMap := make(map[string]*NodeTreeItem)
	for _, node := range nodes {
		nodeMap[node.ID] = &NodeTreeItem{
			ID:          node.ID,
			Title:       node.Title,
			NodeType:    node.NodeType,
			NodeStatus:  node.NodeStatus,
			NodeLevel:   node.NodeLevel,
			NodeOrder:   node.NodeOrder,
			Description: node.Description,
			Children:    []NodeTreeItem{},
		}
	}

	// 构建树
	rootItems := []NodeTreeItem{}
	for _, node := range nodes {
		item := nodeMap[node.ID]
		if node.ParentID == nil || *node.ParentID == "" {
			rootItems = append(rootItems, *item)
		} else {
			if parent, exists := nodeMap[*node.ParentID]; exists {
				parent.Children = append(parent.Children, *item)
			}
		}
	}

	return rootItems
}

// UpdateNode 更新节点
// @Summary 更新叙事节点
// @Description 更新指定叙事节点的信息
// @Tags narrative-nodes
// @Accept json
// @Produce json
// @Param id path string true "项目ID"
// @Param nodeId path string true "节点ID"
// @Param request body UpdateNodeRequest true "更新信息"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/narrative-nodes/{nodeId} [put]
func (h *NarrativeNodeHandler) UpdateNode(c *gin.Context) {
	projectID := c.Param("projectId")
	nodeID := c.Param("nodeId")

	var req UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 获取节点
	node, err := h.db.GetNarrativeNode(nodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "节点不存在", ""))
		return
	}

	// 验证权限
	if node.ProjectID != projectID {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "无权限", ""))
		return
	}

	// 更新字段
	if req.Title != nil {
		node.Title = *req.Title
	}
	if req.Description != nil {
		node.Description = *req.Description
	}
	if req.Content != nil {
		node.Content = *req.Content
	}
	if req.Metadata != nil {
		node.Metadata = *req.Metadata
	}
	node.UpdatedAt = time.Now()

	// 保存
	if err := h.db.SaveNarrativeNode(node); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("SAVE_FAILED", "保存失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(toNodeResponse(node)))
}

// DeleteNode 删除节点
// @Summary 删除叙事节点
// @Description 删除指定的叙事节点
// @Tags narrative-nodes
// @Produce json
// @Param id path string true "项目ID"
// @Param nodeId path string true "节点ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/narrative-nodes/{nodeId} [delete]
func (h *NarrativeNodeHandler) DeleteNode(c *gin.Context) {
	projectID := c.Param("projectId")
	nodeID := c.Param("nodeId")

	// 获取节点
	node, err := h.db.GetNarrativeNode(nodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "节点不存在", ""))
		return
	}

	// 验证权限
	if node.ProjectID != projectID {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "无权限", ""))
		return
	}

	// 检查是否有子节点
	children := h.db.ListNarrativeNodesByParent(nodeID)
	if len(children) > 0 {
		c.JSON(http.StatusBadRequest, errorResponse("HAS_CHILDREN", "节点包含子节点，无法删除", ""))
		return
	}

	// 删除
	if err := h.db.DeleteNarrativeNode(nodeID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DELETE_FAILED", "删除失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"message": "删除成功",
	}))
}

// GenerateBranches 生成分支（核心功能）
// @Summary 生成分支选项
// @Description 为指定节点生成3-5个AI分支选项
// @Tags narrative-nodes
// @Accept json
// @Produce json
// @Param id path string true "项目ID"
// @Param nodeId path string true "节点ID"
// @Param request body GenerateBranchesRequest true "分支生成参数"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/narrative-nodes/{nodeId}/branches [post]
func (h *NarrativeNodeHandler) GenerateBranches(c *gin.Context) {
	projectID := c.Param("projectId")
	nodeID := c.Param("nodeId")

	var req GenerateBranchesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 获取节点
	node, err := h.db.GetNarrativeNode(nodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "节点不存在", ""))
		return
	}

	// 验证权限
	if node.ProjectID != projectID {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "无权限", ""))
		return
	}

	// 更新节点状态为分支生成中
	node.NodeStatus = models.NodeStatusBranching
	h.db.SaveNarrativeNode(node)

	// 调用分支生成器
	// 获取世界设定
	world, _ := h.db.GetWorld(node.WorldID)

	// 获取前序节点（用于保持连贯性）
	allNodes := h.db.ListNarrativeNodesByProject(projectID)
	var prevNodes []narrative.NodeSummary
	for _, n := range allNodes {
		if n.NodeOrder < node.NodeOrder || (n.NodeOrder == node.NodeOrder && n.ID < node.ID) {
			metadataBytes, _ := json.Marshal(n.Metadata)
			prevNodes = append(prevNodes, narrative.NodeSummary{
				ID:       n.ID,
				Title:    n.Title,
				Content:  n.Content,
				Metadata: string(metadataBytes),
			})
		}
	}

	// 构建分支生成输入
	// 转换分支类型
	branchTypes := make([]models.BranchType, 0, len(req.BranchTypes))
	for _, bt := range req.BranchTypes {
		branchTypes = append(branchTypes, models.BranchType(bt))
	}

	branchInput := narrative.BranchGenerationInput{
		NodeContent:     node.Content,
		NodeDescription: node.Description,
		WorldContext:    world,
		BranchCount:     req.BranchCount,
		BranchTypes:     branchTypes,
		Characters:      node.Metadata.Characters,
		PrevNodes:       prevNodes,
	}

	// 调用AI生成分支
	output, err := h.branchGenerator.GenerateBranches(branchInput)
	if err != nil {
		node.NodeStatus = models.NodeStatusDraft
		h.db.SaveNarrativeNode(node)
		c.JSON(http.StatusInternalServerError, errorResponse("GENERATION_FAILED", "分支生成失败", err.Error()))
		return
	}

	branches := output.Branches

	// 保存分支
	node.Branches = branches
	node.NodeStatus = models.NodeStatusGenerated
	node.UpdatedAt = time.Now()
	h.db.SaveNarrativeNode(node)

	c.JSON(http.StatusOK, successResponse(gin.H{
		"node_id":  nodeID,
		"branches": branches,
		"count":    len(branches),
	}))
}

// SelectBranch 选择分支（设置节点内容）
// @Summary 选择分支
// @Description 选择某个分支作为节点的内容
// @Tags narrative-nodes
// @Param id path string true "项目ID"
// @Param nodeId path string true "节点ID"
// @Param branchId path string true "分支ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/narrative-nodes/{nodeId}/branches/{branchId}/select [post]
func (h *NarrativeNodeHandler) SelectBranch(c *gin.Context) {
	projectID := c.Param("projectId")
	nodeID := c.Param("nodeId")
	branchID := c.Param("branchId")

	// 获取节点
	node, err := h.db.GetNarrativeNode(nodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "节点不存在", ""))
		return
	}

	// 验证权限
	if node.ProjectID != projectID {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "无权限", ""))
		return
	}

	// 查找分支
	var selectedBranch *models.NodeBranch
	for i := range node.Branches {
		if node.Branches[i].ID == branchID {
			selectedBranch = &node.Branches[i]
			break
		}
	}

	if selectedBranch == nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "分支不存在", ""))
		return
	}

	// 设置节点内容
	node.Content = selectedBranch.FullContent
	node.SelectedBranchID = &branchID
	node.NodeStatus = models.NodeStatusGenerated
	node.UpdatedAt = time.Now()

	// 保存
	if err := h.db.SaveNarrativeNode(node); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("SAVE_FAILED", "保存失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(toNodeResponse(node)))
}

// MergeToChapter 合并节点到章节
// @Summary 合并节点到章节
// @Description 将节点内容合并到指定章节
// @Tags narrative-nodes
// @Accept json
// @Produce json
// @Param id path string true "项目ID"
// @Param nodeId path string true "节点ID"
// @Param request body MergeToChapterRequest true "合并参数"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/narrative-nodes/{nodeId}/merge [post]
func (h *NarrativeNodeHandler) MergeToChapter(c *gin.Context) {
	projectID := c.Param("projectId")
	nodeID := c.Param("nodeId")

	var req MergeToChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 获取节点
	node, err := h.db.GetNarrativeNode(nodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "节点不存在", ""))
		return
	}

	// 验证权限
	if node.ProjectID != projectID {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "无权限", ""))
		return
	}

	// 获取或创建章节
	var chapter *models.Chapter
	if req.ChapterID != nil {
		chapter, err = h.db.GetChapter(*req.ChapterID)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
			return
		}
	} else {
		// 创建新章节
		chapter = &models.Chapter{
			ID:        uuid.New().String(),
			ProjectID: projectID,
			ChapterNum: req.ChapterNum,
			Title:     req.ChapterTitle,
			Content:   "",
			Status:    models.ChapterStatusDraft,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// 根据合并策略合并内容
	switch req.MergeStrategy {
	case models.MergeStrategyAppend:
		chapter.Content += node.Content
	case models.MergeStrategyPrepend:
		chapter.Content = node.Content + chapter.Content
	case models.MergeStrategyReplace:
		chapter.Content = node.Content
	case models.MergeStrategyInsert:
		if req.Position != nil {
			// 在指定位置插入
			runes := []rune(chapter.Content)
			if *req.Position > len(runes) {
				*req.Position = len(runes)
			}
			chapter.Content = string(runes[:*req.Position]) + node.Content + string(runes[*req.Position:])
		} else {
			chapter.Content += node.Content
		}
	default:
		chapter.Content += node.Content
	}

	// 更新章节元数据
	chapter.WordCount = len(chapter.Content)
	chapter.UpdatedAt = time.Now()

	// 保存章节
	if err := h.db.SaveChapter(chapter); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("SAVE_FAILED", "保存章节失败", err.Error()))
		return
	}

	// 创建映射关系
	mapping := &models.NodeChapterMapping{
		ID:            uuid.New().String(),
		ProjectID:     projectID,
		ChapterID:     chapter.ID,
		NodeID:        nodeID,
		MappingType:   req.MappingType,
		Sequence:      req.Sequence,
		MergeStrategy: req.MergeStrategy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	h.db.SaveNodeChapterMapping(mapping)

	// 更新节点状态
	node.ChapterID = &chapter.ID
	node.NodeStatus = models.NodeStatusMerged
	node.UpdatedAt = time.Now()
	h.db.SaveNarrativeNode(node)

	c.JSON(http.StatusOK, successResponse(gin.H{
		"chapter": toChapterResponse(chapter),
		"mapping": toMappingResponse(mapping),
	}))
}


