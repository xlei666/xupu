# 叙事器与世界设定器的联动

## 数据流向

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              数据库 (Database)                              │
│  ┌─────────────────┐      ┌──────────────────────────────────────────┐    │
│  │  WorldSetting   │◄─────│         NarrativeBlueprint               │    │
│  │  (世界设定)      │      │         (叙事蓝图)                       │    │
│  │                 │      │                                          │    │
│  │ - Philosophy    │      │ - WorldID ──────────────────────────────┤    │
│  │ - Worldview     │      │ - StoryOutline  (基于世界设定生成)       │    │
│  │ - Laws          │      │ - ChapterPlans  (基于世界设定生成)       │    │
│  │ - Geography     │      │ - Scenes        (地点来自世界设定)       │    │
│  │ - Civilization  │      │ - CharacterArcs (种族来自文明设定)       │    │
│  │ - Society       │      │ - ThemePlan     (主题来自哲学核心问题)   │    │
│  │ - StorySoil     │      │                                          │    │
│  └─────────────────┘      └──────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
           ▲                                                         │
           │                                                         │
           │ GetWorld(worldID)                                      │ SaveBlueprint()
           │                                                         │
           │                                                         ▼
┌─────────────────────────┐                              ┌─────────────────────┐
│   WorldBuilder          │                              │  NarrativeEngine    │
│   (世界设定器)           │                              │  (叙事器)            │
│                          │  CreateBlueprint()        │                      │
│  1. Philosophy (哲学)   │◄────────────────────────────│ WorldID             │
│  2. Worldview (世界观)  │                              │                     │
│  3. Laws (法则)         │                              │ 1. buildWorldSummary()│
│  4. StorySoil (故事土壤)│                              │    ↓                 │
│  5. Geography (地理)    │                              │  提取世界信息        │
│  6. Civilization (文明) │                              │    ↓                 │
│  7. Society (社会)      │                              │  生成故事大纲        │
│                          │                              │    ↓                 │
│                          │                              │  生成章节规划        │
│                          │                              │    ↓                 │
│                          │                              │  生成场景指令        │
└─────────────────────────┘                              └─────────────────────┘
```

## 联动方式

### 1. 通过 WorldID 关联

```go
// CreateParams - 叙事器创建参数
type CreateParams struct {
    WorldID    string  // ← 指向已创建的世界设定
    StoryType  string  // 故事类型
    Theme      string  // 核心主题
    Protagonist string // 主角概念
    Length     string  // 篇幅
    ChapterCount int   // 章节数
    Structure  NarrativeStructure // 叙事结构
}
```

### 2. 世界设定摘要提取

叙事器从世界设定中提取关键信息，传递给LLM生成叙事蓝图：

```go
func (ne *NarrativeEngine) buildWorldSummary(world *models.WorldSetting) string {
    summary := fmt.Sprintf("世界名称: %s\n类型: %s\n规模: %s\n\n",
        world.Name, world.Type, world.Scale)

    // 提取哲学信息
    summary += fmt.Sprintf("【哲学】核心问题: %s\n", world.Philosophy.CoreQuestion)
    summary += fmt.Sprintf("【价值观】最高善: %s, 最大恶: %s\n",
        world.Philosophy.ValueSystem.HighestGood,
        world.Philosophy.ValueSystem.UltimateEvil)

    // 提取故事土壤（社会冲突、情节钩子）
    if len(world.StorySoil.SocialConflicts) > 0 {
        summary += fmt.Sprintf("【社会冲突】%d个主要矛盾\n", len(world.StorySoil.SocialConflicts))
        for i, conflict := range world.StorySoil.SocialConflicts {
            summary += fmt.Sprintf("  - %s: %s\n", conflict.Type, conflict.Description)
        }
    }

    // 提取地理环境（用于场景地点）
    if len(world.Geography.Regions) > 0 {
        summary += fmt.Sprintf("【地理】%d个区域\n", len(world.Geography.Regions))
    }

    // 提取文明种族（用于角色创建）
    if len(world.Civilization.Races) > 0 {
        summary += "【种族】"
        for i, race := range world.Civilization.Races {
            summary += race.Name + " "
        }
    }

    return summary
}
```

### 3. 世界设定 → 叙事蓝图的映射

| 世界设定 (输入) | 叙事蓝图 (输出) | 映射方式 |
|----------------|----------------|----------|
| **Philosophy.CoreQuestion** | StoryOutline | 核心问题成为故事的主题线索 |
| **Philosophy.ValueSystem** | ThemePlan | 价值观决定主题的道德框架 |
| **StorySoil.SocialConflicts** | StoryOutline.Act2 | 社会冲突转化为上升动作 |
| **StorySoil.PotentialPlotHooks** | ChapterPlans.EndingHook | 情节钩子成为章节结尾悬念 |
| **Geography.Regions** | SceneInstruction.Location | 地理区域成为场景地点 |
| **Geography.Climate** | SceneInstruction.Mood | 气候影响场景氛围 |
| **Civilization.Races** | CharacterArcs | 种族特征影响角色弧光 |
| **Society.PoliticalSystem** | StoryOutline | 政治制度成为冲突背景 |

### 4. 实际调用流程

```go
// 1. 用户先创建世界设定
wb, _ := worldbuilder.New()
world, _ := wb.Build(worldbuilder.BuildParams{
    Name:  "艾尔德里亚大陆",
    Type:  models.WorldFantasy,
    Theme: "力量与腐败",
})
// world.ID = "world_1234567890"

