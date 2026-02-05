// Package handlers HTTP处理器
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/pkg/config"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/llm"
)

// WriterHandler 写作器处理器
type WriterHandler struct {
	db  db.Database
	cfg *config.Config
}

// NewWriterHandler 创建写作器处理器
func NewWriterHandler(database db.Database) *WriterHandler {
	cfg, err := config.LoadDefault()
	if err != nil {
		cfg = &config.Config{}
	}

	return &WriterHandler{
		db:  database,
		cfg: cfg,
	}
}

// ContinueChapterRequest 继续章节请求
// ContinueChapterRequest 继续章节请求
type ContinueChapterRequest struct {
	Length             string `json:"length"` // short, medium, long
	Style              string `json:"style"`  // balanced, creative, formal
	IncludeDialogue    bool   `json:"include_dialogue"`
	IncludeAction      bool   `json:"include_action"`
	IncludeDescription bool   `json:"include_description"`
	ContinueCount      int    `json:"continue_count"` // 继续次数
	Instructions       string `json:"instructions"`   // 用户指令
	WordCount          int    `json:"word_count"`     // 目标字数
}

// ContinueChapter AI继续章节内容
// @Summary AI继续章节内容
// @Description 基于当前章节内容、世界设定和角色信息，使用AI继续生成章节内容
// @Tags writer
// @Accept json
// @Produce json
// @Param project_id path string true "项目ID"
// @Param chapter_id path string true "章节ID"
// @Param request body ContinueChapterRequest true "继续参数"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{project_id}/chapters/{chapter_id}/continue [post]
func (h *WriterHandler) ContinueChapter(c *gin.Context) {
	projectID := c.Param("projectId")
	chapterID := c.Param("chapterId")

	// 检查项目是否存在
	project, err := h.db.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	// 获取章节
	chapter, err := h.db.GetChapter(chapterID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
		return
	}

	// 验证章节是否属于该项目
	if chapter.ProjectID != projectID {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
		return
	}

	// 解析请求参数
	var req ContinueChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 使用默认值
		req.Length = "medium"
		req.Style = "balanced"
		req.IncludeDialogue = true
		req.IncludeAction = true
		req.IncludeDescription = true
		req.ContinueCount = 1
	}

	// 获取世界设定
	worldSettings, err := h.db.GetWorld(project.WorldID)
	if err != nil {
		// 如果没有世界设定，使用空的世界设定
		worldSettings = &models.WorldSetting{ID: project.WorldID}
	}

	// 获取角色列表
	characters := h.db.ListCharactersByWorld(project.WorldID)

	// 获取叙事蓝图（如果有细纲则使用）
	blueprint, _ := h.db.GetNarrativeBlueprint(projectID)

	// 调用AI生成继续内容
	generatedText, err := h.generateContinuation(project, chapter, worldSettings, characters, blueprint, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("GENERATION_ERROR", "生成内容失败", err.Error()))
		return
	}

	// 更新章节内容
	newContent := chapter.Content + generatedText
	chapter.Content = newContent
	chapter.WordCount = utf8.RuneCountInString(newContent)
	chapter.AIWordCount += utf8.RuneCountInString(generatedText)

	if err := h.db.SaveChapter(chapter); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "保存章节失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"chapter": gin.H{
			"id":               chapter.ID,
			"chapter_num":      chapter.ChapterNum,
			"title":            chapter.Title,
			"content":          chapter.Content,
			"word_count":       chapter.WordCount,
			"ai_word_count":    chapter.AIWordCount,
			"generated":        generatedText,
			"generated_length": utf8.RuneCountInString(generatedText),
		},
	}))
}

