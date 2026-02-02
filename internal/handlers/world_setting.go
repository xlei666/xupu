// Package handlers HTTP处理器
package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/worldbuilder"
)

// WorldSettingHandler 世界设定处理器
type WorldSettingHandler struct {
	db            db.Database
	worldBuilder  *worldbuilder.WorldBuilder
}

// NewWorldSettingHandler 创建世界设定处理器
func NewWorldSettingHandler(database db.Database, worldBuilder *worldbuilder.WorldBuilder) *WorldSettingHandler {
	return &WorldSettingHandler{
		db:           database,
		worldBuilder: worldBuilder,
	}
}

// ============================================
// 请求/响应 DTO
// ============================================

// SaveWorldStagesRequest 保存世界设定请求
type SaveWorldStagesRequest struct {
	// 7个阶段的内容
	Philosophy   *models.Philosophy    `json:"philosophy"`
	Worldview    *models.Worldview     `json:"worldview"`
	Laws         *models.Laws          `json:"laws"`
	Geography    *models.Geography     `json:"geography"`
	Civilization *models.Civilization  `json:"civilization"`
	Society      *models.Society        `json:"society"`
	History      *models.History        `json:"history"`
}

// GenerateWorldStageRequest 生成特定阶段请求
type GenerateWorldStageRequest struct {
	Stage    string                 `json:"stage" binding:"required"` // 阶段名称
	Context  string                 `json:"context"`                 // 用户输入的上下文或约束
	Settings map[string]interface{} `json:"settings"`                // 额外设置
}

// ============================================
// API Handlers
// ============================================

// SaveWorldStages 保存世界设定的7个阶段
// @Summary 保存世界设定
// @Description 保存或更新项目的世界设定（7个阶段）
// @Tags world-settings
// @Accept json
// @Produce json
// @Param projectId path string true "项目ID"
// @Param request body SaveWorldStagesRequest true "世界设定数据"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{projectId}/world-stages [post]
func (h *WorldSettingHandler) SaveWorldStages(c *gin.Context) {
	projectID := c.Param("projectId")

	// 验证项目存在
	project, err := h.db.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	var req SaveWorldStagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 获取或创建世界设定
	var world *models.WorldSetting
	if project.WorldID != "" {
		world, err = h.db.GetWorld(project.WorldID)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "世界设定不存在", ""))
			return
		}
	} else {
		// 创建新的世界设定
		world = &models.WorldSetting{
			ID:        db.GenerateID("world"),
			Name:      project.Name + "的世界",
			Type:      models.WorldFantasy, // 默认类型
			Scale:     models.ScaleNation,  // 默认规模
			Style:     "通用",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// 更新项目的 WorldID
		project.WorldID = world.ID
		h.db.SaveProject(project)
	}

	// 更新各个阶段
	if req.Philosophy != nil {
		world.Philosophy = *req.Philosophy
	}
	if req.Worldview != nil {
		world.Worldview = *req.Worldview
	}
	if req.Laws != nil {
		world.Laws = *req.Laws
	}
	if req.Geography != nil {
		world.Geography = *req.Geography
	}
	if req.Civilization != nil {
		world.Civilization = *req.Civilization
	}
	if req.Society != nil {
		world.Society = *req.Society
	}
	if req.History != nil {
		world.History = *req.History
	}

	world.UpdatedAt = time.Now()

	// 保存
	if err := h.db.SaveWorld(world); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("SAVE_FAILED", "保存失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"world_id": world.ID,
		"message":  "保存成功",
	}))
}

// GetWorldStages 获取世界设定的7个阶段
// @Summary 获取世界设定
// @Description 获取项目的世界设定（7个阶段）
// @Tags world-settings
// @Produce json
// @Param projectId path string true "项目ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{projectId}/world-stages [get]
func (h *WorldSettingHandler) GetWorldStages(c *gin.Context) {
	projectID := c.Param("projectId")

	// 验证项目存在
	project, err := h.db.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	// 如果项目没有关联世界，返回空数据
	if project.WorldID == "" {
		c.JSON(http.StatusOK, successResponse(gin.H{
			"world_id": nil,
			"stages": gin.H{
				"philosophy":   nil,
				"worldview":    nil,
				"laws":         nil,
				"geography":    nil,
				"civilization": nil,
				"society":      nil,
				"history":      nil,
			},
		}))
		return
	}

	// 获取世界设定
	world, err := h.db.GetWorld(project.WorldID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "世界设定不存在", ""))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"world_id": world.ID,
		"stages": gin.H{
			"philosophy":   world.Philosophy,
			"worldview":    world.Worldview,
			"laws":         world.Laws,
			"geography":    world.Geography,
			"civilization": world.Civilization,
			"society":      world.Society,
			"history":      world.History,
		},
	}))
}

