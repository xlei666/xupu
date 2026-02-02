// Package handlers HTTP处理器
package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
)

// SynopsisHandler 简介处理器
type SynopsisHandler struct {
	db db.Database
}

// NewSynopsisHandler 创建简介处理器
func NewSynopsisHandler(database db.Database) *SynopsisHandler {
	return &SynopsisHandler{
		db: database,
	}
}

// GachaSynopsis 抽卡生成简介设定
// @Summary 抽卡生成简介设定
// @Description AI抽卡生成作品简介大纲
// @Tags synopsis
// @Accept json
// @Produce json
// @Param projectId path string true "项目ID"
// @Param request body GachaSynopsisRequest true "生成参数"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{projectId}/synopsis/gacha [post]
func (h *SynopsisHandler) GachaSynopsis(c *gin.Context) {
	projectID := c.Param("projectId")

	// 验证项目存在
	project, err := h.db.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	var req GachaSynopsisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 生成简介
	synopsis := &models.Synopsis{
		ID:        db.GenerateID("synopsis"),
		ProjectID: projectID,
		WorldID:   project.WorldID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 基于项目名称生成
	synopsis.OneLineSummary = fmt.Sprintf("关于%s的故事", project.Name)
	synopsis.ShortSummary = fmt.Sprintf("%s是一个充满冒险与成长的故事。主角在面对重重困难时，不断突破自我，最终实现目标。作品融合了动作、冒险和成长元素，展现了人性的光辉与黑暗。", project.Name)

	synopsis.DetailedSummary = fmt.Sprintf(`《%s》是一部以成长为核心主题的作品。

故事从主角的平凡生活开始，在一次偶然的事件中，主角被卷入了一个宏大的冒险。在这个过程中，主角遇到了各种挑战：来自敌人的威胁、内心的恐惧、信任的考验等等。

通过一系列精心设计的情节，作品展现了主角从一个普通人成长为英雄的过程。每一个转折点都推动着故事的发展，每一个选择都影响着最终的结局。

核心冲突围绕着主角与反派之间的对抗展开，同时也探讨了更深层的主题：权力与责任、勇气与恐惧、爱与牺牲。

作品采用三幕式结构，节奏紧凑，情节跌宕起伏，既有扣人心弦的动作场面，也有细腻动人的情感描写。`, project.Name)

	synopsis.MainPlot = fmt.Sprintf("主角从%s出发，经历一系列冒险，最终战胜强敌，实现自我成长。", project.Name)
	synopsis.SubPlots = []string{
		"主角的成长线",
		"友情与背叛",
		"爱情线索",
	}
	synopsis.CoreConflict = "主角与反派之间的生死对决"
	synopsis.Resolution = "主角最终获得胜利，但也付出了代价"

	synopsis.StructureType = "三幕式"
	synopsis.KeyEvents = []string{
		"开篇：介绍主角和世界",
		"激励事件：主角踏上冒险之路",
		"情节上升：一系列挑战和成长",
		"高潮：最终对决",
		"结局：新的开始",
	}
	synopsis.Themes = []string{"成长", "勇气", "友谊", "牺牲"}

	// 保存
	if err := h.db.SaveSynopsis(synopsis); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("SAVE_FAILED", "保存失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"synopsis": synopsis,
		"message":  "简介设定生成成功！",
	}))
}

// GachaSynopsisRequest 抽卡生成简介请求
type GachaSynopsisRequest struct {
	Tone string `json:"tone"` // 作品基调
}
