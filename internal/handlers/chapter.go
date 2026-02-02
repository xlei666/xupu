// Package handlers HTTP处理器
package handlers

import (
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/internal/repositories"
	"github.com/xlei/xupu/pkg/db"
)

// ChapterHandler 章节处理器
type ChapterHandler struct {
	chapterRepo *repositories.ChapterRepository
}

// NewChapterHandler 创建章节处理器
func NewChapterHandler() *ChapterHandler {
	return &ChapterHandler{
		chapterRepo: repositories.NewChapterRepository(),
	}
}

// ListChapters 获取章节列表
// @Summary 获取项目的章节列表
// @Description 获取指定项目的所有章节
// @Tags chapters
// @Produce json
// @Param project_id path string true "项目ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{project_id}/chapters [get]
func (h *ChapterHandler) ListChapters(c *gin.Context) {
	projectID := c.Param("projectId")

	// 检查项目是否存在
	project, err := db.Get().GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	// 获取章节列表
	chapters, err := h.chapterRepo.ListByProjectID(c, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "获取章节列表失败", err.Error()))
		return
	}

	// 转换为响应格式
	response := make([]ChapterResponse, 0, len(chapters))
	for _, chapter := range chapters {
		response = append(response, toChapterResponse(&chapter))
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"project_id": projectID,
		"project_name": project.Name,
		"chapters":  response,
		"total":     len(response),
	}))
}

// GetChapter 获取章节详情
// @Summary 获取章节详情
// @Description 获取指定章节的详细信息
// @Tags chapters
// @Produce json
// @Param project_id path string true "项目ID"
// @Param chapter_id path string true "章节ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{project_id}/chapters/{chapter_id} [get]
func (h *ChapterHandler) GetChapter(c *gin.Context) {
	projectID := c.Param("projectId")
	chapterID := c.Param("chapterId")

	// 检查项目是否存在
	if _, err := db.Get().GetProject(projectID); err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	// 获取章节
	chapter, err := h.chapterRepo.GetByID(c, chapterID)
	if err != nil {
		if err == repositories.ErrChapterNotFound {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "获取章节失败", err.Error()))
		return
	}

	// 验证章节是否属于该项目
	if chapter.ProjectID != projectID {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"chapter": toChapterResponse(chapter),
	}))
}

// CreateChapter 创建章节
// @Summary 创建新章节
// @Description 在指定项目中创建新章节
// @Tags chapters
// @Accept json
// @Produce json
// @Param project_id path string true "项目ID"
// @Param request body CreateChapterRequest true "章节信息"
// @Success 201 {object} APIResponse
// @Router /api/v1/projects/{project_id}/chapters [post]
func (h *ChapterHandler) CreateChapter(c *gin.Context) {
	projectID := c.Param("projectId")

	// 检查项目是否存在
	if _, err := db.Get().GetProject(projectID); err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	var req CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 如果没有指定章节号，自动设置为最大章节号+1
	chapterNum := req.ChapterNum
	if chapterNum == 0 {
		maxChapterNum, err := h.chapterRepo.GetMaxChapterNum(c, projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "获取章节号失败", err.Error()))
			return
		}
		chapterNum = maxChapterNum + 1
	}

	// 创建章节
	chapter := &models.Chapter{
		ProjectID:  projectID,
		ChapterNum: chapterNum,
		Title:      req.Title,
		Content:    "",
		WordCount:  0,
		AIWordCount: 0,
		Status:     models.ChapterStatusDraft,
	}

	if err := h.chapterRepo.Create(c, chapter); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "创建章节失败", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, successResponse(gin.H{
		"chapter": toChapterResponse(chapter),
	}))
}

// UpdateChapter 更新章节
// @Summary 更新章节
// @Description 更新指定章节的内容
// @Tags chapters
// @Accept json
// @Produce json
// @Param project_id path string true "项目ID"
// @Param chapter_id path string true "章节ID"
// @Param request body UpdateChapterRequest true "更新内容"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{project_id}/chapters/{chapter_id} [put]
func (h *ChapterHandler) UpdateChapter(c *gin.Context) {
	projectID := c.Param("projectId")
	chapterID := c.Param("chapterId")

	// 检查项目是否存在
	if _, err := db.Get().GetProject(projectID); err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	// 获取章节
	chapter, err := h.chapterRepo.GetByID(c, chapterID)
	if err != nil {
		if err == repositories.ErrChapterNotFound {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "获取章节失败", err.Error()))
		return
	}

	// 验证章节是否属于该项目
	if chapter.ProjectID != projectID {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
		return
	}

	var req UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 更新字段
	if req.Title != "" {
		chapter.Title = req.Title
	}
	if req.Content != "" {
		chapter.Content = req.Content
		chapter.WordCount = utf8.RuneCountInString(req.Content)
	}
	if req.Status != "" {
		chapter.Status = models.ChapterStatus(req.Status)
	}

	// 保存更新
	if err := h.chapterRepo.Update(c, chapter); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "更新章节失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"chapter": toChapterResponse(chapter),
	}))
}

// DeleteChapter 删除章节
// @Summary 删除章节
// @Description 删除指定章节
// @Tags chapters
// @Produce json
// @Param project_id path string true "项目ID"
// @Param chapter_id path string true "章节ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{project_id}/chapters/{chapter_id} [delete]
func (h *ChapterHandler) DeleteChapter(c *gin.Context) {
	projectID := c.Param("projectId")
	chapterID := c.Param("chapterId")

	// 检查项目是否存在
	if _, err := db.Get().GetProject(projectID); err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	// 获取章节
	chapter, err := h.chapterRepo.GetByID(c, chapterID)
	if err != nil {
		if err == repositories.ErrChapterNotFound {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "获取章节失败", err.Error()))
		return
	}

	// 验证章节是否属于该项目
	if chapter.ProjectID != projectID {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
		return
	}

	// 删除章节
	if err := h.chapterRepo.Delete(c, chapterID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "删除章节失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"deleted_chapter_id": chapterID,
		"chapter_num":        chapter.ChapterNum,
		"title":              chapter.Title,
	}))
}
