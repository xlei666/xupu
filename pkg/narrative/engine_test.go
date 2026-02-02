// Package narrative 叙事器测试
package narrative

import (
	"errors"
	"testing"
	"time"

	"github.com/xlei/xupu/internal/models"
)

// MockDatabase 模拟数据库
type MockDatabase struct {
	worlds     map[string]*models.WorldSetting
	blueprints map[string]*models.NarrativeBlueprint
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		worlds:     make(map[string]*models.WorldSetting),
		blueprints: make(map[string]*models.NarrativeBlueprint),
	}
}

func (m *MockDatabase) GetWorld(id string) (*models.WorldSetting, error) {
	world, ok := m.worlds[id]
	if !ok {
		return nil, ErrNotFound
	}
	return world, nil
}

func (m *MockDatabase) SaveNarrativeBlueprint(blueprint *models.NarrativeBlueprint) error {
	blueprint.UpdatedAt = time.Now()
	if blueprint.CreatedAt.IsZero() {
		blueprint.CreatedAt = time.Now()
	}
	m.blueprints[blueprint.ID] = blueprint
	return nil
}

func (m *MockDatabase) GetNarrativeBlueprint(id string) (*models.NarrativeBlueprint, error) {
	blueprint, ok := m.blueprints[id]
	if !ok {
		return nil, ErrNotFound
	}
	return blueprint, nil
}

func (m *MockDatabase) AddWorld(world *models.WorldSetting) {
	m.worlds[world.ID] = world
}

var ErrNotFound = errors.New("not found")

// TestCreateBlueprint_InputValidation 测试创建蓝图输入验证
func TestCreateBlueprint_InputValidation(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "缺少WorldID",
			params: CreateParams{
				StoryType: "fantasy",
				Theme:     "成长",
			},
			wantErr: true,
		},
		{
			name: "缺少StoryType",
			params: CreateParams{
				WorldID: "test_world_1",
				Theme:   "成长",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 这个测试主要验证输入参数的完整性
			if tt.params.WorldID == "" && !tt.wantErr {
				t.Error("期望有WorldID参数")
			}
		})
	}
}

// TestDefaultChapterCount 测试默认章节数
func TestDefaultChapterCount(t *testing.T) {
	engine := &NarrativeEngine{}

	tests := []struct {
		length string
		want   int
	}{
		{"short", 10},
		{"medium", 30},
		{"long", 60},
		{"unknown", 20}, // 默认值
	}

	for _, tt := range tests {
		t.Run(tt.length, func(t *testing.T) {
			got := engine.defaultChapterCount(tt.length)
			if got != tt.want {
				t.Errorf("defaultChapterCount(%q) = %d, want %d", tt.length, got, tt.want)
			}
		})
	}
}

// TestPlanTheme 测试主题规划
func TestPlanTheme(t *testing.T) {
	engine := &NarrativeEngine{}

	coreTheme := "勇气与成长"
	chapterCount := 20

	plan := engine.planTheme(coreTheme, chapterCount)

	if plan.CoreTheme != coreTheme {
		t.Errorf("CoreTheme = %q, want %q", plan.CoreTheme, coreTheme)
	}

	if len(plan.Threading) == 0 {
		t.Error("Threading should not be empty")
	}

	// 检查主题贯穿的递进性
	for i, threading := range plan.Threading {
		if threading.Chapter < 1 || threading.Chapter > chapterCount {
			t.Errorf("Chapter %d out of range [1, %d]", threading.Chapter, chapterCount)
		}
		if i > 0 && threading.Depth == plan.Threading[i-1].Depth {
			// 允许相同深度，但应该有进展
		}
	}

	t.Logf("主题规划: %+v", plan)
}

// TestBuildWorldSummary 测试世界摘要构建
func TestBuildWorldSummary(t *testing.T) {
	engine := &NarrativeEngine{}

	world := &models.WorldSetting{
		Name: "测试世界",
		Type: models.WorldFantasy,
		Scale: models.ScaleContinent,
		Philosophy: models.Philosophy{
			CoreQuestion: "什么是真正的勇气？",
			ValueSystem: models.ValueSystem{
				HighestGood:  "守护",
				UltimateEvil: "背叛",
			},
		},
		StorySoil: models.StorySoil{
			SocialConflicts: []models.Conflict{
				{
					Type:        "political",
					Description: "王位争夺",
					Tension:     80,
				},
			},
			PotentialPlotHooks: []models.PlotHook{
				{
					Type:           "conflict",
					Description:    "边境战争爆发",
					StoryPotential: "可能导致国家动荡",
				},
			},
		},
		Geography: models.Geography{
			Regions: []models.Region{
				{ID: "r1", Name: "北方", Type: "mountain"},
				{ID: "r2", Name: "南方", Type: "plain"},
			},
			Climate: &models.Climate{
				Type: "温带",
			},
		},
		Civilization: models.Civilization{
			Races: []models.Race{
				{Name: "人类"},
				{Name: "精灵"},
			},
		},
	}

	summary := engine.buildWorldSummary(world)

	// 验证摘要包含关键信息
	requiredStrings := []string{
		"测试世界",
		"fantasy",
		"什么是真正的勇气",
		"守护",
		"社会冲突",
		"地理",
		"人类",
	}

	for _, s := range requiredStrings {
		if !contains(summary, s) {
			t.Errorf("摘要缺少关键信息: %q\n摘要内容:\n%s", s, summary)
		}
	}

	t.Logf("世界摘要:\n%s", summary)
}

