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
		WorldID:     req.WorldID,
		StoryType:   req.StoryType,
		Theme:       req.Theme,
		Protagonist: req.Protagonist,
		Length:      req.Length,
		ChapterCount: req.ChapterCount,
		Structure:   parseNarrativeStructure(req.Structure),
	}

	// 创建蓝图
	blueprint, err := engine.CreateBlueprint(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("CREATE_FAILED", "创建蓝图失败", err.Error()))
		return
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
		"message": "导出功能开发中",
		"blueprint_id": id,
		"format":     format,
	}))
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
		CreatedAt:     b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
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
