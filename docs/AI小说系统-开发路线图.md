# NovelFlow 叙谱 - AI小说创作系统开发路线图

## 推荐开发顺序

```
阶段0: 项目搭建
    ↓
阶段1: LLM客户端 ← 核心基础设施
    ↓
阶段2: 数据模型
    ↓
阶段3: 世界设定器
    ↓
阶段4: 叙事器
    ↓
阶段5: 写作器
    ↓
阶段6: 编排器
    ↓
阶段7: 调度器与API
```

---

## 阶段0：项目搭建 (Day 1)

### 0.1 创建项目结构

```bash
xupu/
├── config/
│   ├── __init__.py
│   ├── llm.yaml           # LLM配置
│   └── settings.py         # 全局设置
├── llm/
│   ├── __init__.py
│   ├── client.py           # LLM客户端
│   └── retry_client.py     # 重试包装
├── db/
│   ├── __init__.py
│   ├── models.py           # 数据模型
│   └── database.py         # 数据库操作
├── modules/
│   ├── __init__.py
│   ├── world_builder/      # 世界设定器
│   ├── narrative_engine/   # 叙事器
│   └── writer/             # 写作器
├── orchestrator/
│   ├── __init__.py
│   └── orchestrator.py     # 编排器
├── scheduler/
│   ├── __init__.py
│   └── scheduler.py        # 调度器
├── api/
│   └── main.py             # FastAPI入口
├── tests/
│   └── test_client.py
├── .env
├── requirements.txt
└── main.py
```

### 0.2 安装依赖

```bash
# requirements.txt
fastapi==0.104.1
uvicorn==0.24.0
pydantic==2.5.0
pyyaml==6.0.1
requests==2.31.0
tenacity==8.2.3
sqlalchemy==2.0.23
python-dotenv==1.0.0
```

```bash
pip install -r requirements.txt
```

---

## 阶段1：LLM客户端 (Day 1-2) ← 最优先

### 为什么先做这个？
- 所有模块都依赖LLM调用
- 验证API配置是否正确
- 后续开发可以立即测试

### 最小可用版本

```python
# llm/client.py - 最小版本
import os
import requests
from typing import Optional

class SimpleLLMClient:
    def __init__(self):
        self.api_key = os.getenv("GLM_API_KEY")
        self.base_url = "https://open.bigmodel.cn/api/paas/v4"
        self.model = "glm-4-flash"

    def generate(self, prompt: str, system_prompt: Optional[str] = None) -> str:
        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json"
        }

        messages = []
        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})
        messages.append({"role": "user", "content": prompt})

        data = {
            "model": self.model,
            "messages": messages,
            "temperature": 0.7
        }

        response = requests.post(
            f"{self.base_url}/chat/completions",
            headers=headers,
            json=data,
            timeout=60
        )
        response.raise_for_status()

        return response.json()["choices"][0]["message"]["content"]

# 测试
if __name__ == "__main__":
    client = SimpleLLMClient()
    result = client.generate("你好")
    print(result)
```

### 测试运行

```bash
# .env
GLM_API_KEY=3d1c141ae95e4704ba8a0f7f9cf4407d.EFzJmtZooKF9EpTX

python -m llm.client
```

**验证通过后再继续。**

---

## 阶段2：数据模型 (Day 2-3)

### 2.1 先定义核心数据结构