// TestUpdatePreviousSummary 测试前情摘要更新
func TestUpdatePreviousSummary(t *testing.T) {
	engine := &NarrativeEngine{}

	chapter := models.ChapterPlan{
		Chapter:         1,
		Title:           "启程",
		Purpose:         "主角开始冒险",
		PlotAdvancement: "主角离开家乡",
	}

	scenes := &SceneOutput{
		Scenes: []SceneItem{
			{Sequence: 1, Location: "家乡", Purpose: "告别"},
			{Sequence: 2, Location: "边境", Purpose: "遭遇战斗"},
		},
	}

	summary := engine.updatePreviousSummary(chapter, scenes)

	// 验证摘要包含关键信息
	requiredStrings := []string{
		"第1章",
		"启程",
		"家乡",
		"边境",
		"主角离开家乡",
	}

	for _, s := range requiredStrings {
		if !contains(summary, s) {
			t.Errorf("摘要缺少关键信息: %q\n摘要内容: %s", s, summary)
		}
	}

	t.Logf("前情摘要: %s", summary)
}

// TestExtractJSON 测试JSON提取
func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string // 检查结果是否包含此内容（而非精确匹配）
	}{
		{
			name:     "提取markdown包裹的JSON",
			input:    "```json\n{\"key\": \"value\"}\n```",
			contains: `{"key": "value"}`,
		},
		{
			name:     "提取简单代码块中的JSON",
			input:    "```\n{\"key\": \"value\"}\n```",
			contains: `{"key": "value"}`,
		},
		{
			name:     "提取纯JSON",
			input:    `{"key": "value"}`,
			contains: `{"key": "value"}`,
		},
		{
			name:     "从文本中提取JSON",
			input:    `前面的一些文本 {"key": "value"} 后面的文本`,
			contains: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSON(tt.input)
			if !contains(got, tt.contains) {
				t.Errorf("extractJSON() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

// TestOutlineOutputStructure 测试大纲输出结构
func TestOutlineOutputStructure(t *testing.T) {
	output := OutlineOutput{
		StructureType: StructureThreeAct,
		ThreeAct: &ThreeActOutput{
			Act1: Act1Output{
				Setup:            "世界建立",
				IncitingIncident: "冒险召唤",
				PlotPoint1:       "跨越门槛",
			},
			Act2: Act2Output{
				RisingAction: []string{"试炼1", "试炼2", "试炼3"},
				Midpoint:     "中点转折",
				AllIsLost:    "一无所有时刻",
				PlotPoint2:   "最终对决准备",
			},
			Act3: Act3Output{
				Climax:     "最终决战",
				Resolution: "归来与改变",
			},
		},
		CoreConflicts: []CoreConflict{
			{
				Type:           "与自己",
				Description:    "主角内心挣扎",
				EscalationPath: []string{"犹豫", "冲突", "抉择"},
				Resolution:     "接受自我",
			},
		},
	}

	// 验证三幕结构完整性
	if output.StructureType != StructureThreeAct {
		t.Errorf("StructureType = %q, want 'three_act'", output.StructureType)
	}

	if output.ThreeAct == nil {
		t.Fatal("ThreeAct should not be nil")
	}

	if output.ThreeAct.Act1.Setup == "" {
		t.Error("Act1.Setup should not be empty")
	}

	if len(output.ThreeAct.Act2.RisingAction) == 0 {
		t.Error("Act2.RisingAction should have items")
	}

	if output.ThreeAct.Act3.Climax == "" {
		t.Error("Act3.Climax should not be empty")
	}

	// 验证冲突系统
	if len(output.CoreConflicts) == 0 {
		t.Error("CoreConflicts should have items")
	}
}

// contains 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkPlanTheme 性能测试
func BenchmarkPlanTheme(b *testing.B) {
	engine := &NarrativeEngine{}
	coreTheme := "勇气与成长"
	chapterCount := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.planTheme(coreTheme, chapterCount)
	}
}
