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

// CharacterHandler 角色处理器
type CharacterHandler struct {
	db db.Database
}

// NewCharacterHandler 创建角色处理器
func NewCharacterHandler(database db.Database) *CharacterHandler {
	return &CharacterHandler{
		db: database,
	}
}

// GachaCharacters 抽卡生成角色设定
// @Summary 抽卡生成角色设定
// @Description AI抽卡生成主要角色档案
// @Tags characters
// @Accept json
// @Produce json
// @Param projectId path string true "项目ID"
// @Param request body GachaCharactersRequest true "生成参数"
// @Success 200 {object} APIResponse
// @Router /api/v1/projects/{projectId}/characters/gacha [post]
func (h *CharacterHandler) GachaCharacters(c *gin.Context) {
	projectID := c.Param("projectId")

	// 验证项目存在
	project, err := h.db.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "项目不存在", ""))
		return
	}

	var req GachaCharactersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", "请求参数错误", err.Error()))
		return
	}

	// 获取世界设定（如果存在）
	var world *models.WorldSetting
	if project.WorldID != "" {
		world, _ = h.db.GetWorld(project.WorldID)
	}

	// 生成主角
	mainCharacter := h.generateMainCharacter(project, world, req.ProtagonistName)

	// 生成2-3个主要配角
	var supportingCharacters []models.Character
	characterCount := req.CharacterCount
	if characterCount < 2 {
		characterCount = 2
	}
	if characterCount > 3 {
		characterCount = 3
	}

	for i := 0; i < characterCount; i++ {
		char := h.generateSupportingCharacter(project, world, i)
		supportingCharacters = append(supportingCharacters, char)
	}

	// 保存所有角色
	var characterIDs []string
	characterIDs = append(characterIDs, mainCharacter.ID)

	h.db.SaveCharacter(mainCharacter)
	for _, char := range supportingCharacters {
		h.db.SaveCharacter(&char)
		characterIDs = append(characterIDs, char.ID)
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"character_ids":          characterIDs,
		"main_character":         mainCharacter,
		"supporting_characters": supportingCharacters,
		"message":                fmt.Sprintf("成功生成%d个角色档案", len(characterIDs)),
	}))
}

// generateMainCharacter 生成主角
func (h *CharacterHandler) generateMainCharacter(project *models.Project, world *models.WorldSetting, name string) *models.Character {
	if name == "" {
		name = "主角"
	}

	character := &models.Character{
		ID:        db.GenerateID("char"),
		WorldID:   project.WorldID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 如果有世界设定，基于世界设定生成
	if world != nil {
		character.StaticProfile = models.StaticProfile{
			Background:   fmt.Sprintf("出生于%s世界", world.Type),
			Race:         "人类",
			Age:          18,
			Gender:       "男",
			Appearance:   "平凡外表，但眼中透着坚毅",
			Abilities:    []string{"学习能力", "适应能力"},
			SocialStatus: "普通平民",
			Occupation:   "学生",
		}

		character.NarrativeProfile = models.NarrativeProfile{
			Personality: []models.Trait{
				{Name: "勇敢", Category: "positive", Intensity: 70},
				{Name: "善良", Category: "positive", Intensity: 60},
			},
			Motivation: models.Motivation{
				CoreNeed:      "寻找归属感",
				ExternalGoal: "成为强者",
				InnerConflict: "渴望力量但害怕失去人性",
			},
			Flaw:       "过于冲动",
			Fear:       "失去重要的人",
			BeliefSystem: models.BeliefSystem{
				WorldView:   "世界充满可能",
				HumanNature: "人性本善",
				Morality:    []string{"诚实", "正义", "保护弱小"},
			},
		}
	} else {
		// 默认主角设定
		character.StaticProfile = models.StaticProfile{
			Background:   "神秘出身",
			Race:         "人类",
			Age:          20,
			Gender:       "男",
			Appearance:   "英俊外表，气质不凡",
			Abilities:    []string{"天赋异禀"},
			SocialStatus: "未知",
			Occupation:   "冒险者",
		}

		character.NarrativeProfile = models.NarrativeProfile{
			Personality: []models.Trait{
				{Name: "果断", Category: "positive", Intensity: 80},
			},
			Motivation: models.Motivation{
				CoreNeed:      "自我实现",
				ExternalGoal: "探索世界",
				InnerConflict: "追求自由与责任的平衡",
			},
			Flaw:       "过于自信",
			Fear:       "失败",
			BeliefSystem: models.BeliefSystem{
				WorldView:   "世界充满机遇",
				HumanNature: "人性复杂",
				Morality:    []string{"理性", "实用主义"},
			},
		}
	}

	return character
}

// generateSupportingCharacter 生成配角
func (h *CharacterHandler) generateSupportingCharacter(project *models.Project, world *models.WorldSetting, index int) models.Character {
	roles := []string{"导师", "伙伴", "对手"}
	role := "伙伴"
	if index < len(roles) {
		role = roles[index]
	}

	gender := "男"
	if index%2 != 0 {
		gender = "女"
	}

	character := models.Character{
		ID:        db.GenerateID("char"),
		WorldID:   project.WorldID,
		Name:      fmt.Sprintf("%s%d", role, index+1),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	character.StaticProfile = models.StaticProfile{
		Background:   fmt.Sprintf("作为一名%s，有着丰富的经历", role),
		Race:         "人类",
		Age:          20 + index*5,
		Gender:       gender,
		Appearance:   "气质独特",
		Abilities:    []string{"经验丰富"},
		SocialStatus: "受人尊敬",
		Occupation:   role,
	}

	character.NarrativeProfile = models.NarrativeProfile{
		Personality: []models.Trait{
			{Name: "智慧", Category: "positive", Intensity: 70},
			{Name: "忠诚", Category: "positive", Intensity: 80},
		},
		Motivation: models.Motivation{
			CoreNeed:      "成就感",
			ExternalGoal: "帮助主角",
			InnerConflict: "责任与个人愿望的冲突",
		},
		Flaw:         "固执",
		Fear:         "被抛弃",
		BeliefSystem: models.BeliefSystem{
			WorldView:   "世界有秩序",
			HumanNature: "人性可塑",
			Morality:    []string{"责任", "荣誉"},
		},
	}

	return character
}

// GachaCharactersRequest 抽卡生成角色请求
type GachaCharactersRequest struct {
	ProtagonistName string `json:"protagonist_name"`
	CharacterCount  int    `json:"character_count"`
}