```python
# db/models.py
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any
from datetime import datetime
from enum import Enum

class EmotionType(Enum):
    CALM = "平静"
    ANGRY = "愤怒"
    SAD = "悲伤"
    HAPPY = "欢乐"
    FEAR = "恐惧"

@dataclass
class CharacterState:
    """角色动态状态"""
    character_id: str
    location: str = ""
    emotion: str = EmotionType.CALM.value
    emotion_intensity: int = 50
    health: str = "健康"
    knowledge: List[str] = field(default_factory=list)
    arc_progress: int = 0  # 0-100

@dataclass
class CharacterProfile:
    """角色静态档案"""
    character_id: str
    name: str
    background: str
    personality: List[str] = field(default_factory=list)
    abilities: Dict[str, Any] = field(default_factory=dict)

@dataclass
class WorldPackage:
    """世界设定包"""
    world_id: str
    story_soil: Dict[str, Any] = field(default_factory=dict)
    setting_constraints: Dict[str, Any] = field(default_factory=dict)

@dataclass
class NarrativeBlueprint:
    """叙事蓝图"""
    blueprint_id: str
    world_id: str
    story_outline: Dict[str, Any] = field(default_factory=dict)
    chapter_plans: List[Dict] = field(default_factory=list)
    character_arcs: Dict[str, Any] = field(default_factory=dict)

@dataclass
class SceneInstruction:
    """场景指令"""
    chapter: int
    scene: int
    purpose: str
    location: str
    characters: List[str]
    action: str
    dialogue_focus: str = ""
```

### 2.2 简单的内存存储

```python
# db/database.py
from typing import Dict, Optional
from db.models import CharacterProfile, CharacterState, WorldPackage, NarrativeBlueprint

class MemoryDatabase:
    """内存数据库 - 用于开发测试"""

    def __init__(self):
        self.characters: Dict[str, CharacterProfile] = {}
        self.character_states: Dict[str, CharacterState] = {}
        self.worlds: Dict[str, WorldPackage] = {}
        self.narratives: Dict[str, NarrativeBlueprint] = {}

    # 角色相关
    def save_character(self, character: CharacterProfile):
        self.characters[character.character_id] = character

    def get_character(self, character_id: str) -> Optional[CharacterProfile]:
        return self.characters.get(character_id)

    def update_character_state(self, character_id: str, state: CharacterState):
        self.character_states[character_id] = state

    def get_character_state(self, character_id: str) -> Optional[CharacterState]:
        return self.character_states.get(character_id)

    # 世界相关
    def save_world(self, world: WorldPackage):
        self.worlds[world.world_id] = world

    def get_world(self, world_id: str) -> Optional[WorldPackage]:
        return self.worlds.get(world_id)

    # 叙事相关
    def save_narrative(self, narrative: NarrativeBlueprint):
        self.narratives[narrative.blueprint_id] = narrative

    def get_narrative(self, blueprint_id: str) -> Optional[NarrativeBlueprint]:
        return self.narratives.get(blueprint_id)


# 全局实例
_db = MemoryDatabase()

def get_db() -> MemoryDatabase:
    return _db
```

---

## 阶段3：世界设定器 (Day 3-5)

### 3.1 最小版本：只生成哲学+世界观

```python
# modules/world_builder/builder.py
import json
from llm.client import SimpleLLMClient
from db.models import WorldPackage
from db.database import get_db
import uuid

class WorldBuilder:
    def __init__(self):
        self.client = SimpleLLMClient()
        self.db = get_db()

    def build(self, params: dict) -> WorldPackage:
        """构建世界设定"""
        world_id = f"world_{uuid.uuid4().hex[:8]}"

        # 阶段1: 生成哲学
        philosophy = self._generate_philosophy(params)

        # 阶段2: 生成世界观
        worldview = self._generate_worldview(philosophy)

        # 阶段3: 生成故事土壤
        story_soil = self._generate_story_soil(philosophy, worldview, params)

        world_package = WorldPackage(
            world_id=world_id,
            story_soil=story_soil,
            setting_constraints={
                "philosophy": philosophy,
                "worldview": worldview
            }
        )

        self.db.save_world(world_package)
        return world_package

    def _generate_philosophy(self, params: dict) -> dict:
        prompt = f"""
为以下小说生成哲学基础：
- 世界类型：{params.get('world_type', '奇幻')}
- 核心主题：{params.get('theme', '权力')}

请以JSON格式返回：
{{
    "core_question": "探讨的根本问题",
    "value_system": {{
        "highest_good": "最高善",
        "ultimate_evil": "最大恶"
    }}
}}
"""
        response = self.client.generate(
            prompt,
            system_prompt="你是世界哲学家，精通构建虚构世界的基础思想体系"
        )
        return json.loads(response)

    def _generate_worldview(self, philosophy: dict) -> dict:
        prompt = f"""
基于以下哲学基础生成世界观：
{json.dumps(philosophy, ensure_ascii=False)}

请以JSON格式返回：
{{
    "cosmology": "世界起源",
    "metaphysics": {{
        "soul_exists": true/false,
        "afterlife": "死后世界描述"
    }}
}}
"""
        response = self.client.generate(
            prompt,
            system_prompt="你是世界构建师，负责从哲学推导世界观"
        )
        return json.loads(response)

    def _generate_story_soil(self, philosophy: dict, worldview: dict, params: dict) -> dict:
        prompt = f"""
基于以下设定生成故事土壤：
哲学：{json.dumps(philosophy, ensure_ascii=False)}
世界观：{json.dumps(worldview, ensure_ascii=False)}

请以JSON格式返回：
{{
    "social_conflicts": [
        {{
            "type": "经济/政治/社会矛盾",
            "description": "描述",
            "parties": ["甲方", "乙方"]
        }}
    ],
    "potential_plot_hooks": [
        {{
            "type": "类型",
            "description": "描述",
            "story_potential": "故事潜力"
        }}
    ]
}}
"""
        response = self.client.generate(
            prompt,
            system_prompt="你是故事策划师，负责从世界设定中提取故事要素"
        )
        return json.loads(response)
```

