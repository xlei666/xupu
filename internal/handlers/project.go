// Package handlers HTTP处理器
package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/orchestrator"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	orchestrator *orchestrator.Orchestrator
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(orc *orchestrator.Orchestrator) *ProjectHandler {
	return &ProjectHandler{
		orchestrator: orc,
	}
}

// CreateProject 创建项目
// @Summary 创建新项目
// @Description 创建一个新的AI小说创作项目
// @Tags projects
// @Accept json
// @Produce json
// @Param request body CreateProjectRequest true "项目信息"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	var project *models.Project
	var err error

	// 获取当前用户ID
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "未授权", ""))
		return
	}

	// 如果没有提供创作参数，创建简单的空项目草稿
	if req.Params == nil {
		// 创建空项目
		project = &models.Project{
			ID:          db.GenerateID("project"),
			UserID:      userID,
			Name:        req.Name,
			Description: req.Description,
			Mode:        models.OrchestrationMode(req.Mode),
			Status:      models.StatusDraft,
			Progress:    0,
		}

		if err := db.Get().SaveProject(project); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("CREATE_FAILED", "创建项目失败", err.Error()))
			return
		}
	} else {
		// 使用 orchestrator 创建完整的AI项目
		// 构建创作参数
		params := orchestrator.CreationParams{
			UserID:       userID,
			ProjectName:  req.Name,
			Description:  req.Description,
			WorldName:    req.Params.WorldName,
			WorldType:    req.Params.WorldType,
			WorldTheme:   req.Params.WorldTheme,
			WorldScale:   req.Params.WorldScale,
			WorldStyle:   req.Params.WorldStyle,
			StoryType:    req.Params.StoryType,
			StoryTheme:   req.Params.Theme,
			Protagonist:  req.Params.Protagonist,
			StoryLength:  req.Params.Length,
			ChapterCount: req.Params.ChapterCount,
			Structure:    req.Params.Structure,
			Options: orchestrator.GenerationOptions{
				SkipWorldBuild:      req.Params.Options.SkipWorldBuild,
				ExistingWorldID:     req.Params.Options.ExistingWorldID,
				SkipNarrative:       req.Params.Options.SkipNarrative,
				ExistingBlueprintID: req.Params.Options.ExistingBlueprintID,
				GenerateContent:     req.Params.Options.GenerateContent,
				StartChapter:        req.Params.Options.StartChapter,
				EndChapter:          req.Params.Options.EndChapter,
				Style:               req.Params.Options.Style,
			},
		}

		// 创建项目
		project, err = h.orchestrator.CreateProject(params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("CREATE_FAILED", "创建项目失败", err.Error()))
			return
		}
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"project": toProjectResponse(project),
	}))
}

// ListProjects 列出所有项目
// @Summary 获取项目列表
// @Description 获取当前用户的所有AI小说创作项目
// @Tags projects
// @Produce json
// @Param status query string false "状态筛选 (all/draft/generating/completed)"
// @Param sortBy query string false "排序字段 (updated/created/name)"
// @Param search query string false "搜索关键词"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "未授权", ""))
		return
	}

	// 获取筛选参数
	status := c.DefaultQuery("status", "all")
	sortBy := c.DefaultQuery("sortBy", "updated")
	search := c.Query("search")

	// 只查询当前用户的项目
	projects := db.Get().ListProjectsByUser(userID)

	// 筛选
	filtered := make([]*models.Project, 0, len(projects))
	for _, p := range projects {
		// 状态筛选
		if status != "all" && string(p.Status) != status {
			continue
		}

		// 搜索筛选
		if search != "" {
			searchLower := strings.ToLower(search)
			nameMatch := strings.Contains(strings.ToLower(p.Name), searchLower)
			descMatch := strings.Contains(strings.ToLower(p.Description), searchLower)
			if !nameMatch && !descMatch {
				continue
			}
		}

		filtered = append(filtered, p)
	}

	// 排序
	switch sortBy {
	case "created":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
		})
	case "name":
		sort.Slice(filtered, func(i, j int) bool {
			return strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		})
	default: // updated
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].UpdatedAt.After(filtered[j].UpdatedAt)
		})
	}

	response := make([]ProjectResponse, 0, len(filtered))
	for _, p := range filtered {
		response = append(response, toProjectResponse(p))
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"projects": response,
		"page":     1,
		"pageSize": len(filtered),
		"total":    len(filtered),
	}))
}

