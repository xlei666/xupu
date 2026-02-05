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

// SyncConfigs 同步默认系统配置
func (h *AdminHandler) SyncConfigs(c *gin.Context) {
	defaultConfigs := []models.SysConfig{
		{
			Key:         "llm_provider",
			Value:       "openai",
			Type:        "string",
			Description: "默认LLM提供商 (openai, anthropic, gemini)",
			Group:       "llm",
		},
		{
			Key:         "default_model",
			Value:       "gpt-4o",
			Type:        "string",
			Description: "默认使用的模型名称",
			Group:       "llm",
		},
		{
			Key:         "max_tokens",
			Value:       "4096",
			Type:        "int",
			Description: "单次生成最大Token数",
			Group:       "llm",
		},
		{
			Key:         "temperature",
			Value:       "0.7",
			Type:        "float",
			Description: "默认随机性 (0.0 - 1.0)",
			Group:       "llm",
		},
		{
			Key:         "enable_search",
			Value:       "false",
			Type:        "bool",
			Description: "是否启用联网搜索增强",
			Group:       "feature",
		},
		{
			Key:         "auth_registration_open",
			Value:       "true",
			Type:        "bool",
			Description: "是否开放用户注册",
			Group:       "system",
		},
	}

	syncedCount := 0
	for _, rawCfg := range defaultConfigs {
		existing, _ := h.db.GetSysConfig(rawCfg.Key)
		if existing == nil {
			cfg := rawCfg // Copy
			cfg.UpdatedAt = time.Now()
			if err := h.db.SaveSysConfig(&cfg); err == nil {
				syncedCount++
			}
		}
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"message":      "系统配置同步完成",
		"synced_count": syncedCount,
	}))
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

// SyncStructures 同步默认叙事结构
func (h *AdminHandler) SyncStructures(c *gin.Context) {
	templates := []models.NarrativeTemplate{
		{
			ID:          "three_act",
			Name:        "三幕剧结构 (Three Act Structure)",
			Description: "最经典的故事结构，由铺垫、冲突和解决三个部分组成。",
			Structure: models.JSON(`{
				"stages": [
					{ "name": "第一幕：铺垫", "beats": ["建置 (Setup)", "激励事件 (Inciting Incident)", "第一情节点 (Plot Point 1)"] },
					{ "name": "第二幕：冲突", "beats": ["上升动作 (Rising Action)", "中点 (Midpoint)", "一无所有 (All Is Lost)", "第二情节点 (Plot Point 2)"] },
					{ "name": "第三幕：结局", "beats": ["高潮 (Climax)", "结局 (Resolution)"] }
				]
			}`),
			PromptRules: models.JSON(`{
				"system_prompt": "你是一位精通三幕剧结构的编剧。",
				"beat_prompts": {
					"Setup": "描述主角的现状...",
					"Inciting Incident": "发生了一个打破平衡的事件..."
				}
			}`),
		},
		{
			ID:          "heros_journey",
			Name:        "英雄之旅 (Hero's Journey)",
			Description: "约瑟夫·坎贝尔提出的神话叙事结构，适用于冒险故事。",
			Structure: models.JSON(`{
				"stages": [
					{ "name": "第一阶段：启程", "beats": ["平凡世界", "冒险召唤", "拒绝召唤", "遇见导师", "跨越第一道门槛"] },
					{ "name": "第二阶段：启蒙", "beats": ["试炼、盟友、敌人", "接近最深处的洞穴", "严峻考验", "奖赏"] },
					{ "name": "第三阶段：归来", "beats": ["归途", "复活", "带着灵药回归"] }
				]
			}`),
		},
		{
			ID:          "save_the_cat",
			Name:        "救猫咪 (Save The Cat)",
			Description: "布莱克·斯奈德提出的商业剧本结构，节奏紧凑。",
			Structure: models.JSON(`{
				"stages": [
					{ "name": "第一幕", "beats": ["开篇画面", "主题陈述", "铺垫", "触发事件", "争论", "第二幕衔接点"] },
					{ "name": "第二幕", "beats": ["B故事", "游戏时间", "中点", "坏人逼近", "一无所有", "灵魂黑夜", "第三幕衔接点"] },
					{ "name": "第三幕", "beats": ["终局", "结束画面"] }
				]
			}`),
		},
		{
			ID:          "kishotenketsu",
			Name:        "起承转合 (Kishotenketsu)",
			Description: "源自中国诗歌的传统叙事结构，强调转折而非冲突。",
			Structure: models.JSON(`{
				"stages": [
					{ "name": "起 (Ki)", "beats": ["介绍背景", "引入角色"] },
					{ "name": "承 (Sho)", "beats": ["发展情节", "深化关系"] },
					{ "name": "转 (Ten)", "beats": ["意外转折", "视角转换"] },
					{ "name": "合 (Ketsu)", "beats": ["收束线索", "余韵悠长"] }
				]
			}`),
		},
	}

	syncedCount := 0
	for _, rawTmpl := range templates {
		existing, _ := h.db.GetNarrativeTemplate(rawTmpl.ID)
		if existing == nil {
			tmpl := rawTmpl // Copy
			tmpl.CreatedAt = time.Now()
			tmpl.UpdatedAt = time.Now()
			// Default active
			tmpl.IsActive = true
			if err := h.db.SaveNarrativeTemplate(&tmpl); err == nil {
				syncedCount++
			}
		}
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"message":      "结构模版同步完成",
		"synced_count": syncedCount,
	}))
}