// ContinueChapterStream 流式AI继续章节内容
// @Summary AI继续章节内容(流式)
// @Description 基于当前章节内容、世界设定和角色信息，使用AI继续生成章节内容，流式返回
// @Tags writer
// @Accept json
// @Produce text/event-stream
// @Param project_id path string true "项目ID"
// @Param chapter_id path string true "章节ID"
// @Param request body ContinueChapterRequest true "继续参数"
// @Success 200 {string} string "stream data"
// @Router /api/v1/projects/{project_id}/chapters/{chapter_id}/continue-stream [post]
func (h *WriterHandler) ContinueChapterStream(c *gin.Context) {
	projectID := c.Param("projectId")
	chapterID := c.Param("chapterId")

	// 检查项目
	project, err := h.db.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	// 获取章节
	chapter, err := h.db.GetChapter(chapterID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
		return
	}

	if chapter.ProjectID != projectID {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
		return
	}

	var req ContinueChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Length = "medium"
		req.Style = "balanced"
		req.IncludeDialogue = true
		req.IncludeAction = true
		req.IncludeDescription = true
		req.ContinueCount = 1
	}

	// 获取相关数据
	worldSettings, err := h.db.GetWorld(project.WorldID)
	if err != nil {
		worldSettings = &models.WorldSetting{ID: project.WorldID}
	}
	characters := h.db.ListCharactersByWorld(project.WorldID)
	blueprint, _ := h.db.GetNarrativeBlueprint(projectID)

	// 设置SSE Header
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 准备回调
	var fullContent strings.Builder

	err = h.generateContinuationStream(project, chapter, worldSettings, characters, blueprint, req, func(content string) bool {
		// 检查客户端是否断开
		select {
		case <-c.Request.Context().Done():
			return false
		default:
		}

		fullContent.WriteString(content)

		// 构造SSE消息
		// data: {"content": "..."}
		msg := gin.H{"content": content}
		msgBytes, _ := json.Marshal(msg)

		fmt.Fprintf(c.Writer, "data: %s\n\n", msgBytes)
		c.Writer.Flush()
		return true
	})

	if err != nil {
		// 如果只写了一部分，可能没法完美报错，但SSE通常在流中断前发送error event
		errMsg := gin.H{"error": err.Error()}
		errBytes, _ := json.Marshal(errMsg)
		fmt.Fprintf(c.Writer, "data: %s\n\n", errBytes)
		c.Writer.Flush()
		return
	}

	// 发送结束标记
	fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
	c.Writer.Flush()

	// 保存到数据库
	generatedText := fullContent.String()
	// 清理可能的markdown标记（虽然流式传输可能已经发出去了，但保存时清理一下）
	// 注意：流式传输给前端的是原始内容，前端自己处理展示。保存到数据库的最好也是干净的。
	// 这里简单清理一下首尾空白
	// generatedText = strings.TrimSpace(generatedText)
	// (流式可能导致Trim破坏格式，暂时不Trim或者只Trim首部)

	newContent := chapter.Content + generatedText
	chapter.Content = newContent
	chapter.WordCount = utf8.RuneCountInString(newContent)
	chapter.AIWordCount += utf8.RuneCountInString(generatedText)

	h.db.SaveChapter(chapter)
}

// generateContinuationStream 流式生成
func (h *WriterHandler) generateContinuationStream(
	project *models.Project,
	chapter *models.Chapter,
	worldSettings *models.WorldSetting,
	characters []*models.Character,
	blueprint *models.NarrativeBlueprint,
	req ContinueChapterRequest,
	callback llm.StreamCallback,
) error {
	client, _, err := llm.NewClientForModule("writer_scene")
	if err != nil {
		return fmt.Errorf("创建LLM客户端失败: %w", err)
	}

	prompt := h.buildContinuationPrompt(project, chapter, worldSettings, characters, blueprint, req)
	systemPrompt := h.buildWriterSystemPrompt(req)

	return client.GenerateStream(prompt, systemPrompt, callback)
}

