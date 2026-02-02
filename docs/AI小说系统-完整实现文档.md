# NovelFlow 叙谱 - AI小说创作系统完整实现文档

## 目录

1. [系统概述](#1-系统概述)
2. [调度器](#2-调度器)
3. [编排器](#3-编排器)
4. [世界设定器](#4-世界设定器)
5. [叙事器](#5-叙事器)
6. [写作器](#6-写作器)
7. [知识数据服务](#7-知识数据服务)
8. [数据库设计](#8-数据库设计)
9. [LLM调用规范](#9-llm调用规范)
10. [实现路径](#10-实现路径)

---

## 1. 系统概述

### 1.1 系统定位

一个分层协作的AI小说创作系统，通过六个模块协同工作，完成从创意到成稿的全流程。

### 1.2 设计原则

| 原则 | 说明 |
|------|------|
| 分层解耦 | 每层专注单一职责，通过清晰接口交互 |
| 约束驱动 | 所有生成在设定约束下进行，确保一致性 |
| 理论驱动 | 叙事器整合成熟叙事理论，非随机生成 |
| 状态持久 | 所有状态实时保存，支持中断恢复 |
| 可扩展 | 模块可独立升级，知识库可自我扩充 |

### 1.3 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                         调度器                             │
│                      (Scheduler)                           │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────┐
│                        编排器                              │
│                      (Orchestrator)                         │
└─────┬─────────┬─────────┬─────────┬─────────┬──────────────┘
      │         │         │         │         │
      ↓         ↓         ↓         ↓         ↓
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
│世界设定 │ │ 叙事器  │ │ 写作器  │ │知识服务 │ │角色DB  │
└─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘
```

---

## 2. 调度器

### 2.1 功能定义

系统入口，负责创建和管理编排器实例。

### 2.2 核心职责

| 职责 | 说明 |
|------|------|
| 创建编排器 | 每部小说创建一个独立编排器实例 |
| 并发管理 | 同时管理多个编排器，分配资源 |
| 状态监控 | 监控各编排器执行状态 |
| 结果汇总 | 收集并返回最终成果 |

### 2.3 接口定义

**输入接口：**
```json
{
  "action": "create",
  "mode": "planning|intervention|random|story_core|short|script",
  "parameters": {
    // 根据模式不同，参数不同
  }
}
```

**输出接口：**
```json
{
  "status": "success",
  "orchestrator_id": "orc_001",
  "message": "编排器已创建"
}
```

### 2.4 并发管理规则

1. 每个编排器独立运行，互不干扰
2. 共享资源（知识库、角色库）使用读写锁
3. 超时编排器自动挂起，支持手动恢复

---

## 3. 编排器

### 3.1 功能定义

单部小说的编排中心，协调三大核心层完成创作。

### 3.2 编排模式

#### 3.2.1 规划生成模式

```
用户参数搜集 → 世界设定器 → 叙事器 → 写作器 → 成稿
```

| 步骤 | 模块 | 输出 |
|------|------|------|
| 1 | 搜集参数 | 创作参数 |
| 2 | 世界设定器 | 世界设定包 |
| 3 | 叙事器 | 叙事蓝图 |
| 4 | 写作器 | 小说文本 |

#### 3.2.2 干预生成模式

```
生成 → 用户干预 → 调整 → 继续生成 → ...
```

允许用户在任意节点介入：
- 修改世界设定
- 调整情节规划
- 修改角色设定
- 重写某段文本

#### 3.2.3 随机生成模式

```
随机参数 → 快速生成 → 展示选项 → 用户选择
```

#### 3.2.4 故事核模式

```
一句话创意 → 扩写 → 完整故事
```

#### 3.2.5 短篇模式

一次性生成短篇小说，不保存中间状态。

#### 3.2.6 剧本模式

按剧本格式输出，包含场景标题、角色名、对话、动作描述。

### 3.3 编排流程

```python
def orchestrate(mode, parameters):
    # 步骤1: 初始化
    context = initialize_context(parameters)

    # 步骤2: 世界构建
    world_package = world_builder.build(context)

    # 步骤3: 叙事规划
    narrative_blueprint = narrative_engine.plan(
        world_package,
        context.creative_intent
    )

    # 步骤4: 内容生成
    text = writer.write(
        narrative_blueprint,
        world_package,
        knowledge_service
    )

    # 步骤5: 汇总
    return compile_result(text, world_package, narrative_blueprint)
```

---

## 4. 世界设定器

### 4.1 功能定义

构建有机、自洽、能产生故事的世界系统。

### 4.2 设计原则

1. **因果链原则**：一切皆有因，历史事件能追溯到社会矛盾
2. **代价机制原则**：所有力量都有代价，限制产生张力
3. **深度真实原则**：社会有多层结构，非简单二元对立
4. **依赖链原则**：每层生成依赖上一层，用`derivation`记录

### 4.3 分层结构

```
哲学思考（第1层）
    ↓ 依赖
世界观（第2层）
    ↓ 依赖
设定（第3层）
    ↓ 依赖
故事土壤（第4层）
```

### 4.4 生成器列表

| 生成器 | 功能 | 输出 |
|--------|------|------|
| 法则生成器 | 物理/魔法法则 | 法则手册 |
| 地理生成器 | 地形/气候/资源 | 区域档案 |
| 文明生成器 | 种族/语言/宗教 | 文明档案 |
| 社会生成器 | 政治/阶级/经济 | 社会结构图 |
| 历史生成器 | 时代/事件/遗产 | 历史年表 |
| 一致性检查器 | 逻辑验证 | 一致性报告 |

### 4.5 生成流程

#### 阶段1：哲学思考

**输入：**
```json
{
  "world_type": "fantasy|scifi|historical|urban",
  "core_theme": "权力/信仰/生存/爱情/复仇",
  "style_tendency": "realistic|romantic|dark|epic"
}
```

**Prompt模板：**
```
# 角色
你是一位世界哲学家，负责为虚构世界构建底层思想体系。

# 输入参数
- 世界类型：{world_type}
- 核心主题：{core_theme}
- 风格倾向：{style_tendency}

# 任务
生成这个世界的【哲学基础】：

1. **核心问题**
   - 这个世界要探讨什么根本问题？
   - 例如：人性本善还是本恶？命运能否改变？

2. **价值体系**
   - 这个世界认为什么是"善"？
   - 这个世界认为什么是"恶"？
   - 这些价值观如何形成？

3. **主题方向**
   - 哪些主题会被深入探讨？
   - 这些主题如何冲突？

# 输出格式（JSON）
{
  "philosophy": {
    "core_question": "要探讨的根本问题",
    "derivation": "说明为什么提出这个问题",
    "value_system": {
      "highest_good": "最高善",
      "ultimate_evil": "最大恶",
      "moral_dilemmas": [
        {"dilemma": "道德困境", "description": "描述"}
      ]
    },
    "themes": [
      {"theme": "主题名称", "exploration_angle": "探讨角度"}
    ]
  }
}
```

**输出：**
```json
{
  "philosophy": {
    "core_question": "权力是否必然导致腐败？",
    "derivation": "基于{core_theme}，探索权力的本质",
    "value_system": {
      "highest_good": "守护与责任",
      "ultimate_evil": "自私与背叛"
    },
    "themes": [...]
  }
}
```

#### 阶段2：世界观

**输入：** 上一阶段的`philosophy`输出

**Prompt模板：**
```
# 角色
你是一位世界构建师，负责从哲学基础推导世界观。

# 输入
## 哲学基础
{philosophy JSON}

# 任务
基于【哲学基础】推导【世界观】：

1. **宇宙论** - 基于{core_question}，世界如何起源？
2. **形而上学** - 基于{highest_good}，灵魂是否存在？
3. **存在意义** - 生命的意义是什么？

# 输出格式（JSON）
{
  "worldview": {
    "derivation_logic": "说明如何从哲学推导",
    "cosmology": {"origin": "...", "structure": "..."},
    "metaphysics": {
      "soul": {"exists": true, "nature": "..."},
      "afterlife": "...",
      "fate": {"exists": true, "relationship": "..."}
    }
  }
}
```

#### 阶段3-6：以此类推

每个阶段都接收上一阶段输出，在Prompt中包含：
1. 上一阶段的完整输出
2. `derivation`字段说明依赖关系
3. 明确的约束条件

### 4.6 输出格式

```json
{
  "world_package": {
    "id": "world_001",
    "name": "世界名称",
    "metadata": {
      "type": "奇幻",
      "scale": "大陆级",
      "generated_at": "2025-01-16T10:00:00Z"
    },

    "philosophy": {},      // 第1层：哲学思考
    "worldview": {},       // 第2层：世界观
    "settings": {         // 第3层：设定
      "physics": {},
      "geography": {},
      "civilizations": [],
      "society": {}
    },

    // 第4层：故事土壤（叙事器最需要）
    "story_soil": {
      "social_conflicts": [
        {
          "type": "经济矛盾",
          "description": "土地兼并严重",
          "parties": ["地主", "农民"],
          "tension_level": 8,
          "potential_triggers": ["天灾", "重税"],
          "derivation": "基于{society.class_structure}产生"
        }
      ],

      "power_structures": [
        {
          "name": "朝廷",
          "formal": {"type": "君主制", "ruler": "皇帝"},
          "actual": {
            "surface_power": "皇帝",
            "real_power_holders": [
              {"entity": "宦官", "power_source": "接近皇帝"},
              {"entity": "世家", "power_source": "经济垄断"}
            ]
          }
        }
      ],

      "historical_context": {
        "recent_events": [
          {"event": "叛乱", "impact": "信任度下降"}
        ],
        "collective_traumas": ["外族恐惧"],
        "derivation": "基于{history}形成"
      },

      "potential_plot_hooks": [
        {
          "type": "权力真空",
          "description": "皇帝病重",
          "story_potential": "各方蠢蠢欲动"
        }
      ]
    },

    // 设定约束（写作器需要）
    "setting_constraints": {
      "magic_system": {"exists": true, "rules": "..."},
      "technology_level": "封建时代",
      "geography_summary": {}
    },

    // 角色基础设定
    "character_templates": [
      {
        "template_id": "tmpl_noble",
        "background": "贵族家庭",
        "typical_traits": ["傲慢", "责任感"],
        "typical_abilities": ["骑术", "礼仪"]
      }
    ]
  }
}
```

### 4.7 一致性检查

```python
def check_consistency(world_package):
    """检查世界设定的一致性"""

    checks = [
        check_physics_consistency,
        check_geography_economy_consistency,
        check_technology_society_consistency,
        check_history_causality_consistency,
        check_population_resource_consistency
    ]

    issues = []
    for check in checks:
        issues.extend(check(world_package))

    return {
        "score": calculate_score(issues),
        "issues": issues,
        "suggestions": generate_suggestions(issues)
    }
```

---

## 5. 叙事器

### 5.1 功能定义

规划"写什么"——情节、冲突、弧光、节奏。

### 5.2 设计原则

1. **理论驱动**：整合成熟叙事理论
2. **冲突至上**：没有冲突就没有故事
3. **因果严谨**：情节发展必须有因果链
4. **角色驱动**：情节源于角色欲望与行动

### 5.3 整合的叙事理论

| 理论 | 适用场景 |
|------|----------|
| 三幕结构 | 通用类型 |
| 英雄之旅 | 成长故事 |
| 救猫咪节拍表 | 商业类型片 |
| 起承转合 | 东方叙事 |
| 弗赖塔格金字塔 | 传统正剧 |

### 5.4 生成流程

#### 轮次1：整体大纲

**输入：**
```json
{
  "world_package": {...},
  "creative_intent": {
    "story_type": "冒险/爱情/悬疑/成长",
    "protagonist_concept": "...",
    "length_target": "短篇/中篇/长篇",
    "special_requirements": []
  }
}
```

**Prompt模板：**
```
# 角色
你是一位专业的故事策划师，精通叙事理论。

# 输入
## 世界设定摘要
{从world_package提取故事土壤}

## 创作意图
- 故事类型：{story_type}
- 核心主题：{从world_package.philosophy提取}
- 篇幅预期：{length_target}

## 主角概念
{protagonist_concept}

# 任务
使用【三幕结构】生成故事大纲：

1. **第一幕（Setup）**
   - 建立世界：展示什么？
   - 主角现状：起始状态是什么？
   - 激励事件：什么打破平衡？
   - 第一幕情节点：主角如何踏上旅程？

2. **第二幕**
   - 试炼与盟友：遇到什么障碍？
   - 中点：重大转折是什么？
   - 一无所有：最低点是什么？
   - 第二幕情节点：如何进入最终对抗？

3. **第三幕**
   - 高潮：最终对抗如何展开？
   - 结局：主角变化了什么？

# 输出格式（JSON）
{
  "story_outline": {
    "act1": {
      "setup": "世界建立",
      "inciting_incident": "激励事件",
      "plot_point1": "第一幕情节点"
    },
    "act2": {
      "rising_action": ["试炼1", "试炼2"],
      "midpoint": "中点",
      "all_is_lost": "一无所有",
      "plot_point2": "第二幕情节点"
    },
    "act3": {
      "climax": "高潮",
      "resolution": "结局"
    }
  },
  "character_arc": {
    "start_state": "起始状态",
    "end_state": "目标状态",
    "turning_points": ["转折点1", "转折点2"]
  },
  "core_conflicts": [
    {
      "type": "人与人/与社会/与自己",
      "description": "冲突描述",
      "escalation_path": ["阶段1", "阶段2", "阶段3"]
    }
  ]
}
```

#### 轮次2：章节规划

**输入：** 上一轮大纲 + 篇幅参数

**Prompt模板：**
```
# 角色
你是一位章节规划师，负责将大纲细化为章节。

# 输入
## 故事大纲
{story_outline}

## 篇幅要求
- 总字数目标：{target_words}
- 章节数估计：{estimated_chapters}

# 任务
将大纲细化为{estimated_chapters}章：

1. 每章确定：
   - 章节目的
   - 关键场景
   - 推进的情节
   - 变化的角色弧光

2. 确保章节间：
   - 因果连贯
   - 节奏递进
   - 悬念连接

# 输出格式（JSON）
{
  "chapter_plans": [
    {
      "chapter": 1,
      "title": "章节标题",
      "purpose": "本章目的",
      "key_scenes": ["场景1", "场景2"],
      "plot_advancement": "推进了什么",
      "arc_progress": "弧光进展",
      "ending_hook": "结尾悬念"
    }
  ]
}
```

#### 轮次3：场景序列

**输入：** 章节规划 + 世界设定

**Prompt模板：**
```
# 角色
你是一位场景规划师，负责生成详细的场景序列。

# 输入
## 第{N}章规划
{chapter_plan}

## 当前状态
- 已发生情节：{previous_summary}
- 角色状态：{character_states}

# 任务
为第{N}章生成场景序列：

每个场景包含：
- 场景目的
- 在场角色
- 发生地点
- 关键行动
- 对话重点
- 结果/变化

# 输出格式（JSON）
{
  "scenes": [
    {
      "sequence": 1,
      "purpose": "场景目的",
      "location": "地点",
      "characters": ["角色A", "角色B"],
      "action": "发生什么",
      "dialogue_focus": "对话重点",
      "outcome": "结果",
      "state_changes": {
        "character_updates": [...],
        "plot_advancement": "..."
      }
    }
  ]
}
```

#### 轮次4：动态演化

在写作过程中，根据已生成内容调整后续规划。

**Prompt模板：**
```
# 角色
你是一位故事调整师，负责根据前文调整后续规划。

# 输入
## 原始规划
{original_plan}

## 已生成内容摘要
{generated_summary}

## 出现的新情况
{new_developments}

# 任务
分析已生成内容，调整后续规划：

1. 评估：
   - 原规划是否需要调整？
   - 新引入了哪些伏笔需要回收？
   - 角色状态是否偏离原规划？

2. 调整：
   - 更新后续章节规划
   - 确保新伏笔有回收计划
   - 必要时调整角色弧光

# 输出格式（JSON）
{
  "adjustments": {
    "reason": "调整原因",
    "updated_chapters": [...],
    "new_plot_threads": [...],
    "foreshadowing_to_plant": [...]
  }
}
```

### 5.5 角色弧光规划

**输入：** 世界设定器的角色基础设定 + 主角概念

**输出：**
```json
{
  "character_arc": {
    "character_id": "char_001",
    "arc_type": "成长弧光/堕落弧光/扁平弧光",
    "start_state": {
      "personality": ["傲慢", "自大"],
      "motivation": "外在目标",
      "flaw": "致命缺陷"
    },
    "end_state": {
      "personality": ["谦逊", "责任"],
      "motivation": "内在需求",
      "growth": "学会了什么"
    },
    "turning_points": [
      {
        "chapter": 3,
        "event": "关键事件",
        "change": "发生了什么变化"
      }
    ],
    "derivation": "基于{protagonist_concept}和{world_theme}设计"
  }
}
```

### 5.6 输出格式

```json
{
  "narrative_blueprint": {
    "id": "narrative_001",
    "world_id": "world_001",

    "story_outline": {},     // 整体大纲
    "chapter_plans": [],     // 章节规划
    "scene_instructions": [], // 场景指令

    "character_arcs": {},    // 角色弧光

    "theme_plan": {         // 主题规划
      "core_theme": "...",
      "threading": [
        {"chapter": 1, "expression": "主题体现"},
        {"chapter": 5, "expression": "主题深化"},
        {"chapter": 10, "expression": "主题升华"}
      ]
    },

    "metadata": {
      "structure_used": "三幕结构",
      "estimated_length": "100000字",
      "generated_at": "2025-01-16T10:00:00Z"
    }
  }
}
```

---

## 6. 写作器

### 6.1 功能定义

生成"怎么写"——具体文本内容。

### 6.2 设计原则

1. **约束驱动**：在世界设定约束下生成
2. **一致性优先**：角色行为、语言风格前后一致
3. **展示而非讲述**：通过行动和对话展现
4. **感官沉浸**：调动读者感官体验

### 6.3 处理流程

```
输入：场景指令 + 世界状态 + 角色状态
    ↓
1. 解析指令（场景目的、在场角色、视角）
    ↓
2. 查询状态（角色、关系、环境）
    ↓
3. 生成内容（对话、动作、描写）
    ↓
4. 应用视角过滤
    ↓
5. 应用风格控制
    ↓
输出：场景文本 + 状态更新
```

### 6.4 角色系统

#### 角色数据结构

```json
{
  "character_id": "char_001",

  "static_profile": {      // 世界设定器生成
    "name": "...",
    "background": "...",
    "race": "...",
    "abilities": {...}
  },

  "narrative_profile": {   // 叙事器生成
    "personality": [...],
    "motivation": {...},
    "flaw": "...",
    "arc_plan": {...}
  },

  "dynamic_state": {       // 写作器维护
    "location": "...",
    "emotion": {
      "current": "平静",
      "intensity": 50,
      "trigger": "无"
    },
    "knowledge": {
      "known": [...],
      "unknown": [...],
      "mistaken": [...]
    },
    "relationships": {
      "char_002": {
        "emotion": 80,      // -100到100
        "power": "equal",
        "secrets": ["秘密1"]
      }
    },
    "health": {...},
    "arc_progress": 30      // 0-100
  }
}
```

#### 一致性检查

```python
def check_character_consistency(character, action, dialogue):
    """检查角色行为一致性"""

    checks = {
        "personality": is_consistent_with_personality(character, action),
        "emotion": is_consistent_with_emotion(character, action),
        "knowledge": is_consistent_with_knowledge(character, dialogue),
        "ability": is_within_ability(character, action),
        "relationship": is_consistent_with_relationship(character, action)
    }

    return all(checks.values()), checks
```

### 6.5 对话生成

**输入：**
```json
{
  "speaker": "char_001",
  "listener": "char_002",
  "context": {...},
  "dialogue_purpose": "请求/说服/威胁/安慰",
  "current_emotion": "焦急"
}
```

**Prompt模板：**
```
# 角色
你是一位对话生成专家。

# 输入
## 发言者
{speaker的narrative_profile和dynamic_state}

## 听话者
{listener的narrative_profile和dynamic_state}

## 关系状态
{relationship_data}

## 对话目的
{dialogue_purpose}

## 当前情境
{context}

# 任务
生成符合角色的对话：

1. 考虑：
   - 发言者的性格和语言风格
   - 发言者的当前情绪
   - 发言者对听话者的态度
   - 发言者想达到什么目的

2. 包含：
   - 表面话语
   - 潜台词（可选）
   - 非语言信号（动作/表情）

# 输出格式（JSON）
{
  "dialogue": "对话内容",
  "subtext": "潜台词",
  "non_verbal": "动作/表情描述",
  "tone": "语气"
}
```

### 6.6 场景渲染

**Prompt模板：**
```
# 角色
你是一位场景描写专家。

# 输入
## 场景信息
- 地点：{location}
- 时间：{time}
- 天气：{weather}
- 在场角色：{characters}

## 氛围要求
{mood_requirements}

# 任务
生成场景描写：

1. 调动感官：
   - 视觉：光线、色彩、空间
   - 听觉：环境声、人声
   - 嗅觉/味觉：气味
   - 触觉：温度、质感

2. 营造氛围：
   - 符合场景目的
   - 与角色情绪共鸣/反差

3. 长度控制：
   - 过渡场景：简洁
   - 氛围场景：详细
   - 动作场景：精简

# 输出格式（JSON）
{
  "scene_text": "场景描写文本",
  "sensory_elements": {
    "visual": [...],
    "auditory": [...],
    "olfactory": [...]
  }
}
```

### 6.7 视角处理

```python
def apply_pov_filter(content, pov_character, knowledge_state):
    """应用视角过滤"""

    # 1. 信息过滤
    filtered_content = filter_by_knowledge(content, knowledge_state)

    # 2. 注入主观性
    subjective_content = add_subjectivity(filtered_content, pov_character)

    # 3. 处理不可靠叙事
    final_content = handle_unreliable_narration(
        subjective_content,
        pov_character
    )

    return final_content
```

### 6.8 输出格式

```json
{
  "scene_output": {
    "chapter": 1,
    "scene": 1,
    "text": "场景文本内容...",

    "state_updates": {
      "characters": [
        {
          "id": "char_001",
          "location": "新位置",
          "emotion": {"current": "愤怒", "intensity": 80},
          "knowledge_gain": ["新信息"],
          "relationships": {
            "char_002": {"emotion_change": -20}
          }
        }
      ],
      "world": {
        "time": "第二天早晨",
        "events": ["发生了什么"]
      }
    },

    "metadata": {
      "word_count": 1500,
      "pov_character": "char_001",
      "generated_at": "2025-01-16T10:00:00Z"
    }
  }
}
```

---

## 7. 知识数据服务

### 7.1 功能定义

为写作器提供写作规则和知识素材。

### 7.2 知识库结构

```
knowledge_base/
├── writing_rules/
│   ├── prose_styles/      # 散文风格
│   ├── dialogue_tips/     # 对话技巧
│   ├── scene_techniques/  # 场景技法
│   └── genre_conventions/ # 类型惯例
├── domain_knowledge/
│   ├── history/           # 历史知识
│   ├── science/           # 科学知识
│   ├── culture/           # 文化知识
│   └── professions/       # 职业知识
└── templates/
    ├── character_templates/
    └── plot_templates/
```

### 7.3 自我扩充机制

```python
def expand_knowledge(query):
    """查询并扩充知识"""

    # 1. 查询现有知识库
    result = search_knowledge_base(query)

    if result:
        return result

    # 2. 知识库中不存在，触发扩充
    if should_expand(query):
        new_knowledge = fetch_external_knowledge(query)
        if validate_knowledge(new_knowledge):
            add_to_knowledge_base(new_knowledge)
            return new_knowledge

    return None
```

**扩充来源：**
1. 维基百科
2. 专业数据库
3. 用户上传
4. LLM生成并验证

### 7.4 API接口

```python
# 查询规则
GET /api/rules/{category}

# 查询知识
GET /api/knowledge/{domain}/{topic}

# 添加知识
POST /api/knowledge

# 触发扩充
POST /api/knowledge/expand
```

---

## 8. 数据库设计

### 8.1 世界数据库

```sql
CREATE TABLE worlds (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(200),
    type ENUM('fantasy', 'scifi', 'historical', 'urban'),
    philosophy JSON,
    worldview JSON,
    settings JSON,
    story_soil JSON,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### 8.2 角色数据库

```sql
CREATE TABLE characters (
    id VARCHAR(50) PRIMARY KEY,
    world_id VARCHAR(50),
    name VARCHAR(100),

    -- 静态档案
    static_profile JSON,

    -- 叙事档案
    narrative_profile JSON,

    -- 动态状态
    dynamic_state JSON,

    FOREIGN KEY (world_id) REFERENCES worlds(id),
    updated_at TIMESTAMP
);
```

### 8.3 叙事蓝图数据库

```sql
CREATE TABLE narratives (
    id VARCHAR(50) PRIMARY KEY,
    world_id VARCHAR(50),
    blueprint JSON,
    FOREIGN KEY (world_id) REFERENCES worlds(id)
);
```

### 8.4 生成文本数据库

```sql
CREATE TABLE generated_text (
    id VARCHAR(50) PRIMARY KEY,
    narrative_id VARCHAR(50),
    chapter INT,
    scene INT,
    content TEXT,
    metadata JSON,
    FOREIGN KEY (narrative_id) REFERENCES narratives(id)
);
```

---

## 9. LLM调用规范

### 9.1 通用Prompt结构

```
# 角色
定义AI扮演的角色

# 输入
提供所需的上下文和数据

# 任务
明确要求完成的任务

# 约束
必须遵守的规则

# 输出格式
指定输出格式（优先JSON）
```

### 9.2 调用原则

| 原则 | 说明 |
|------|------|
| 分阶段 | 复杂任务分多阶段，每阶段接收前阶段输出 |
| 结构化 | 输出统一为JSON格式 |
| 依赖记录 | 用`derivation`字段记录依赖关系 |
| 约束明确 | 在Prompt中明确列出约束条件 |
| 示例驱动 | 必要时提供输入/输出示例 |

### 9.3 错误处理

```python
def call_llm(prompt, max_retries=3):
    """调用LLM并处理错误"""

    for attempt in range(max_retries):
        try:
            response = llm_client.generate(prompt)

            # 验证输出格式
            parsed = validate_and_parse(response)

            # 验证内容
            if validate_content(parsed):
                return parsed

        except ParseError as e:
            if attempt < max_retries - 1:
                # 在Prompt中添加错误反馈
                prompt = add_error_feedback(prompt, str(e))
                continue
            raise

    raise LLMCallError("Max retries exceeded")
```

### 9.4 Token估算

```
| 任务类型 | 输入Token | 输出Token |
|----------|-----------|-----------|
| 世界设定（单阶段）| 2000-3000 | 1000-2000 |
| 叙事大纲 | 3000-5000 | 1500-2500 |
| 章节规划 | 2000-4000 | 1000-1500 |
| 场景生成 | 1500-3000 | 500-1500 |
| 对话生成 | 1000-2000 | 200-500 |
```

---

## 10. 实现路径

### 10.1 MVP阶段

**功能范围：**
- 简单世界设定（简化流程）
- 三幕结构叙事
- 基础对话和描写生成
- 短篇小说生成

**技术栈：**
- Python 3.10+
- OpenAI API / Claude API
- SQLite / PostgreSQL
- FastAPI

### 10.2 核心模块阶段

**新增功能：**
- 完整世界设定流程
- 多种叙事理论支持
- 角色弧光系统
- 状态持久化

### 10.3 完整系统阶段

**新增功能：**
- 长篇小说支持
- 并发多小说生成
- 知识库自我扩充
- 用户干预界面

---

## 附录A：角色职责总结

| 模块 | 负责内容 | 输入 | 输出 |
|------|----------|------|------|
| 调度器 | 创建/管理编排器 | 用户请求 | 编排器实例 |
| 编排器 | 协调创作流程 | 模式+参数 | 小说文本 |
| 世界设定器 | 构建世界设定 | 创作参数 | 世界设定包 |
| 叙事器 | 规划情节结构 | 世界设定+意图 | 叙事蓝图 |
| 写作器 | 生成具体文本 | 蓝图+设定+知识 | 场景文本 |
| 知识服务 | 提供规则/知识 | 查询请求 | 知识内容 |

---

## 附录B：角色设定分工

| 内容 | 负责模块 | 存储位置 |
|------|----------|----------|
| 基础背景、种族、能力 | 世界设定器 | static_profile |
| 性格、动机、弧光 | 叙事器 | narrative_profile |
| 动态状态、关系 | 写作器 | dynamic_state |

---

## 附录C：数据流转

```
用户请求
    ↓
调度器 → 创建编排器
    ↓
编排器 → 世界设定器
    ↓ 世界设定包
编排器 → 叙事器
    ↓ 叙事蓝图
编排器 → 写作器 + 知识服务
    ↓ 场景文本
编排器 → 汇总成稿
    ↓
返回用户
```

---

*文档版本：1.0*
*最后更新：2025-01-16*
