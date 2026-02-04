package handlers

import (
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/config"
	"github.com/xlei/xupu/pkg/db"
)

type AdminHandler struct {
	db db.Database
}

func NewAdminHandler(database db.Database) *AdminHandler {
	return &AdminHandler{db: database}
}

// ============================================
// System Configs
// ============================================

// GetConfigs 获取所有系统配置
func (h *AdminHandler) GetConfigs(c *gin.Context) {
	configs, err := h.db.GetSysConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "获取配置失败", err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(configs))
}

// UpdateConfig 更新系统配置
func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	key := c.Param("key")
	var req models.SysConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "无效请求", err.Error()))
		return
	}

	// 确保Key一致
	req.Key = key

	if err := h.db.SaveSysConfig(&req); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "保存配置失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(req))
}

// ============================================
// Prompts
// ============================================

// GetPrompts 获取所有提示词模板
func (h *AdminHandler) GetPrompts(c *gin.Context) {
	prompts, err := h.db.GetPromptTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "获取提示词失败", err.Error()))
		return
	}
	// 按Key排序
	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].Key < prompts[j].Key
	})
	c.JSON(http.StatusOK, successResponse(prompts))
}

// GetPrompt 获取单个提示词模板
func (h *AdminHandler) GetPrompt(c *gin.Context) {
	key := c.Param("key")
	prompt, err := h.db.GetPromptTemplate(key)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "提示词不存在", ""))
		return
	}
	c.JSON(http.StatusOK, successResponse(prompt))
}

// UpdatePrompt 更新提示词模板
func (h *AdminHandler) UpdatePrompt(c *gin.Context) {
	key := c.Param("key")
	var req models.PromptTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "无效请求", err.Error()))
		return
	}

	req.Key = key

	// 如果是新建或更新，确保Version增加? 这里简单处理，用户手动控制版本或不控制
	// 获取旧版本以保留某些字段如果需要
	old, _ := h.db.GetPromptTemplate(key)
	if old != nil {
		req.Version = old.Version + 1
	} else {
		req.Version = 1
	}

	if err := h.db.SavePromptTemplate(&req); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "保存提示词失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(req))
}

// SyncFromConfig 从配置文件同步到数据库（初始化/重置）
func (h *AdminHandler) SyncFromConfig(c *gin.Context) {
	cfg := config.Get() // 获取全局配置
	allPrompts := cfg.GetAllPrompts()

	syncedCount := 0
	for key, content := range allPrompts {
		// 检查是否存在
		existing, _ := h.db.GetPromptTemplate(key)
		if existing == nil {
			// 不存在则创建
			newPrompt := &models.PromptTemplate{
				Key:         key,
				Content:     content,
				Description: "Synced from config.yaml",
				Version:     1,
				UpdatedAt:   time.Now(),
			}
			// 简单的变量推断（非必须，UI可稍后补充）
			// newPrompt.Variables = extractVariables(content)

			if err := h.db.SavePromptTemplate(newPrompt); err == nil {
				syncedCount++
			}
		}
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"message":      "同步完成",
		"synced_count": syncedCount,
	}))
}

// ============================================
// Narrative Templates
// ============================================

func (h *AdminHandler) GetStructures(c *gin.Context) {
	templates, err := h.db.GetNarrativeTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "获取模板失败", err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(templates))
}

func (h *AdminHandler) UpdateStructure(c *gin.Context) {
	id := c.Param("id")
	var req models.NarrativeTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "无效请求", err.Error()))
		return
	}
	req.ID = id

	if err := h.db.SaveNarrativeTemplate(&req); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", "保存模板失败", err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(req))
}