// generateContinuation 生成继续内容
func (h *WriterHandler) generateContinuation(
	project *models.Project,
	chapter *models.Chapter,
	worldSettings *models.WorldSetting,
	characters []*models.Character,
	blueprint *models.NarrativeBlueprint,
	req ContinueChapterRequest,
) (string, error) {
	// 创建LLM客户端
	client, _, err := llm.NewClientForModule("writer_scene")
	if err != nil {
		return "", fmt.Errorf("创建LLM客户端失败: %w", err)
	}

	// 构建提示词
	prompt := h.buildContinuationPrompt(project, chapter, worldSettings, characters, blueprint, req)
	systemPrompt := h.buildWriterSystemPrompt(req)

	// 调用LLM
	result, err := client.GenerateWithParams(prompt, systemPrompt, 0.8, 3000)
	if err != nil {
		return "", fmt.Errorf("LLM调用失败: %w", err)
	}

	// 清理生成的文本
	result = strings.TrimSpace(result)
	// 移除可能的markdown代码块标记
	if strings.HasPrefix(result, "```") {
		lines := strings.Split(result, "\n")
		if len(lines) > 1 {
			// 查找第一个非代码块行
			startIdx := 0
			for i, line := range lines {
				if !strings.HasPrefix(line, "```") && line != "" {
					startIdx = i
					break
				}
			}
			// 查找结束的代码块
			endIdx := len(lines)
			for i := startIdx; i < len(lines); i++ {
				if strings.HasPrefix(lines[i], "```") {
					endIdx = i
					break
				}
			}
			if startIdx < endIdx {
				result = strings.Join(lines[startIdx:endIdx], "\n")
			}
		}
	}

	return result, nil
}

// buildContinuationPrompt 构建继续写作提示词
func (h *WriterHandler) buildContinuationPrompt(
	project *models.Project,
	chapter *models.Chapter,
	worldSettings *models.WorldSetting,
	characters []*models.Character,
	blueprint *models.NarrativeBlueprint,
	req ContinueChapterRequest,
) string {
	var prompt strings.Builder

	prompt.WriteString("# 继续写作任务\n\n")

	prompt.WriteString(fmt.Sprintf("## 作品信息\n"))
	prompt.WriteString(fmt.Sprintf("- 书名: %s\n", project.Name))
	prompt.WriteString(fmt.Sprintf("- 模式: %s\n", project.Mode))
	prompt.WriteString(fmt.Sprintf("- 章节: 第%d章 %s\n\n", chapter.ChapterNum, chapter.Title))

	// 当前内容摘要
	if chapter.Content != "" {
		contentRunes := []rune(chapter.Content)
		contentLength := len(contentRunes)
		// 获取最后500字作为上下文
		contextLength := 500
		if contentLength < contextLength {
			contextLength = contentLength
		}
		recentContent := string(contentRunes[contentLength-contextLength:])

		prompt.WriteString(fmt.Sprintf("## 当前内容（最后%d字）\n", contextLength))
		prompt.WriteString(recentContent)
		prompt.WriteString("\n\n")
	}

	// 章节细纲（如果有）
	if blueprint != nil && len(blueprint.ChapterPlans) > 0 {
		for _, plan := range blueprint.ChapterPlans {
			if plan.Chapter == chapter.ChapterNum {
				prompt.WriteString(fmt.Sprintf("## 章节细纲\n"))
				if plan.Purpose != "" {
					prompt.WriteString(fmt.Sprintf("- 本章目的: %s\n", plan.Purpose))
				}
				if len(plan.KeyScenes) > 0 {
					prompt.WriteString(fmt.Sprintf("- 关键场景: %s\n", strings.Join(plan.KeyScenes, "、")))
				}
				if plan.PlotAdvancement != "" {
					prompt.WriteString(fmt.Sprintf("- 情节推进: %s\n", plan.PlotAdvancement))
				}
				prompt.WriteString("\n")
				break
			}
		}
	}

	// 世界背景
	prompt.WriteString("## 世界背景\n")
	if len(worldSettings.StorySoil.SocialConflicts) > 0 {
		conflicts := make([]string, 0, len(worldSettings.StorySoil.SocialConflicts))
		for _, c := range worldSettings.StorySoil.SocialConflicts {
			if c.Description != "" {
				conflicts = append(conflicts, c.Description)
			}
		}
		if len(conflicts) > 0 {
			prompt.WriteString(fmt.Sprintf("- 核心冲突: %s\n", strings.Join(conflicts, "、")))
		}
	}
	if len(worldSettings.StorySoil.PotentialPlotHooks) > 0 {
		prompt.WriteString(fmt.Sprintf("- 情节钩子: %d个\n", len(worldSettings.StorySoil.PotentialPlotHooks)))
	}
	prompt.WriteString("\n")

	// 角色信息
	if len(characters) > 0 {
		prompt.WriteString("## 角色信息\n")
		for _, char := range characters {
			if char.Name != "" {
				prompt.WriteString(fmt.Sprintf("- %s", char.Name))
				if char.StaticProfile.Occupation != "" {
					prompt.WriteString(fmt.Sprintf("（%s）", char.StaticProfile.Occupation))
				}
				if char.DynamicState.Emotion.Current != "" {
					prompt.WriteString(fmt.Sprintf(" - 当前情绪: %s", char.DynamicState.Emotion.Current))
				}
				prompt.WriteString("\n")
			}
		}
		prompt.WriteString("\n")
	}

	// 生成要求
	prompt.WriteString("## 生成要求\n")

	// 字数处理
	targetLength := 800
	if req.WordCount > 0 {
		targetLength = req.WordCount
	} else {
		switch req.Length {
		case "short":
			targetLength = 500
		case "long":
			targetLength = 1500
		}
	}
	prompt.WriteString(fmt.Sprintf("- 目标字数: 约%d字\n", targetLength))

	styleDesc := "平衡"
	switch req.Style {
	case "creative":
		styleDesc = "创意丰富，语言生动"
	case "formal":
		styleDesc = "正式严肃，叙述严谨"
	}
	prompt.WriteString(fmt.Sprintf("- 风格: %s\n", styleDesc))

	if req.IncludeDialogue {
		prompt.WriteString("- 包含对话\n")
	}
	if req.IncludeAction {
		prompt.WriteString("- 包含动作描写\n")
	}
	if req.IncludeDescription {
		prompt.WriteString("- 包含环境/心理描写\n")
	}

	// 用户额外指令
	if req.Instructions != "" {
		prompt.WriteString(fmt.Sprintf("- 特别指令: %s\n", req.Instructions))
	}

	prompt.WriteString("\n")
	prompt.WriteString("请根据以上信息，自然地继续撰写本章内容。保持文风连贯，角色性格一致。\n\n")
	prompt.WriteString("只返回继续的文本内容，不要包含任何说明或注释。")

	return prompt.String()
}