### 测试

```python
# 测试世界设定器
from modules.world_builder.builder import WorldBuilder

builder = WorldBuilder()
world = builder.build({
    "world_type": "奇幻",
    "theme": "权力与腐败"
})

print(f"世界ID: {world.world_id}")
print(f"故事土壤: {json.dumps(world.story_soil, ensure_ascii=False, indent=2)}")
```

---

## 阶段4：叙事器 (Day 5-7)

### 4.1 最小版本：生成三幕大纲

```python
# modules/narrative_engine/engine.py
import json
from llm.client import SimpleLLMClient
from db.models import NarrativeBlueprint, WorldPackage
from db.database import get_db
import uuid

class NarrativeEngine:
    def __init__(self):
        self.client = SimpleLLMClient()
        self.db = get_db()

    def plan(self, world: WorldPackage, intent: dict) -> NarrativeBlueprint:
        """规划叙事蓝图"""
        blueprint_id = f"narr_{uuid.uuid4().hex[:8]}"

        # 生成大纲
        outline = self._generate_outline(world, intent)

        blueprint = NarrativeBlueprint(
            blueprint_id=blueprint_id,
            world_id=world.world_id,
            story_outline=outline
        )

        self.db.save_narrative(blueprint)
        return blueprint

    def _generate_outline(self, world: WorldPackage, intent: dict) -> dict:
        prompt = f"""
基于以下世界设定生成三幕剧大纲：

世界设定：
{json.dumps(world.story_soil, ensure_ascii=False)}

创作意图：
- 故事类型：{intent.get('story_type', '冒险')}
- 主角概念：{intent.get('protagonist', '一个普通少年')}

请以JSON格式返回：
{{
    "act1": {{
        "setup": "世界建立",
        "inciting_incident": "激励事件",
        "plot_point1": "第一幕情节点"
    }},
    "act2": {{
        "rising_action": ["试炼1", "试炼2"],
        "midpoint": "中点",
        "all_is_lost": "一无所有"
    }},
    "act3": {{
        "climax": "高潮",
        "resolution": "结局"
    }}
}}
"""
        response = self.client.generate(
            prompt,
            system_prompt="你是故事策划师，精通三幕剧结构"
        )
        return json.loads(response)
```

---

## 阶段5：写作器 (Day 7-10)

### 5.1 最小版本：生成场景文本

