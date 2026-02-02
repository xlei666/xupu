// Package handlers HTTP处理器 - 导出功能
package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/db"
)

// ExportHandler 导出处理器
type ExportHandler struct{}

// NewExportHandler 创建导出处理器
func NewExportHandler() *ExportHandler {
	return &ExportHandler{}
}

// ExportProject 导出项目
// @Summary 导出项目
// @Description 将项目导出为指定格式
// @Tags export
// @Produce json, plain, markdown
// @Param id path string true "项目ID"
// @Param format query string false "导出格式" Enums(json, markdown, txt)
// @Success 200 {object} APIResponse
// @Router /api/v1/export/project/{id} [get]
func (h *ExportHandler) ExportProject(c *gin.Context) {
	id := c.Param("id")
	format := c.DefaultQuery("format", "json")

	project, err := db.Get().GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	switch format {
	case "markdown", "md":
		h.exportProjectMarkdown(c, project)
	case "txt":
		h.exportProjectTxt(c, project)
	default:
		c.JSON(http.StatusOK, successResponse(toProjectResponse(project)))
	}
}

// ExportWorld 导出世界设定
// @Summary 导出世界设定
// @Description 将世界设定导出为指定格式
// @Tags export
// @Produce json, plain, markdown
// @Param id path string true "世界ID"
// @Param format query string false "导出格式" Enums(json, markdown, txt)
// @Success 200 {object} APIResponse
// @Router /api/v1/export/world/{id} [get]
func (h *ExportHandler) ExportWorld(c *gin.Context) {
	id := c.Param("id")
	format := c.DefaultQuery("format", "json")

	world, err := db.Get().GetWorld(id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "世界不存在", ""))
		return
	}

	switch format {
	case "markdown", "md":
		h.exportWorldMarkdown(c, world)
	case "txt":
		h.exportWorldTxt(c, world)
	default:
		c.JSON(http.StatusOK, successResponse(toWorldResponse(world)))
	}
}

// ExportBlueprint 导出叙事蓝图
// @Summary 导出叙事蓝图
// @Description 将叙事蓝图导出为指定格式
// @Tags export
// @Produce json, plain, markdown
// @Param id path string true "蓝图ID"
// @Param format query string false "导出格式" Enums(json, markdown, txt)
// @Success 200 {object} APIResponse
// @Router /api/v1/export/blueprint/{id} [get]
func (h *ExportHandler) ExportBlueprint(c *gin.Context) {
	id := c.Param("id")
	format := c.DefaultQuery("format", "json")

	blueprint, err := db.Get().GetNarrativeBlueprint(id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "蓝图不存在", ""))
		return
	}

	switch format {
	case "markdown", "md":
		h.exportBlueprintMarkdown(c, blueprint)
	case "txt":
		h.exportBlueprintTxt(c, blueprint)
	default:
		c.JSON(http.StatusOK, successResponse(toBlueprintResponse(blueprint)))
	}
}