// buildWriterSystemPrompt 构建写作系统提示词
func (h *WriterHandler) buildWriterSystemPrompt(req ContinueChapterRequest) string {
	var prompt strings.Builder

	prompt.WriteString("你是一位专业的小说作家，擅长续写引人入胜的叙事内容。\n\n")
	prompt.WriteString("# 写作原则\n")
	prompt.WriteString("1. 展示而非讲述（Show, Don't Tell）\n")
	prompt.WriteString("2. 通过动作和对话展现角色性格\n")
	prompt.WriteString("3. 保持叙事连贯性和节奏感\n")
	prompt.WriteString("4. 注意感官描写的层次感\n")
	prompt.WriteString("5. 对话要推动情节或揭示角色\n")
	prompt.WriteString("6. 自然衔接现有内容，不突兀\n\n")

	switch req.Style {
	case "creative":
		prompt.WriteString("# 风格要求\n")
		prompt.WriteString("- 使用丰富的修辞和比喻\n")
		prompt.WriteString("- 注重氛围营造和情感描写\n")
		prompt.WriteString("- 语言富有张力\n\n")
	case "formal":
		prompt.WriteString("# 风格要求\n")
		prompt.WriteString("- 语言简洁准确\n")
		prompt.WriteString("- 叙述严谨客观\n")
		prompt.WriteString("- 避免过度修饰\n\n")
	default:
		prompt.WriteString("# 风格要求\n")
		prompt.WriteString("- 叙述与对话平衡\n")
		prompt.WriteString("- 节奏适中\n")
		prompt.WriteString("- 语言自然流畅\n\n")
	}

	return prompt.String()
}