```python
# modules/writer/writer.py
import json
from llm.client import SimpleLLMClient
from db.models import SceneInstruction
from db.database import get_db

class Writer:
    def __init__(self):
        self.client = SimpleLLMClient()
        self.db = get_db()

    def write_scene(self, instruction: SceneInstruction, world_id: str) -> str:
        """生成场景文本"""
        world = self.db.get_world(world_id)

        prompt = f"""
根据以下指令生成场景内容：

场景目的：{instruction.purpose}
地点：{instruction.location}
在场角色：{', '.join(instruction.characters)}
发生事件：{instruction.action}
对话重点：{instruction.dialogue_focus}

世界设定参考：
{json.dumps(world.story_soil, ensure_ascii=False)[:500]}

要求：
1. 使用第三人称限制视角
2. 展示而非讲述
3. 对话符合角色性格
4. 字数约800字
"""
        response = self.client.generate(
            prompt,
            system_prompt="你是小说作家，擅长场景描写和对话"
        )
        return response
```

---

## 阶段6：编排器 (Day 10-12)

```python
# orchestrator/orchestrator.py
from modules.world_builder.builder import WorldBuilder
from modules.narrative_engine.engine import NarrativeEngine
from modules.writer.writer import Writer
from db.database import get_db

class Orchestrator:
    def __init__(self):
        self.db = get_db()

    def create_story(self, params: dict) -> dict:
        """创建故事（规划生成模式）"""

        # 步骤1: 世界构建
        world_builder = WorldBuilder()
        world = world_builder.build(params)
        print(f"[✓] 世界构建完成: {world.world_id}")

        # 步骤2: 叙事规划
        narrative_engine = NarrativeEngine()
        narrative = narrative_engine.plan(world, params.get('creative_intent', {}))
        print(f"[✓] 叙事规划完成: {narrative.blueprint_id}")

        # 步骤3: 生成第一章（示例）
        writer = Writer()
        # ... 调用 writer

        return {
            "world_id": world.world_id,
            "narrative_id": narrative.blueprint_id,
            "outline": narrative.story_outline
        }
```

---

## 阶段7：API接口 (Day 12-14)

```python
# api/main.py
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from orchestrator.orchestrator import Orchestrator

app = FastAPI(title="AI小说创作系统")
orchestrator = Orchestrator()

class StoryRequest(BaseModel):
    world_type: str = "奇幻"
    theme: str = "权力"
    story_type: str = "冒险"
    protagonist: str = "一个普通少年"

@app.post("/api/story/create")
async def create_story(request: StoryRequest):
    try:
        result = orchestrator.create_story({
            "world_type": request.world_type,
            "theme": request.theme,
            "creative_intent": {
                "story_type": request.story_type,
                "protagonist_concept": request.protagonist
            }
        })
        return {"status": "success", "data": result}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

---

## 今日可以立即开始的任务

### Task 1: 创建项目结构 (5分钟)

```bash
cd /home/xlei/project/xupu
mkdir -p config llm db modules/{world_builder,narrative_engine,writer} orchestrator scheduler api tests
touch config/__init__.py llm/__init__.py db/__init__.py modules/__init__.py
```

### Task 2: 创建 LLM 客户端并测试 (15分钟)

```python
# 创建 llm/client.py 并测试
# 测试命令：python -c "from llm.client import SimpleLLMClient; print(SimpleLLMClient().generate('你好'))"
```

### Task 3: 创建数据模型 (10分钟)

```python
# 创建 db/models.py 和 db/database.py
```

### Task 4: 测试世界设定器 (20分钟)

```python
# 创建 modules/world_builder/builder.py
# 运行测试
```

---

## 开发检查清单

- [ ] 项目结构创建完成
- [ ] LLM客户端测试通过
- [ ] 数据模型定义完成
- [ ] 世界设定器可生成哲学基础
- [ ] 世界设定器可生成世界观
- [ ] 世界设定器可生成故事土壤
- [ ] 叙事器可生成三幕大纲
- [ ] 写作器可生成场景文本
- [ ] 编排器可串联全流程
- [ ] API可调用

---

## 建议

1. **先跑通最小流程**：世界设定 → 叙事大纲 → 场景文本
2. **每步都要测试**：确保LLM输出格式正确
3. **使用简单的内存数据库**：后期再换PostgreSQL
4. **先完成功能**：重试、限流、成本追踪等功能后期添加

现在可以开始了吗？需要我帮你创建某个具体文件吗？