// GenerateWorldStage 生成特定阶段（调用WorldBuilder）
// @Summary AI生成世界阶段
// @Description 使用AI生成指定阶段的设定内容
// @Tags world-settings
// @Accept json
// @Produce json
// @Param projectId path string true "项目ID"
// @Param stage path string true "阶段名称" Enums(philosophy, worldview, laws, geography, civilization, society, history)
// @Param request body GenerateWorldStageRequest true "生成参数"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{projectId}/world-stages/{stage}/generate [post]
func (h *WorldSettingHandler) GenerateWorldStage(c *gin.Context) {
	projectID := c.Param("projectId")
	stage := c.Param("stage")

	// 验证项目存在
	project, err := h.db.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	var req GenerateWorldStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 获取或创建世界设定
	var world *models.WorldSetting
	if project.WorldID != "" {
		world, err = h.db.GetWorld(project.WorldID)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "世界设定不存在", ""))
			return
		}
	} else {
		// 创建新的世界设定
		world = &models.WorldSetting{
			ID:        db.GenerateID("world"),
			Name:      project.Name + "的世界",
			Type:      models.WorldFantasy,
			Scale:     models.ScaleNation,
			Style:     "通用",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// 更新项目的 WorldID
		project.WorldID = world.ID
		h.db.SaveProject(project)
	}

	// 调用 WorldBuilder 生成指定阶段
	switch stage {
	case "philosophy":
		result, err := h.worldBuilder.GenerateStage1(worldbuilder.Stage1Input{
			WorldType: string(world.Type),
			Theme:     req.Context,
			Style:     world.Style,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("GENERATE_FAILED", "生成失败", err.Error()))
			return
		}
		world.Philosophy = *result

	case "worldview":
		if world.Philosophy.CoreQuestion == "" {
			c.JSON(http.StatusBadRequest, errorResponse("MISSING_DEPENDENCY", "缺少前置阶段（哲学）", ""))
			return
		}
		result, err := h.worldBuilder.GenerateStage2(worldbuilder.Stage2Input{
			CoreQuestion:  world.Philosophy.CoreQuestion,
			HighestGood:   world.Philosophy.ValueSystem.HighestGood,
			UltimateEvil: world.Philosophy.ValueSystem.UltimateEvil,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("GENERATE_FAILED", "生成失败", err.Error()))
			return
		}
		world.Worldview = *result

	case "laws":
		if world.Worldview.Cosmology.Origin == "" {
			c.JSON(http.StatusBadRequest, errorResponse("MISSING_DEPENDENCY", "缺少前置阶段（世界观）", ""))
			return
		}
		worldviewSummary := fmt.Sprintf("起源:%s 结构:%s", world.Worldview.Cosmology.Origin, world.Worldview.Cosmology.Structure)
		result, err := h.worldBuilder.GenerateStage3(worldbuilder.Stage3Input{
			WorldType: string(world.Type),
			Worldview: worldviewSummary,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("GENERATE_FAILED", "生成失败", err.Error()))
			return
		}
		world.Laws = *result

	case "story_soil":
		if world.Philosophy.CoreQuestion == "" {
			c.JSON(http.StatusBadRequest, errorResponse("MISSING_DEPENDENCY", "缺少前置阶段（哲学）", ""))
			return
		}
		mainConflicts := ""
		if len(world.Philosophy.ValueSystem.MoralDilemmas) > 0 {
			mainConflicts = world.Philosophy.ValueSystem.MoralDilemmas[0].Dilemma
		}
		result, err := h.worldBuilder.GenerateStage4(worldbuilder.Stage4Input{
			CoreQuestion:  world.Philosophy.CoreQuestion,
			MainConflicts: mainConflicts,
			WorldType:     string(world.Type),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("GENERATE_FAILED", "生成失败", err.Error()))
			return
		}
		world.StorySoil = *result

	case "geography":
		if world.Laws.Physics.Gravity == "" {
			c.JSON(http.StatusBadRequest, errorResponse("MISSING_DEPENDENCY", "缺少前置阶段（法则）", ""))
			return
		}
		lawsSummary := fmt.Sprintf("物理:%s 超自然:%v", world.Laws.Physics.Gravity, world.Laws.Supernatural != nil && world.Laws.Supernatural.Exists)
		civilizationNeeds := fmt.Sprintf("资源需求基于%s类型的世界", world.Type)
		result, err := h.worldBuilder.GenerateStage5(worldbuilder.Stage5Input{
			WorldType:         string(world.Type),
			WorldScale:        string(world.Scale),
			LawsSummary:       lawsSummary,
			CivilizationNeeds: civilizationNeeds,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("GENERATE_FAILED", "生成失败", err.Error()))
			return
		}
		world.Geography = *result

	case "civilization_society":
		if len(world.Geography.Regions) == 0 {
			c.JSON(http.StatusBadRequest, errorResponse("MISSING_DEPENDENCY", "缺少前置阶段（地理）", ""))
			return
		}
		geographySummary := fmt.Sprintf("%d个区域, 气候:%s", len(world.Geography.Regions), func() string {
			if world.Geography.Climate != nil {
				return world.Geography.Climate.Type
			}
			return "未知"
		}())
		valueSystem := fmt.Sprintf("最高善:%s", world.Philosophy.ValueSystem.HighestGood)

		result, err := h.worldBuilder.GenerateStage6(worldbuilder.Stage6Input{
			WorldType:       string(world.Type),
			GeographySummary: geographySummary,
			ValueSystem:     valueSystem,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("GENERATE_FAILED", "生成失败", err.Error()))
			return
		}
		world.Civilization = *result.Civilization
		world.Society = *result.Society

	default:
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STAGE", "无效的阶段名称", "有效阶段: philosophy, worldview, laws, story_soil, geography, civilization_society"))
		return
	}

	world.UpdatedAt = time.Now()

	// 保存
	if err := h.db.SaveWorld(world); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("SAVE_FAILED", "保存失败", err.Error()))
		return
	}

	// 返回生成的阶段
	var result interface{}
	switch stage {
	case "philosophy":
		result = world.Philosophy
	case "worldview":
		result = world.Worldview
	case "laws":
		result = world.Laws
	case "story_soil":
		result = world.StorySoil
	case "geography":
		result = world.Geography
	case "civilization_society":
		result = gin.H{
			"civilization": world.Civilization,
			"society":      world.Society,
		}
	default:
		result = gin.H{}
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"world_id": world.ID,
		"stage":    stage,
		"content":  result,
		"message":  "生成成功",
	}))
}