// GetProject 获取项目详情
// @Summary 获取项目详情
// @Description 获取指定项目的详细信息
// @Tags projects
// @Produce json
// @Param id path string true "项目ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id := c.Param("projectId")

	project, err := db.Get().GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	// 检查权限：只能访问自己的项目
	userID, exists := GetUserID(c)
	if !exists || project.UserID != userID {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "无权访问", ""))
		return
	}

	// 获取进度信息
	progress, _ := h.orchestrator.GetProjectProgress(id)

	response := gin.H{
		"project":  toProjectResponse(project),
		"progress": progress,
	}

	c.JSON(http.StatusOK, successResponse(response))
}

// DeleteProject 删除项目
// @Summary 删除项目
// @Description 删除指定的项目及其所有关联数据
// @Tags projects
// @Produce json
// @Param id path string true "项目ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("projectId")

	// 检查项目是否存在
	project, err := db.Get().GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	// 检查权限：只能删除自己的项目
	userID, exists := GetUserID(c)
	if !exists || project.UserID != userID {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "无权删除", ""))
		return
	}

	// 删除项目
	if err := db.Get().DeleteProject(id); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DELETE_FAILED", "删除项目失败", err.Error()))
		return
	}

	// TODO: 级联删除相关数据（蓝图、场景等）

	c.JSON(http.StatusOK, successResponse(gin.H{
		"deleted_project_id":   project.ID,
		"deleted_project_name": project.Name,
	}))
}

// GenerateChapter 生成章节
// @Summary 生成章节内容
// @Description 为指定项目生成章节内容
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "项目ID"
// @Param request body GenerateChapterRequest false "生成选项"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/generate [post]
func (h *ProjectHandler) GenerateChapter(c *gin.Context) {
	id := c.Param("projectId")

	project, err := db.Get().GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	// 检查项目状态
	if project.Status != models.StatusCompleted {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATUS", "项目状态不允许生成", "请先完成项目规划"))
		return
	}

	// TODO: 实现章节生成逻辑
	c.JSON(http.StatusOK, successResponse(gin.H{
		"message":    "章节生成功能开发中",
		"project_id": id,
	}))
}

// Intervene 干预剧情
// @Summary 干预剧情
// @Description 在创作过程中干预剧情发展
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "项目ID"
// @Param request body InterveneRequest true "干预内容"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/intervene [post]
func (h *ProjectHandler) Intervene(c *gin.Context) {
	id := c.Param("projectId")

	var req InterveneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// TODO: 实现干预逻辑
	c.JSON(http.StatusOK, successResponse(gin.H{
		"message":           "干预功能开发中",
		"project_id":        id,
		"intervention_type": req.Type,
	}))
}

// PauseGeneration 暂停生成
// @Summary 暂停生成
// @Description 暂停正在进行的生成任务
// @Tags projects
// @Produce json
// @Param id path string true "项目ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/pause [post]
func (h *ProjectHandler) PauseGeneration(c *gin.Context) {
	id := c.Param("projectId")

	if err := h.orchestrator.PauseGeneration(id); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("PAUSE_FAILED", "暂停失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"project_id": id,
		"status":     "paused",
	}))
}

// ResumeGeneration 恢复生成
// @Summary 恢复生成
// @Description 恢复暂停的生成任务
// @Tags projects
// @Produce json
// @Param id path string true "项目ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/resume [post]
func (h *ProjectHandler) ResumeGeneration(c *gin.Context) {
	id := c.Param("projectId")

	if err := h.orchestrator.ResumeGeneration(id); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("RESUME_FAILED", "恢复失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"project_id": id,
		"status":     "generating",
	}))
}

// GetProgress 获取项目进度
// @Summary 获取项目进度
// @Description 获取项目的详细进度信息
// @Tags projects
// @Produce json
// @Param id path string true "项目ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{id}/progress [get]
func (h *ProjectHandler) GetProgress(c *gin.Context) {
	id := c.Param("projectId")

	progress, err := h.orchestrator.GetProjectProgress(id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	c.JSON(http.StatusOK, successResponse(progress))
}