// exportProjectMarkdown 导出项目为Markdown
func (h *ExportHandler) exportProjectMarkdown(c *gin.Context, p *models.Project) {
	var sb strings.Builder

	sb.WriteString("# ")
	sb.WriteString(p.Name)
	sb.WriteString("\n\n")

	if p.Description != "" {
		sb.WriteString(p.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## 项目信息\n\n")
	sb.WriteString("- **项目ID**: `")
	sb.WriteString(p.ID)
	sb.WriteString("`\n")
	sb.WriteString("- **模式**: ")
	sb.WriteString(string(p.Mode))
	sb.WriteString("\n")
	sb.WriteString("- **状态**: ")
	sb.WriteString(string(p.Status))
	sb.WriteString("\n")
	sb.WriteString("- **进度**: ")
	sb.WriteString(fmt.Sprintf("%.1f", p.Progress))
	sb.WriteString("%\n")

	if p.WorldID != "" {
		sb.WriteString("- **世界ID**: `")
		sb.WriteString(p.WorldID)
		sb.WriteString("`\n")
	}

	if p.NarrativeID != "" {
		sb.WriteString("- **叙事ID**: `")
		sb.WriteString(p.NarrativeID)
		sb.WriteString("`n")
	}

	sb.WriteString("- **创建时间**: ")
	sb.WriteString(p.CreatedAt.Format("2006-01-02 15:04:05"))
	sb.WriteString("\n")

	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.String(http.StatusOK, sb.String())
}

// exportProjectTxt 导出项目为纯文本
func (h *ExportHandler) exportProjectTxt(c *gin.Context, p *models.Project) {
	var sb strings.Builder

	sb.WriteString("========================================\n")
	sb.WriteString(p.Name)
	sb.WriteString("\n")
	sb.WriteString("========================================\n\n")

	if p.Description != "" {
		sb.WriteString(p.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("项目信息:\n")
	sb.WriteString("  项目ID: ")
	sb.WriteString(p.ID)
	sb.WriteString("\n")
	sb.WriteString("  模式: ")
	sb.WriteString(string(p.Mode))
	sb.WriteString("\n")
	sb.WriteString("  状态: ")
	sb.WriteString(string(p.Status))
	sb.WriteString("\n")

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, sb.String())
}

// exportWorldMarkdown 导出世界设定为Markdown
func (h *ExportHandler) exportWorldMarkdown(c *gin.Context, w *models.WorldSetting) {
	var sb strings.Builder

	sb.WriteString("# 世界设定: ")
	sb.WriteString(w.Name)
	sb.WriteString("\n\n")

	// 基本信息
	sb.WriteString("## 基本信息\n\n")
	sb.WriteString("- **类型**: ")
	sb.WriteString(string(w.Type))
	sb.WriteString("\n")
	sb.WriteString("- **规模**: ")
	sb.WriteString(string(w.Scale))
	sb.WriteString("\n")
	if w.Style != "" {
		sb.WriteString("- **风格**: ")
		sb.WriteString(w.Style)
		sb.WriteString("\n")
	}

	// 哲学设定
	sb.WriteString("\n## 哲学设定\n\n")
	if w.Philosophy.CoreQuestion != "" {
		sb.WriteString("### 核心问题\n")
		sb.WriteString(w.Philosophy.CoreQuestion)
		sb.WriteString("\n\n")
	}

	sb.WriteString("### 价值体系\n\n")
	if w.Philosophy.ValueSystem.HighestGood != "" {
		sb.WriteString("- **至善**: ")
		sb.WriteString(w.Philosophy.ValueSystem.HighestGood)
		sb.WriteString("\n")
	}
	if w.Philosophy.ValueSystem.UltimateEvil != "" {
		sb.WriteString("- **至恶**: ")
		sb.WriteString(w.Philosophy.ValueSystem.UltimateEvil)
		sb.WriteString("\n")
	}

	// 地理
	if len(w.Geography.Regions) > 0 {
		sb.WriteString("\n## 地理环境\n\n")
		for _, region := range w.Geography.Regions {
			sb.WriteString("### ")
			sb.WriteString(region.Name)
			sb.WriteString("\n")
			if region.Description != "" {
				sb.WriteString(region.Description)
				sb.WriteString("\n")
			}
		}
	}

	// 文明
	if len(w.Civilization.Races) > 0 {
		sb.WriteString("\n## 文明种族\n\n")
		for _, race := range w.Civilization.Races {
			sb.WriteString("### ")
			sb.WriteString(race.Name)
			sb.WriteString("\n")
			if race.Description != "" {
				sb.WriteString(race.Description)
				sb.WriteString("\n")
			}
		}
	}

	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.String(http.StatusOK, sb.String())
}

// exportWorldTxt 导出世界设定为纯文本
func (h *ExportHandler) exportWorldTxt(c *gin.Context, w *models.WorldSetting) {
	var sb strings.Builder

	sb.WriteString("========================================\n")
	sb.WriteString("世界设定: ")
	sb.WriteString(w.Name)
	sb.WriteString("\n")
	sb.WriteString("========================================\n\n")

	sb.WriteString("类型: ")
	sb.WriteString(string(w.Type))
	sb.WriteString("\n")
	sb.WriteString("规模: ")
	sb.WriteString(string(w.Scale))
	sb.WriteString("\n\n")

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, sb.String())
}

// exportBlueprintMarkdown 导出蓝图为Markdown
func (h *ExportHandler) exportBlueprintMarkdown(c *gin.Context, b *models.NarrativeBlueprint) {
	var sb strings.Builder

	sb.WriteString("# 叙事蓝图\n\n")

	// 故事大纲
	sb.WriteString("## 故事大纲\n\n")
	sb.WriteString("**结构类型**: ")
	sb.WriteString(b.StoryOutline.StructureType)
	sb.WriteString("\n\n")

	// 第一幕
	if b.StoryOutline.Act1.Setup != "" || b.StoryOutline.Act1.IncitingIncident != "" {
		sb.WriteString("### 第一幕\n\n")
		if b.StoryOutline.Act1.Setup != "" {
			sb.WriteString("**铺垫**: ")
			sb.WriteString(b.StoryOutline.Act1.Setup)
			sb.WriteString("\n\n")
		}
		if b.StoryOutline.Act1.IncitingIncident != "" {
			sb.WriteString("**激励事件**: ")
			sb.WriteString(b.StoryOutline.Act1.IncitingIncident)
			sb.WriteString("\n\n")
		}
		if b.StoryOutline.Act1.PlotPoint1 != "" {
			sb.WriteString("**情节点1**: ")
			sb.WriteString(b.StoryOutline.Act1.PlotPoint1)
			sb.WriteString("\n\n")
		}
	}

	// 第二幕
	if len(b.StoryOutline.Act2.RisingAction) > 0 || b.StoryOutline.Act2.Midpoint != "" {
		sb.WriteString("### 第二幕\n\n")
		for _, action := range b.StoryOutline.Act2.RisingAction {
			sb.WriteString("- ")
			sb.WriteString(action)
			sb.WriteString("\n")
		}
		if b.StoryOutline.Act2.Midpoint != "" {
			sb.WriteString("\n**中点**: ")
			sb.WriteString(b.StoryOutline.Act2.Midpoint)
			sb.WriteString("\n\n")
		}
	}

	// 第三幕
	if b.StoryOutline.Act3.Climax != "" || b.StoryOutline.Act3.Resolution != "" {
		sb.WriteString("### 第三幕\n\n")
		if b.StoryOutline.Act3.Climax != "" {
			sb.WriteString("**高潮**: ")
			sb.WriteString(b.StoryOutline.Act3.Climax)
			sb.WriteString("\n\n")
		}
		if b.StoryOutline.Act3.Resolution != "" {
			sb.WriteString("**结局**: ")
			sb.WriteString(b.StoryOutline.Act3.Resolution)
			sb.WriteString("\n\n")
		}
	}

	// 主题
	if b.ThemePlan.CoreTheme != "" {
		sb.WriteString("## 核心主题\n\n")
		sb.WriteString(b.ThemePlan.CoreTheme)
		sb.WriteString("\n\n")
	}

	// 章节规划
	if len(b.ChapterPlans) > 0 {
		sb.WriteString("## 章节规划\n\n")
		for _, chapter := range b.ChapterPlans {
			sb.WriteString("### 第")
			sb.WriteString(fmt.Sprintf("%d", chapter.Chapter))
			sb.WriteString("章: ")
			sb.WriteString(chapter.Title)
			sb.WriteString("\n\n")
			if chapter.Purpose != "" {
				sb.WriteString("**目的**: ")
				sb.WriteString(chapter.Purpose)
				sb.WriteString("\n\n")
			}
			if chapter.PlotAdvancement != "" {
				sb.WriteString("**情节推进**: ")
				sb.WriteString(chapter.PlotAdvancement)
				sb.WriteString("\n\n")
			}
		}
	}

	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.String(http.StatusOK, sb.String())
}

// exportBlueprintTxt 导出蓝图为纯文本
func (h *ExportHandler) exportBlueprintTxt(c *gin.Context, b *models.NarrativeBlueprint) {
	var sb strings.Builder

	sb.WriteString("========================================\n")
	sb.WriteString("叙事蓝图\n")
	sb.WriteString("========================================\n\n")

	sb.WriteString("结构类型: ")
	sb.WriteString(b.StoryOutline.StructureType)
	sb.WriteString("\n\n")

	if b.ThemePlan.CoreTheme != "" {
		sb.WriteString("核心主题:\n")
		sb.WriteString(b.ThemePlan.CoreTheme)
		sb.WriteString("\n\n")
	}

	if len(b.ChapterPlans) > 0 {
		sb.WriteString("章节总数: ")
		sb.WriteString(fmt.Sprintf("%d", len(b.ChapterPlans)))
		sb.WriteString("\n")
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, sb.String())
}