// GachaWorldSettings 抽卡生成世界设定（一次性生成所有7个阶段）
// @Summary 抽卡生成世界设定
// @Description AI抽卡一次性生成完整的世界设定（7个阶段）
// @Tags world-settings
// @Accept json
// @Produce json
// @Param projectId path string true "项目ID"
// @Param request body GenerateWorldStageRequest true "生成参数"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{projectId}/world-stages/gacha [post]
func (h *WorldSettingHandler) GachaWorldSettings(c *gin.Context) {
	projectID := c.Param("projectId")

	// 验证项目存在
	project, err := h.db.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	var req GenerateWorldStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 请求体可选，如果为空则使用默认值
		req = GenerateWorldStageRequest{
			Context:  "",
			Settings: make(map[string]interface{}),
		}
	}

	// 获取或创建世界设定
	var world *models.WorldSetting
	if project.WorldID != "" {
		world, err = h.db.GetWorld(project.WorldID)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "世界设定不存在", ""))
			return
		}
	} else {
		// 从请求参数中获取世界类型和风格
		worldType := models.WorldFantasy // 默认奇幻
		worldStyle := "通用"               // 默认风格

		// 解析settings中的world_type和style
		if req.Settings != nil {
			if wt, ok := req.Settings["world_type"].(string); ok && wt != "" {
				worldType = models.WorldType(wt)
			}
			if ws, ok := req.Settings["style"].(string); ok && ws != "" {
				worldStyle = ws
			}
		}

		// 创建新的世界设定
		world = &models.WorldSetting{
			ID:        db.GenerateID("world"),
			Name:      project.Name + "的世界",
			Type:      worldType,
			Scale:     models.ScaleNation,
			Style:     worldStyle,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// 更新项目的 WorldID
		project.WorldID = world.ID
		h.db.SaveProject(project)
	}

	// 依次生成所有7个阶段
	stages := []string{"philosophy", "worldview", "laws", "story_soil", "geography", "civilization_society"}

	for _, stage := range stages {
		switch stage {
		case "philosophy":
			result, err := h.worldBuilder.GenerateStage1(worldbuilder.Stage1Input{
				WorldType: string(world.Type),
				Theme:     req.Context,
				Style:     world.Style,
			})
			if err == nil {
				world.Philosophy = *result
			}

		case "worldview":
			if world.Philosophy.CoreQuestion != "" {
				result, err := h.worldBuilder.GenerateStage2(worldbuilder.Stage2Input{
					CoreQuestion:  world.Philosophy.CoreQuestion,
					HighestGood:   world.Philosophy.ValueSystem.HighestGood,
					UltimateEvil: world.Philosophy.ValueSystem.UltimateEvil,
				})
				if err == nil {
					world.Worldview = *result
				}
			}

		case "laws":
			if world.Worldview.Cosmology.Origin != "" {
				worldviewSummary := fmt.Sprintf("起源:%s 结构:%s", world.Worldview.Cosmology.Origin, world.Worldview.Cosmology.Structure)
				result, err := h.worldBuilder.GenerateStage3(worldbuilder.Stage3Input{
					WorldType: string(world.Type),
					Worldview: worldviewSummary,
				})
				if err == nil {
					world.Laws = *result
				}
			}

		case "story_soil":
			if world.Philosophy.CoreQuestion != "" {
				mainConflicts := ""
				if len(world.Philosophy.ValueSystem.MoralDilemmas) > 0 {
					mainConflicts = world.Philosophy.ValueSystem.MoralDilemmas[0].Dilemma
				}
				result, err := h.worldBuilder.GenerateStage4(worldbuilder.Stage4Input{
					CoreQuestion:  world.Philosophy.CoreQuestion,
					MainConflicts: mainConflicts,
					WorldType:     string(world.Type),
				})
				if err == nil {
					world.StorySoil = *result
				}
			}

		case "geography":
			if world.Laws.Physics.Gravity != "" {
				lawsSummary := fmt.Sprintf("物理:%s 超自然:%v", world.Laws.Physics.Gravity, world.Laws.Supernatural != nil && world.Laws.Supernatural.Exists)
				civilizationNeeds := fmt.Sprintf("资源需求基于%s类型的世界", world.Type)
				result, err := h.worldBuilder.GenerateStage5(worldbuilder.Stage5Input{
					WorldType:         string(world.Type),
					WorldScale:        string(world.Scale),
					LawsSummary:       lawsSummary,
					CivilizationNeeds: civilizationNeeds,
				})
				if err == nil {
					world.Geography = *result
				}
			}

		case "civilization_society":
			if len(world.Geography.Regions) > 0 {
				geographySummary := fmt.Sprintf("%d个区域, 气候:%s", len(world.Geography.Regions), func() string {
					if world.Geography.Climate != nil {
						return world.Geography.Climate.Type
					}
					return "未知"
				}())
				valueSystem := fmt.Sprintf("最高善:%s", world.Philosophy.ValueSystem.HighestGood)

				result, err := h.worldBuilder.GenerateStage6(worldbuilder.Stage6Input{
					WorldType:       string(world.Type),
					GeographySummary: geographySummary,
					ValueSystem:     valueSystem,
				})
				if err == nil {
					world.Civilization = *result.Civilization
					world.Society = *result.Society
				}
			}
		}
	}

	world.UpdatedAt = time.Now()

	// 保存
	if err := h.db.SaveWorld(world); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("SAVE_FAILED", "保存失败", err.Error()))
		return
	}

	// 返回生成的所有阶段
	c.JSON(http.StatusOK, successResponse(gin.H{
		"world_id": world.ID,
		"stages": gin.H{
			"philosophy":         world.Philosophy,
			"worldview":          world.Worldview,
			"laws":               world.Laws,
			"story_soil":         world.StorySoil,
			"geography":          world.Geography,
			"civilization":       world.Civilization,
			"society":            world.Society,
		},
		"message": "世界设定抽卡成功！",
	}))
}
