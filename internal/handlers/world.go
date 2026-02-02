// Package handlers HTTP处理器
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/worldbuilder"
)

// WorldHandler 世界处理器
type WorldHandler struct {
	worldBuilder *worldbuilder.WorldBuilder
}

// NewWorldHandler 创建世界处理器
func NewWorldHandler(orc interface{}) *WorldHandler {
	return &WorldHandler{
		// 世界构建器按需创建
		worldBuilder: nil,
	}
}

// CreateWorld 创建世界
// @Summary 创建新世界
// @Description 创建一个新的世界设定
// @Tags worlds
// @Accept json
// @Produce json
// @Param request body CreateWorldRequest true "世界信息"
// @Success 200 {object} APIResponse
// @Router /api/v1/worlds [post]
func (h *WorldHandler) CreateWorld(c *gin.Context) {
	var req CreateWorldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 创建世界构建器
	if h.worldBuilder == nil {
		wb, err := worldbuilder.New()
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("INIT_FAILED", "初始化失败", err.Error()))
			return
		}
		h.worldBuilder = wb
	}

	// 构建世界
	world, err := h.worldBuilder.Build(worldbuilder.BuildParams{
		Name:  req.Name,
		Type:  parseWorldType(req.Type),
		Scale: parseWorldScale(req.Scale),
		Theme: req.Theme,
		Style: req.Style,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("BUILD_FAILED", "构建世界失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(toWorldResponse(world)))
}

// ListWorlds 列出所有世界
// @Summary 获取世界列表
// @Description 获取所有世界设定
// @Tags worlds
// @Produce json
// @Success 200 {object} APIResponse
// @Router /api/v1/worlds [get]
func (h *WorldHandler) ListWorlds(c *gin.Context) {
	worlds := db.Get().ListWorlds()

	response := make([]WorldResponse, 0, len(worlds))
	for _, w := range worlds {
		response = append(response, toWorldResponse(w))
	}

	c.JSON(http.StatusOK, successResponse(response))
}

// GetWorld 获取世界详情
// @Summary 获取世界详情
// @Description 获取指定世界的详细信息
// @Tags worlds
// @Produce json
// @Param id path string true "世界ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/worlds/{id} [get]
func (h *WorldHandler) GetWorld(c *gin.Context) {
	id := c.Param("id")

	world, err := db.Get().GetWorld(id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "世界不存在", ""))
		return
	}

	c.JSON(http.StatusOK, successResponse(toWorldResponse(world)))
}

// DeleteWorld 删除世界
// @Summary 删除世界
// @Description 删除指定的世界设定
// @Tags worlds
// @Produce json
// @Param id path string true "世界ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/worlds/{id} [delete]
func (h *WorldHandler) DeleteWorld(c *gin.Context) {
	id := c.Param("id")

	// 检查世界是否存在
	world, err := db.Get().GetWorld(id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "世界不存在", ""))
		return
	}

	// 删除世界
	if err := db.Get().DeleteWorld(id); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DELETE_FAILED", "删除世界失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"deleted_world_id":   world.ID,
		"deleted_world_name": world.Name,
	}))
}

// toWorldResponse 转换世界响应
func toWorldResponse(w *models.WorldSetting) WorldResponse {
	return WorldResponse{
		ID:            w.ID,
		Name:          w.Name,
		Type:          string(w.Type),
		Scale:         string(w.Scale),
		Style:         w.Style,
		CoreQuestion:  w.Philosophy.CoreQuestion,
		HighestGood:   w.Philosophy.ValueSystem.HighestGood,
		UltimateEvil:  w.Philosophy.ValueSystem.UltimateEvil,
		SocialConflicts: len(w.StorySoil.SocialConflicts),
		RegionCount:   len(w.Geography.Regions),
		RaceCount:     len(w.Civilization.Races),
		CreatedAt:     w.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// parseWorldType 解析世界类型
func parseWorldType(t string) models.WorldType {
	switch t {
	case "fantasy":
		return models.WorldFantasy
	case "scifi":
		return models.WorldScifi
	case "historical":
		return models.WorldHistorical
	case "urban":
		return models.WorldUrban
	case "wuxia":
		return models.WorldWuxia
	case "xianxia":
		return models.WorldXianxia
	default:
		return models.WorldMixed
	}
}

// parseWorldScale 解析世界规模
func parseWorldScale(s string) models.WorldScale {
	switch s {
	case "village":
		return models.ScaleVillage
	case "city":
		return models.ScaleCity
	case "nation":
		return models.ScaleNation
	case "continent":
		return models.ScaleContinent
	case "planet":
		return models.ScalePlanet
	case "universe":
		return models.ScaleUniverse
	default:
		return models.ScaleContinent
	}
}
