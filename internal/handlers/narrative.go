// Package handlers HTTP处理器 - 叙事蓝图
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/narrative"
)

// NarrativeHandler 叙事处理器
type NarrativeHandler struct {
	// narrativeEngine *narrative.NarrativeEngine
}

// NewNarrativeHandler 创建叙事处理器
func NewNarrativeHandler(orc interface{}) *NarrativeHandler {
	return &NarrativeHandler{}
}

// CreateBlueprint 创建蓝图
// @Summary 创建叙事蓝图
// @Description 基于世界设定创建叙事蓝图
// @Tags blueprints
// @Accept json
// @Produce json
// @Param request body CreateBlueprintRequest true "蓝图信息"
// @Success 200 {object} APIResponse
// @Router /api/v1/blueprints [post]
func (h *NarrativeHandler) CreateBlueprint(c *gin.Context) {
	var req CreateBlueprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 创建叙事引擎
	engine, err := narrative.New()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INIT_FAILED", "初始化失败", err.Error()))
		return
	}

	// 构建参数
	params := narrative.CreateParams{
		WorldID:      req.WorldID,
		StoryType:    req.StoryType,
		Theme:        req.Theme,
		Protagonist:  req.Protagonist,
		Length:       req.Length,
		ChapterCount: req.ChapterCount,
		Structure:    parseNarrativeStructure(req.Structure),
	}

	// 创建蓝图
	blueprint, err := engine.CreateBlueprint(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("CREATE_FAILED", "创建蓝图失败", err.Error()))
		return
	}

	// 如果提供了 ProjectID，关联到项目
	if req.ProjectID != "" {
		blueprint.ProjectID = req.ProjectID
		// 更新蓝图以保存 ProjectID
		if err := db.Get().SaveNarrativeBlueprint(blueprint); err != nil {
			// logging error but not failing request
		}

		// 更新项目
		if project, err := db.Get().GetProject(req.ProjectID); err == nil {
			project.NarrativeID = blueprint.ID
			db.Get().SaveProject(project)
		}
	}

	c.JSON(http.StatusOK, successResponse(toBlueprintResponse(blueprint)))
}

// GetBlueprint 获取蓝图详情
// @Summary 获取蓝图详情
// @Description 获取指定蓝图的详细信息
// @Tags blueprints
// @Produce json
// @Param id path string true "蓝图ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/blueprints/{id} [get]
func (h *NarrativeHandler) GetBlueprint(c *gin.Context) {
	id := c.Param("id")

	blueprint, err := db.Get().GetNarrativeBlueprint(id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "蓝图不存在", ""))
		return
	}

	c.JSON(http.StatusOK, successResponse(toBlueprintResponse(blueprint)))
}

// ExportBlueprint 导出蓝图
// @Summary 导出蓝图
// @Description 将蓝图导出为指定格式
// @Tags blueprints
// @Produce json
// @Param id path string true "蓝图ID"
// @Param format query string false "导出格式" Enums(markdown, json)
// @Success 200 {object} APIResponse
// @Router /api/v1/blueprints/{id}/export [get]
func (h *NarrativeHandler) ExportBlueprint(c *gin.Context) {
	id := c.Param("id")
	format := c.DefaultQuery("format", "json")

	_, err := db.Get().GetNarrativeBlueprint(id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "蓝图不存在", ""))
		return
	}

	// TODO: 实现导出逻辑（使用export.go中的ExportHandler）
	c.JSON(http.StatusOK, successResponse(gin.H{
		"message":      "导出功能开发中",
		"blueprint_id": id,
		"format":       format,
	}))
}

// ApplyBlueprint 应用蓝图（创建章节）
// @Summary 应用蓝图
// @Description 将蓝图中的章节规划应用到项目，创建实际的章节记录
// @Tags blueprints
// @Produce json
// @Param project_id path string true "项目ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{project_id}/blueprint/apply [post]
func (h *NarrativeHandler) ApplyBlueprint(c *gin.Context) {
	projectID := c.Param("projectId") // Note: Gin route usually uses :projectId

	project, err := db.Get().GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	if project.NarrativeID == "" {
		c.JSON(http.StatusBadRequest, errorResponse("NO_BLUEPRINT", "项目尚未关联蓝图", ""))
		return
	}

	blueprint, err := db.Get().GetNarrativeBlueprint(project.NarrativeID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "蓝图不存在", ""))
		return
	}

	// 批量创建章节
	createdCount := 0
	for _, plan := range blueprint.ChapterPlans {
		// 检查章节是否已存在
		existing, _ := db.Get().GetChapterByNum(projectID, plan.Chapter)
		if existing != nil {
			continue // 跳过已存在的章节
		}

		chapter := &models.Chapter{
			ProjectID:  projectID,
			ChapterNum: plan.Chapter,
			Title:      plan.Title,
			Status:     models.ChapterStatusDraft,
		}

		if err := db.Get().SaveChapter(chapter); err == nil {
			createdCount++
		}
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"message":       "蓝图应用成功",
		"created_count": createdCount,
	}))
}

// parseNarrativeStructure 解析叙事结构
func parseNarrativeStructure(s string) narrative.NarrativeStructure {
	switch s {
	case "three_act":
		return narrative.StructureThreeAct
	case "heros_journey":
		return narrative.StructureHerosJourney
	case "save_the_cat":
		return narrative.StructureSaveTheCat
	case "kishotenketsu":
		return narrative.StructureKishotenketsu
	case "freytag_pyramid":
		return narrative.StructureFreytagPyramid
	default:
		return narrative.StructureThreeAct
	}
}