// GenerateChapterOutline 生成章节细纲
// @Summary 生成章节细纲
// @Description 为指定章节生成详细的场景级细纲
// @Tags writer
// @Produce json
// @Param project_id path string true "项目ID"
// @Param chapter_id path string true "章节ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{project_id}/chapters/{chapter_id}/outline [get]
func (h *WriterHandler) GenerateChapterOutline(c *gin.Context) {
	projectID := c.Param("projectId")
	chapterID := c.Param("chapterId")

	// 检查项目是否存在
	project, err := h.db.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	// 获取章节信息以获取章节号
	chapter, err := h.db.GetChapter(chapterID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "章节不存在", ""))
		return
	}

	// 获取世界设定
	worldSettings, err := h.db.GetWorld(project.WorldID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "世界设定不存在", ""))
		return
	}

	// 获取叙事蓝图
	blueprint, err := h.db.GetNarrativeBlueprint(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "叙事蓝图不存在，请先生成故事规划", ""))
		return
	}

	// 查找对应章节的计划
	var targetPlan *models.ChapterPlan
	for i := range blueprint.ChapterPlans {
		if blueprint.ChapterPlans[i].Chapter == chapter.ChapterNum {
			targetPlan = &blueprint.ChapterPlans[i]
			break
		}
	}

	if targetPlan == nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "该章节的计划不存在", ""))
		return
	}

	// 检查是否已经有该章节的场景指令
	var existingScenes []models.SceneInstruction
	for _, s := range blueprint.Scenes {
		if s.Chapter == chapter.ChapterNum {
			existingScenes = append(existingScenes, s)
		}
	}
	if len(existingScenes) > 0 {
		c.JSON(http.StatusOK, successResponse(gin.H{
			"chapter":    targetPlan.Chapter,
			"title":      targetPlan.Title,
			"purpose":    targetPlan.Purpose,
			"key_scenes": targetPlan.KeyScenes,
			"scenes":     existingScenes,
		}))
		return
	}

	// 生成场景指令
	scenes, err := h.generateSceneInstructions(project, worldSettings, blueprint, targetPlan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("GENERATION_ERROR", "生成场景指令失败", err.Error()))
		return
	}

	// 更新蓝图 - 添加场景到 Scenes
	for i := range scenes {
		scenes[i].Chapter = targetPlan.Chapter
	}
	blueprint.Scenes = append(blueprint.Scenes, scenes...)
	h.db.SaveNarrativeBlueprint(blueprint)

	c.JSON(http.StatusOK, successResponse(gin.H{
		"chapter":    targetPlan.Chapter,
		"title":      targetPlan.Title,
		"purpose":    targetPlan.Purpose,
		"key_scenes": targetPlan.KeyScenes,
		"scenes":     scenes,
	}))
}

// generateSceneInstructions 生成场景指令
func (h *WriterHandler) generateSceneInstructions(
	project *models.Project,
	worldSettings *models.WorldSetting,
	blueprint *models.NarrativeBlueprint,
	plan *models.ChapterPlan,
) ([]models.SceneInstruction, error) {
	client, _, err := llm.NewClientForModule("narrative_engine")
	if err != nil {
		return nil, fmt.Errorf("创建LLM客户端失败: %w", err)
	}

	prompt := h.buildScenePrompt(project, worldSettings, blueprint, plan)
	systemPrompt := `你是专业的小说场景设计师，负责将章节规划拆解为详细的场景级写作指令。

输出格式（JSON）：
{
  "scenes": [
    {
      "sequence": 1,
      "purpose": "场景目的",
      "location": "地点",
      "time": "时间",
      "pov_character": "视角角色",
      "scene_type": "场景类型（action/dialogue/description/inner_monologue/transition）",
      "mood": "氛围",
      "characters": ["角色ID列表"],
      "action": "主要动作",
      "dialogue_focus": "对话重点",
      "sensory_focus": {
        "visual": "视觉焦点",
        "auditory": "听觉焦点",
        "tactile": "触觉焦点"
      },
      "expected_length": 800
    }
  ]
}

只返回JSON，不要包含其他内容。`

	result, err := client.GenerateJSONWithParams(prompt, systemPrompt, 0.7, 3000)
	if err != nil {
		return nil, err
	}

	// 解析结果
	scenes := make([]models.SceneInstruction, 0)
	if scenesData, ok := result["scenes"].([]interface{}); ok {
		sceneNum := 1
		for _, s := range scenesData {
			if sceneMap, ok := s.(map[string]interface{}); ok {
				scene := models.SceneInstruction{
					Scene:          sceneNum,
					Sequence:       parseIntField(sceneMap, "sequence", sceneNum),
					Purpose:        parseStringField(sceneMap, "purpose", ""),
					Location:       parseStringField(sceneMap, "location", ""),
					POVCharacter:   parseStringField(sceneMap, "pov_character", ""),
					Mood:           parseStringField(sceneMap, "mood", ""),
					Action:         parseStringField(sceneMap, "action", ""),
					DialogueFocus:  parseStringField(sceneMap, "dialogue_focus", ""),
					ExpectedLength: parseIntField(sceneMap, "expected_length", 800),
					Status:         "pending",
				}
				if chars, ok := sceneMap["characters"].([]interface{}); ok {
					for _, c := range chars {
						if charID, ok := c.(string); ok {
							scene.Characters = append(scene.Characters, charID)
						}
					}
				}
				scenes = append(scenes, scene)
				sceneNum++
			}
		}
	}

	return scenes, nil
}