// 2. 然后使用叙事器创建蓝图（传入worldID）
engine, _ := narrative.New()
blueprint, _ := engine.CreateBlueprint(narrative.CreateParams{
    WorldID:    world.ID,  // ← 关联世界设定
    StoryType:  "奇幻冒险",
    Theme:      "力量与腐败",
    Protagonist: "一个年轻的魔法师",
    Length:     "short",
})

// 3. 叙事器内部流程
//    a. 从数据库获取世界设定: db.GetWorld(worldID)
//    b. 构建世界摘要: buildWorldSummary(world)
//    c. 调用LLM生成大纲: GenerateOutline(包含WorldSummary)
//    d. 生成章节规划: GenerateChapterPlans()
//    e. 生成场景指令: GenerateScenes() (地点来自world.Geography.Regions)
//    f. 保存蓝图: db.SaveNarrativeBlueprint(blueprint)
```

## 数据库关联

```
data/
├── worlds.json        # WorldSetting[] → 包含完整的世界设定
└── blueprints.json    # NarrativeBlueprint[] → 每个蓝图有 WorldID 字段
                                              → 通过 WorldID 反查 worlds.json
```

## 演化引擎的联动

```go
// evolution.go 中的演化引擎也使用世界设定
func (ee *EvolutionEngine) CreateEvolutionState(worldID string) (*EvolutionState, error) {
    // 获取世界设定
    world, err := ee.db.GetWorld(worldID)

    // 将世界设定嵌入演化状态
    state := &EvolutionState{
        WorldContext:    world,           // ← 完整的世界设定作为上下文
        ThemeEvolution: &ThemeEvolutionState{
            CoreTheme: world.Philosophy.CoreQuestion, // ← 主题来自世界哲学
        },
        Characters:      make(map[string]*CharacterState),
        Conflicts:       []*ConflictThread{},
    }

    // 从世界设定中提取种族创建角色
    for _, race := range world.Civilization.Races {
        char := ee.createCharacterState(race, state)
        state.Characters[char.ID] = char
    }

    // 从世界设定中提取社会冲突
    for _, soilConflict := range world.StorySoil.SocialConflicts {
        conflict := &ConflictThread{
            Type:         soilConflict.Type,
            CoreQuestion: soilConflict.Description,
            CurrentIntensity: soilConflict.Tension,
        }
        state.Conflicts = append(state.Conflicts, conflict)
    }

    return state, nil
}
```

## 总结

1. **物理关联**: 通过 `WorldID` 字段在数据库层面关联
2. **信息提取**: 叙事器通过 `buildWorldSummary()` 提取世界设定的关键信息
3. **LLM传递**: 提取的信息作为 Prompt 传递给 LLM 生成叙事蓝图
4. **双向使用**:
   - 世界设定 → 叙事蓝图: 一对多（一个世界可以生成多个故事）
   - 演化状态直接引用世界设定作为上下文