// buildScenePrompt 构建场景生成提示词
func (h *WriterHandler) buildScenePrompt(
	project *models.Project,
	worldSettings *models.WorldSetting,
	blueprint *models.NarrativeBlueprint,
	plan *models.ChapterPlan,
) string {
	var prompt strings.Builder

	prompt.WriteString("# 场景指令生成任务\n\n")
	prompt.WriteString(fmt.Sprintf("## 作品信息\n"))
	prompt.WriteString(fmt.Sprintf("- 书名: %s\n", project.Name))
	prompt.WriteString(fmt.Sprintf("- 模式: %s\n\n", project.Mode))

	prompt.WriteString(fmt.Sprintf("## 章节信息\n"))
	prompt.WriteString(fmt.Sprintf("- 章节: 第%d章 %s\n", plan.Chapter, plan.Title))
	prompt.WriteString(fmt.Sprintf("- 本章目的: %s\n", plan.Purpose))
	if len(plan.KeyScenes) > 0 {
		prompt.WriteString(fmt.Sprintf("- 关键场景: %s\n", strings.Join(plan.KeyScenes, "、")))
	}
	if plan.PlotAdvancement != "" {
		prompt.WriteString(fmt.Sprintf("- 情节推进: %s\n", plan.PlotAdvancement))
	}
	prompt.WriteString("\n")

	// 获取前一章的摘要
	for _, p := range blueprint.ChapterPlans {
		if p.Chapter == plan.Chapter-1 {
			// 检查是否有对应的场景
			hasScene := false
			for _, s := range blueprint.Scenes {
				if s.Chapter == p.Chapter {
					hasScene = true
					break
				}
			}
			if hasScene {
				prompt.WriteString("## 前一章概况\n")
				prompt.WriteString(fmt.Sprintf("- 目的: %s\n", p.Purpose))
				if p.ArcProgress != "" {
					prompt.WriteString(fmt.Sprintf("- 角色弧线: %s\n", p.ArcProgress))
				}
				prompt.WriteString("\n")
				break
			}
		}
	}

	prompt.WriteString("## 要求\n")
	prompt.WriteString("1. 将本章拆解为3-6个场景\n")
	prompt.WriteString("2. 每个场景要有明确的目的和推进\n")
	prompt.WriteString("3. 场景之间要有自然的过渡\n")
	prompt.WriteString("4. 合理分配POV角色\n")
	prompt.WriteString("5. 控制总字数在合理范围\n\n")

	return prompt.String()
}

// 辅助函数
func parseChapterNum(s string) int {
	num := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			num = num*10 + int(r-'0')
		}
	}
	return num
}

func parseIntField(m map[string]interface{}, key string, defaultVal int) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		case int64:
			return int(val)
		}
	}
	return defaultVal
}

func parseStringField(m map[string]interface{}, key string, defaultVal string) string {
	if v, ok := m[key]; ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return defaultVal
}
