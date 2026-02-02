# 作品管理功能 - 详细设计文档

## 目录
1. [系统架构](#系统架构)
2. [组件层次结构](#组件层次结构)
3. [数据流设计](#数据流设计)
4. [API接口设计](#api接口设计)
5. [状态管理设计](#状态管理设计)
6. [交互流程设计](#交互流程设计)
7. [数据库Schema](#数据库schema)
8. [技术实现细节](#技术实现细节)

---

## 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        前端应用 (React)                       │
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ 页面组件     │  │ 业务组件     │  │ UI组件       │        │
│  │ Pages       │  │ Features    │  │ Components   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│         │                │                    │              │
│         └────────────────┴────────────────────┘              │
│                           │                                  │
│                  ┌────────▼────────┐                         │
│                  │  状态管理层      │                         │
│                  │  (Zustand)      │                         │
│                  └────────┬────────┘                         │
│                           │                                  │
│                  ┌────────▼────────┐                         │
│                  │  API 服务层      │                         │
│                  │  (axios)        │                         │
│                  └────────┬────────┘                         │
└───────────────────────────┼─────────────────────────────────┘
                            │ HTTP
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      后端 API (Go + Gin)                     │
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ Handler     │  │ Service     │  │ Repository  │        │
│  │ 处理器      │  │ 业务逻辑     │  │ 数据访问     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│         │                │                    │              │
│         └────────────────┴────────────────────┘              │
│                           │                                  │
│                  ┌────────▼────────┐                         │
│                  │  PostgreSQL DB  │                         │
│                  └─────────────────┘                         │
└─────────────────────────────────────────────────────────────┘
```

---

## 组件层次结构

### 目录结构设计

```
web/src/
├── features/
│   └── workspace/                    # 作品管理模块
│       ├── pages/                    # 页面组件
│       │   ├── ProjectListPage.tsx   # 作品列表页
│       │   ├── ProjectDetailPage.tsx # 作品详情页
│       │   ├── ChapterEditPage.tsx   # 章节编辑页
│       │   └── ProjectSettingsPage.tsx # 作品设置页
│       │
│       ├── components/               # 业务组件
│       │   ├── ProjectCard.tsx       # 作品卡片
│       │   ├── ChapterList.tsx       # 章节列表
│       │   ├── ChapterItem.tsx       # 章节项
│       │   ├── NovelEditor.tsx       # 小说编辑器
│       │   ├── EditorToolbar.tsx     # 编辑器工具栏
│       │   ├── AIToolPanel.tsx       # AI工具面板
│       │   ├── WorldSettingPanel.tsx # 世界设定面板
│       │   ├── CharacterCard.tsx     # 角色卡片
│       │   ├── OutlineViewer.tsx     # 大纲查看器
│       │   └── ExportDialog.tsx      # 导出对话框
│       │
│       ├── hooks/                    # 自定义Hooks
│       │   ├── useProjects.ts        # 获取作品列表
│       │   ├── useProject.ts         # 获取作品详情
│       │   ├── useChapters.ts        # 章节管理
│       │   ├── useEditor.ts          # 编辑器状态
│       │   ├── useAutoSave.ts        # 自动保存
│       │   └── useAIGenerate.ts      # AI生成
│       │
│       ├── services/                 # API服务
│       │   ├── projectApi.ts         # 作品API
│       │   ├── chapterApi.ts         # 章节API
│       │   └── aiApi.ts              # AI API
│       │
│       ├── stores/                   # 状态管理
│       │   ├── projectStore.ts       # 作品状态
│       │   ├── chapterStore.ts       # 章节状态
│       │   └── editorStore.ts        # 编辑器状态
│       │
│       └── types/                    # 类型定义
│           ├── project.ts            # 作品类型
│           ├── chapter.ts            # 章节类型
│           └── editor.ts             # 编辑器类型
│
├── components/                       # 通用UI组件
│   └── ui/                           # shadcn/ui组件
│       ├── button.tsx
│       ├── input.tsx
│       ├── dialog.tsx
│       └── ...
│
├── router/                           # 路由配置
│   └── index.tsx
│
└── stores/                           # 全局状态
    └── authStore.ts                  # 认证状态
```

### 页面组件详细设计

#### 1. ProjectListPage.tsx (作品列表页)

```typescript
interface ProjectListPageProps {}

组件结构：
┌─────────────────────────────────────────┐
│ Header                                  │
│   [Logo] NovelFlow 叙谱 [用户头像]      │
├─────────────────────────────────────────┤
│                                         │
│ 页面标题 + 操作栏                        │
│   📚 我的作品      [+ 新建作品]          │
│                                         │
│ 筛选和搜索栏                             │
│   [全部▼] [创作中] [已完成] [草稿]        │
│   🔍 [搜索作品名、主角名...]              │
│                                         │
│ ┌───────────────────────────────────┐  │
│ │ ProjectGrid (作品网格)            │  │
│ │                                   │  │
│ │  ┌────────┐  ┌────────┐          │  │
│ │  │ Card 1 │  │ Card 2 │  ...     │  │
│ │  └────────┘  └────────┘          │  │
│ │                                   │  │
│ │  ┌────────┐  ┌────────┐          │  │
│ │  │ Card 3 │  │ Card 4 │  ...     │  │
│ │  └────────┘  └────────┘          │  │
│ │                                   │  │
│ └───────────────────────────────────┘  │
│                                         │
└─────────────────────────────────────────┘

状态管理：
- projects: Project[]           // 作品列表
- loading: boolean              // 加载状态
- filter: ProjectStatus | 'all'  // 筛选条件
- sortBy: 'updated' | 'created'  // 排序方式
- searchQuery: string           // 搜索关键词

副作用：
- useEffect → 加载作品列表
- useCallback → 处理筛选、搜索、删除
```

#### 2. ProjectDetailPage.tsx (作品详情页)

```typescript
interface ProjectDetailPageProps {
  projectId: string
}

组件结构：
┌─────────────────────────────────────────────────────────┐
│ 顶部导航栏                                               │
│  [◀ 返回] [作品标题] [💾 保存] [⚙️ 设置] [📤 导出]       │
├──────────────┬────────────────────────┬──────────────────┤
│              │                         │                  │
│ Sidebar      │    Main Content         │  Right Panel     │
│ (侧边栏)     │    (主内容区)            │  (右侧面板)      │
│              │                         │                  │
│ ┌──────────┐ │ ┌─────────────────────┐ │ ┌────────────┐ │
│ │ 章节列表 │ │ │   章节标题           │ │ │ 世界设定   │ │
│ │          │ │ ├─────────────────────┤ │ │            │ │
│ │ 第一章   │ │ │                     │ │ │ [折叠面板] │ │
│ │ 第二章   │ │ │ 正文内容...         │ │ │            │ │
│ │ 第三章   │ │ │                     │ │ │ 🌍 世界观  │ │
│ │ •••      │ │ │ [选中] [AI续写]     │ │ │ 🗺️ 地理    │ │
│ │          │ │ │ [扩展] [润色]       │ │ │ 👥 文明    │ │
│ │ [+ 新建] │ │ │                     │ │ │            │ │
│ │          │ │ └─────────────────────┘ │ └────────────┘ │
│ │ ━━━━━━━  │ │                         │                  │
│ │          │ │ 编辑器工具栏             │ │ ┌────────────┐ │
│ │ 📑 大纲  │ │ [B] [I] [U] [H1] [H2]  │ │ │ 角色卡片   │ │
│ │ ⚙️ 设定  │ │ [引用] [AI续写]        │ │ │            │ │
│ │ 🎭 角色  │ │                         │ │ │ 李青云    │ │
│ │          │ │ 底部信息栏              │ │ │ 林婉儿    │ │
│ └──────────┘ │ 字数: 3245 | AI: 1234  │ │ │ •••       │ │
│              │ [版本] [统计] [全屏]    │ │ └────────────┘ │
└──────────────┴────────────────────────┴──────────────────┘

布局配置：
- Sidebar: 固定宽度 280px，可折叠
- Main Content: flex-1，自适应宽度
- Right Panel: 固定宽度 320px，可折叠

子组件：
- ChapterList       (章节列表)
- NovelEditor       (富文本编辑器)
- EditorToolbar     (编辑器工具栏)
- AIToolPanel       (AI工具面板)
- WorldSettingPanel (世界设定面板)
- CharacterCard     (角色卡片)
```

#### 3. ChapterEditPage.tsx (全屏编辑页)

```typescript
interface ChapterEditPageProps {
  projectId: string
  chapterId: string
}

组件结构：
┌─────────────────────────────────────────────────────────┐
│ 沉浸式编辑模式 (全屏)                                    │
│                                                           │
│  [ESC 退出全屏]  第一章：觉醒    [💾 已保存]  [AI助手 ▼]   │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━      │
│                                                           │
│  正文内容...                                             │
│                                                           │
│  [可以继续写...光标位置]                                  │
│                                                           │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━      │
│                                                           │
│  底部工具栏 (浮动)                                         │
│  [字数: 3245] [段落数: 12] [预计阅读: 8分钟]              │
│  [AI续写] [扩展] [润色] [生成对话]                         │
└─────────────────────────────────────────────────────────┘

特性：
- 无干扰编辑界面
- 自动隐藏工具栏
- 快捷键支持
- 焦点模式
```

### 业务组件详细设计

#### 1. ProjectCard.tsx (作品卡片)

```typescript
interface ProjectCardProps {
  project: Project
  onEdit: (id: string) => void
  onDelete: (id: string) => void
  onContinue: (id: string) => void
}

组件结构：
┌─────────────────────────┐
│ ┌─────────────────────┐ │
│ │   封面图 (可选)       │ │
│ │   或默认占位图        │ │
│ └─────────────────────┘ │
│                         │
│ 📖 [作品标题]           │
│ ⭐ [总字数]万字         │
│ 📝 [章节数]章          │
│ 🎭 [状态标签]          │
│ 🕒 [更新时间]          │
│                         │
│ ┌─────────────────────┐ │
│ │ [继续创作] [详情]   │ │
│ └─────────────────────┘ │
└─────────────────────────┘

交互：
- 点击卡片 → 跳转到详情页
- 点击"继续创作" → 打开最后编辑的章节
- 悬停 → 显示更多操作菜单
```

#### 2. ChapterList.tsx (章节列表)

```typescript
interface ChapterListProps {
  chapters: Chapter[]
  currentChapterId: string
  onChapterSelect: (chapterId: string) => void
  onChapterCreate: () => void
  onChapterDelete: (chapterId: string) => void
  onChapterReorder: (chapters: Chapter[]) => void
}

组件结构：
┌─────────────────────────┐
│ 章节列表        [+ 新建] │
│ ━━━━━━━━━━━━━━━━━━━━━  │
│                         │
│ 📂 第一章：觉醒          │
│    ✏️ 3,245字           │
│    ✓ 已完成             │
│                         │
│ 📂 第二章：拜师          │
│    ✏️ 2,890字           │
│    📝 草稿              │
│                         │
│ 📂 第三章：突破          │
│    ✏️ 1,523字           │
│    📝 草稿              │
│                         │
│ •••                     │
│                         │
│ [拖拽以排序]            │
└─────────────────────────┘

功能：
- 点击章节 → 切换编辑
- 拖拽排序
- 右键菜单 → 重命名/删除
- 显示字数和状态
```

#### 3. NovelEditor.tsx (富文本编辑器)

```typescript
interface NovelEditorProps {
  content: string              // 编辑器内容
  onChange: (content: string) => void
  onSave: () => void            // 保存回调
  isLoading?: boolean           // AI生成中
  readOnly?: boolean            // 只读模式
  placeholder?: string          // 占位文本
}

基于 Tiptap 实现：
import { Editor } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'

组件结构：
┌─────────────────────────────────┐
│ 编辑器工具栏 (可选显示)          │
│ [B] [I] [U] [H1] [引用] [列表]  │
├─────────────────────────────────┤
│                                 │
│ 正文内容...                     │
│                                 │
│ 光标位置 │                      │
│                                 │
│ [AI续写建议浮窗]                 │
│                                 │
└─────────────────────────────────┘

功能：
- 基础文本格式化
- 撤销/重做
- 快捷键支持
- AI续写集成
- 自动保存提示
```

#### 4. AIToolPanel.tsx (AI工具面板)

```typescript
interface AIToolPanelProps {
  projectId: string
  chapterId: string
  selectedText?: string           // 选中文本
  onGenerate: (type: string, params: any) => Promise<void>
}

组件结构：
┌─────────────────────────┐
│ 🤖 AI 助手              │
│ ━━━━━━━━━━━━━━━━━━━━━  │
│                         │
│ 📝 快速操作             │
│   [续写下一段]          │
│   [扩展选中文本]        │
│   [润色文字]            │
│   [生成对话]            │
│                         │
│ 🌍 设定参考             │
│   [世界观] [角色]       │
│                         │
│ 📜 剧情大纲             │
│   当前章节位置          │
│   前情提要              │
│   后续规划              │
│                         │
│ ⚙️ 生成参数             │
│   风格: [平衡▼]         │
│   长度: [中等▼]         │
│                         │
└─────────────────────────┘

功能：
- AI生成操作
- 参考资料快速查看
- 生成参数调整
- 历史记录查看
```

---

## 数据流设计

### 1. 作品列表数据流

```
用户操作 → 页面组件 → Hook → API调用 → 后端 → 数据库
   │         │         │        │        │        │
   └─────────┴─────────┴────────┴────────┴────────┘
                     ↓
                更新状态
                     ↓
                重新渲染

详细流程：
1. 用户访问 /projects
2. ProjectListPage 组件挂载
3. useProjects Hook 被调用
4. 发起 API 请求: GET /api/v1/projects
5. 后端返回作品列表
6. 更新 projectStore 状态
7. 组件重新渲染，显示作品列表
```

### 2. 作品详情数据流

```
路径: /projects/:projectId

数据流：
┌──────────────┐
│  URL 参数     │ → projectId
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ProjectDetail │
│   Page       │
└──────┬───────┘
       │
       ├──────────────────────────────────┐
       │                                  │
       ▼                                  ▼
┌──────────────┐                  ┌──────────────┐
│ useProject   │                  │ useChapters  │
│              │                  │              │
│ GET          │                  │ GET          │
│ /projects/:id│                  │ /projects/:id │
│              │                  │   /chapters  │
└──────┬───────┘                  └──────┬───────┘
       │                                  │
       │                                  │
       ▼                                  ▼
┌──────────────┐                  ┌──────────────┐
│ projectStore │                  │chapterStore  │
│              │                  │              │
│ project: {} │                  │chapters: []  │
└──────────────┘                  └──────────────┘
       │                                  │
       └──────────────┬───────────────────┘
                      │
                      ▼
              ┌──────────────┐
              │   页面渲染    │
              └──────────────┘
```

### 3. 章节编辑数据流

```
用户编辑 → 编辑器 → editorStore → 防抖处理 → 自动保存Hook
   │        │        │            │             │
   │        │        │            │             ▼
   │        │        │            │      API: PUT /chapters/:id
   │        │        │            │             │
   │        │        │            │             ▼
   │        │        │            │        更新数据库
   │        │        │            │             │
   │        │        │            └─────────────┘
   │        │        │
   └────────┴────────┴────→ 更新 UI (保存成功提示)

时间线：
T+0s:  用户输入
T+1s:  更新 editorStore (实时)
T+5s:  触发自动保存 (防抖)
T+6s:  API 请求完成
T+7s:  UI 更新为"已保存"
```

### 4. AI生成数据流

```
用户点击"AI续写"
   │
   ▼
AIToolPanel 检查上下文
   │
   ├─→ 获取当前章节内容
   ├─→ 获取世界设定
   ├─→ 获取角色信息
   └─→ 获取剧情大纲
   │
   ▼
构造生成请求
   │
   ▼
API: POST /api/v1/projects/:id/chapters/:chapterId/generate/continue
   │
   ├─→ 传递上下文参数
   ├─→ 传递生成配置
   └─→ 传递引用内容
   │
   ▼
后端处理
   │
   ├─→ 调用 AI 服务
   ├─→ 流式返回结果
   └─→ 保存到数据库
   │
   ▼
前端接收 (流式)
   │
   ├─→ 实时更新编辑器
   ├─→ 显示生成进度
   └─→ 更新字数统计
   │
   ▼
生成完成
   │
   └─→ 保存章节内容
```

---

## API接口设计

### 1. 作品管理 API

```typescript
// 基础URL: http://localhost:8080/api/v1

/**
 * 获取作品列表
 * GET /projects
 */
interface GetProjectsParams {
  page?: number
  pageSize?: number
  status?: 'draft' | 'building' | 'generating' | 'completed' | 'paused' | 'failed'
  sortBy?: 'created_at' | 'updated_at' | 'total_words'
  sortOrder?: 'asc' | 'desc'
  search?: string
}

interface GetProjectsResponse {
  success: true
  data: {
    projects: Project[]
    total: number
    page: number
    pageSize: number
  }
}

/**
 * 获取作品详情
 * GET /projects/:id
 */
interface GetProjectResponse {
  success: true
  data: {
    project: Project
    world: WorldSetting
    narrative: NarrativeBlueprint
    statistics: {
      totalWords: number
      totalChapters: number
      lastGeneratedAt: string
    }
  }
}

/**
 * 创建作品
 * POST /projects
 */
interface CreateProjectRequest {
  name: string
  description?: string
  mode: 'planning' | 'intervention' | 'random' | 'story_core' | 'short' | 'script'
  tags?: string[]
}

/**
 * 更新作品
 * PUT /projects/:id
 */
interface UpdateProjectRequest {
  name?: string
  description?: string
  tags?: string[]
  coverImage?: string
  isPublic?: boolean
}

/**
 * 删除作品
 * DELETE /projects/:id
 */
interface DeleteProjectResponse {
  success: true
  data: {
    message: string
  }
}
```

### 2. 章节管理 API

```typescript
/**
 * 获取章节列表
 * GET /projects/:projectId/chapters
 */
interface GetChaptersResponse {
  success: true
  data: {
    chapters: Chapter[]
    total: number
  }
}

/**
 * 创建章节
 * POST /projects/:projectId/chapters
 */
interface CreateChapterRequest {
  title: string
  content?: string
  chapterNum?: number  // 自动计算
}

/**
 * 更新章节
 * PUT /projects/:projectId/chapters/:chapterId
 */
interface UpdateChapterRequest {
  title?: string
  content?: string
  status?: 'draft' | 'completed'
}

/**
 * 删除章节
 * DELETE /projects/:projectId/chapters/:chapterId
 */

/**
 * 重新排序章节
 * PUT /projects/:projectId/chapters/reorder
 */
interface ReorderChaptersRequest {
  chapterIds: string[]  // 按新顺序排列的ID数组
}
```

### 3. AI生成 API

```typescript
/**
 * AI续写
 * POST /projects/:projectId/chapters/:chapterId/generate/continue
 */
interface ContinueChapterRequest {
  context: {
    content: string        // 当前内容
    wordCount: number      // 已生成字数
  }
  params: {
    length: 'short' | 'medium' | 'long'  // 生成长度
    style: 'balanced' | 'creative' | 'formal'  // 风格
    includeDialogue?: boolean  // 是否包含对话
    includeAction?: boolean   // 是否包含动作
  }
  references?: {
    worldSetting?: boolean   // 引用世界设定
    characters?: string[]     // 引用角色
    outline?: string          // 引用大纲
  }
}

interface ContinueChapterResponse {
  success: true
  data: {
    generatedContent: string
    wordCount: number
    tokensUsed: number
  }
}

// 流式响应版本
interface ContinueChapterStreamResponse {
  success: true
  data: {
    content: string       // 分块内容
    done: boolean         // 是否完成
    wordCount: number     // 当前字数
  }
}

/**
 * AI扩展
 * POST /projects/:projectId/chapters/:chapterId/generate/expand
 */
interface ExpandTextRequest {
  text: string              // 要扩展的文本
  expandBy: number          // 扩展倍数
  style: 'detailed' | 'descriptive' | 'emotional'
}

/**
 * AI润色
 * POST /projects/:projectId/chapters/:chapterId/generate/polish
 */
interface PolishTextRequest {
  text: string
  style: 'smooth' | 'literary' | 'dramatic'
  preserveOriginal: boolean
}
```

### 4. 导出 API

```typescript
/**
 * 导出作品
 * POST /projects/:projectId/export
 */
interface ExportProjectRequest {
  format: 'txt' | 'epub' | 'pdf' | 'docx'
  options: {
    includeFrontMatter?: boolean   // 包含封面
    includeOutline?: boolean       // 包含大纲
    includeWorldSetting?: boolean  // 包含设定
    chapterNumbers?: boolean       // 章节编号
  }
}

interface ExportProjectResponse {
  success: true
  data: {
    downloadUrl: string    // 下载链接
    expiresAt: string      // 过期时间
    fileSize: number       // 文件大小
  }
}
```

---

## 状态管理设计

### Zustand Store 结构

```typescript
// stores/projectStore.ts
interface ProjectStore {
  // 状态
  projects: Project[]
  currentProject: Project | null
  loading: boolean
  error: string | null

  // 分页
  pagination: {
    page: number
    pageSize: number
    total: number
  }

  // 筛选
  filters: {
    status: ProjectStatus | 'all'
    search: string
    sortBy: 'created_at' | 'updated_at' | 'total_words'
  }

  // 操作
  fetchProjects: () => Promise<void>
  fetchProject: (id: string) => Promise<void>
  createProject: (data: CreateProjectRequest) => Promise<Project>
  updateProject: (id: string, data: UpdateProjectRequest) => Promise<void>
  deleteProject: (id: string) => Promise<void>
  setCurrentProject: (project: Project | null) => void

  // 筛选操作
  setFilter: (filter: Partial<ProjectStore['filters']>) => void
  resetFilters: () => void
}

// stores/chapterStore.ts
interface ChapterStore {
  // 状态
  chapters: Chapter[]
  currentChapter: Chapter | null
  loading: boolean
  error: string | null

  // 操作
  fetchChapters: (projectId: string) => Promise<void>
  createChapter: (projectId: string, data: CreateChapterRequest) => Promise<Chapter>
  updateChapter: (projectId: string, chapterId: string, data: UpdateChapterRequest) => Promise<void>
  deleteChapter: (projectId: string, chapterId: string) => Promise<void>
  reorderChapters: (projectId: string, chapterIds: string[]) => Promise<void>

  // 本地状态
  setCurrentChapter: (chapter: Chapter | null) => void
  updateLocalChapter: (chapterId: string, updates: Partial<Chapter>) => void
}

// stores/editorStore.ts
interface EditorStore {
  // 编辑器状态
  content: string
  wordCount: number
  isDirty: boolean
  isSaving: boolean
  lastSavedAt: Date | null
  autoSaveEnabled: boolean

  // AI生成状态
  isGenerating: boolean
  generatedContent: string

  // 编辑器配置
  editorConfig: {
    fontSize: number
    lineHeight: number
    maxWidth: number
    theme: 'light' | 'dark'
  }

  // 操作
  setContent: (content: string) => void
  updateContent: (content: string) => void
  save: () => Promise<void>
  reset: () => void

  // AI操作
  startGeneration: () => void
  updateGeneratedContent: (content: string) => void
  finishGeneration: () => void

  // 配置
  updateConfig: (config: Partial<EditorStore['editorConfig']>) => void
}
```

### 跨组件通信示例

```typescript
// 场景1：从作品列表跳转到编辑页
// ProjectListPage.tsx
const handleContinue = (projectId: string) => {
  // 1. 设置当前作品
  projectStore.setCurrentProject(
    projectStore.projects.find(p => p.id === projectId)!
  )

  // 2. 跳转到作品详情页
  navigate(`/projects/${projectId}`)
}

// ProjectDetailPage.tsx
const { currentProject } = projectStore
const { chapters, fetchChapters } = chapterStore

useEffect(() => {
  if (currentProject) {
    // 加载章节列表
    fetchChapters(currentProject.id)

    // 如果有最后编辑的章节，自动选中
    if (currentProject.currentChapter) {
      chapterStore.setCurrentChapter(
        chapters.find(c => c.id === currentProject.currentChapter)!
      )
    }
  }
}, [currentProject])

// 场景2：编辑器自动保存
// ChapterEditPage.tsx
const { content, save, isSaving, isDirty } = editorStore

// 使用自动保存Hook
useAutoSave({
  content,
  onSave: save,
  delay: 5000,  // 5秒防抖
  enabled: isDirty && !isSaving
})

// useAutoSave Hook实现
function useAutoSave({ content, onSave, delay, enabled }) {
  const [lastSavedContent, setLastSavedContent] = useState(content)

  useEffect(() => {
    if (!enabled) return

    const timer = setTimeout(async () => {
      if (content !== lastSavedContent) {
        await onSave()
        setLastSavedContent(content)
      }
    }, delay)

    return () => clearTimeout(timer)
  }, [content, enabled, delay, lastSavedContent, onSave])
}

// 场景3：AI生成实时更新
// AIToolPanel.tsx
const handleContinue = async () => {
  const { startGeneration, updateGeneratedContent, finishGeneration } = editorStore

  startGeneration()

  try {
    // 流式接收
    const response = await fetch('/api/v1/ai/generate/continue', {
      method: 'POST',
      body: JSON.stringify({ context: editorStore.content }),
    })

    const reader = response.body.getReader()
    const decoder = new TextDecoder()

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      const chunk = decoder.decode(value)
      updateGeneratedContent(chunk)
    }

    finishGeneration()
  } catch (error) {
    // 错误处理
  }
}

// NovelEditor.tsx
const { generatedContent } = editorStore

// 实时追加生成的内容到编辑器
useEffect(() => {
  if (generatedContent) {
    editorStore.updateContent(
      editorStore.content + generatedContent
    )
  }
}, [generatedContent])
```

---

## 交互流程设计

### 流程1: 创建新作品

```
用户操作流程：
1. 点击"新建作品"按钮
   ↓
2. 显示创建对话框
   ┌─────────────────────┐
   │  创建新作品         │
   ├─────────────────────┤
   │  作品名称: [_____]  │
   │  作品简介: [_____]  │
   │  创作模式: [下拉]    │
   │  标签: [_____]      │
   │                     │
   │  [取消]  [创建]     │
   └─────────────────────┘
   ↓
3. 填写信息，点击"创建"
   ↓
4. API调用: POST /api/v1/projects
   ↓
5. 创建成功，跳转到作品详情页
   ↓
6. 提示："开始创建世界设定吧！"
   ↓
7. 引导用户完成初始设置
   - [第一步：创建世界设定]
   - [第二步：规划故事大纲]
   - [第三步：开始创作章节]
```

### 流程2: 章节编辑流程

```
用户操作流程：
1. 进入作品详情页
   ↓
2. 左侧显示章节列表
   - 第一章 (已完成 ✓)
   - 第二章 (草稿 📝)
   - [+ 新建章节]
   ↓
3. 点击"第二章"开始编辑
   ↓
4. 右侧加载编辑器，显示章节内容
   ↓
5. 用户开始编辑
   - 实时保存到 editorStore
   - 5秒后自动保存到后端
   ↓
6. 编辑器工具栏功能
   - [B] [I] [U] - 文字格式
   - [AI续写] - 调用AI生成
   - [扩展] - 扩展选中文字
   - [润色] - AI润色
   ↓
7. 右侧AI工具面板
   - 显示世界设定参考
   - 显示角色卡片
   - 显示剧情大纲
   ↓
8. 完成编辑
   - 自动保存
   - 更新章节状态
   - 更新字数统计
```

### 流程3: AI生成流程

```
用户操作流程：
1. 用户在编辑器选中一段文字
   ↓
2. 点击"AI续写"按钮
   ↓
3. 显示AI参数配置弹窗
   ┌─────────────────────┐
   │  AI 续写设置        │
   ├─────────────────────┤
   │  生成长度: [中等▼]  │
   │  风格倾向: [平衡▼]  │
   │  包含对话: [✓]      │
   │  包含动作: [✓]      │
   │                     │
   │  [取消]  [开始生成] │
   └─────────────────────┘
   ↓
4. 点击"开始生成"
   ↓
5. 编辑器显示"AI正在生成..."
   - 添加加载动画
   - 禁用编辑
   ↓
6. 后端处理
   - 获取上下文（当前章节、世界设定、角色）
   - 调用AI服务
   - 流式返回结果
   ↓
7. 前端接收（流式）
   - 实时追加到编辑器
   - 显示生成进度
   - 更新字数统计
   ↓
8. 生成完成
   - 保存到数据库
   - 启用编辑
   - 显示"生成完成"提示
   ↓
9. 用户可以：
   - 继续编辑
   - 重新生成
   - 撤销生成
```

### 流程4: 导出作品

```
用户操作流程：
1. 点击"导出"按钮
   ↓
2. 显示导出对话框
   ┌─────────────────────┐
   │  导出作品           │
   ├─────────────────────┤
   │  文件格式:           │
   │  ○ TXT              │
   │  ○ EPUB             │
   │  ○ PDF              │
   │  ○ DOCX             │
   │                     │
   │  包含内容:           │
   │  [✓] 封面           │
   │  [✓] 大纲           │
   │  [✓] 世界设定       │
   │  [ ] 章节编号       │
   │                     │
   │  [预览]  [导出]     │
   └─────────────────────┘
   ↓
3. 选择格式和选项，点击"导出"
   ↓
4. API调用: POST /api/v1/projects/:id/export
   ↓
5. 后端生成文件
   - 组装内容
   - 格式转换
   - 上传到存储
   ↓
6. 返回下载链接
   {
     downloadUrl: "https://storage.example.com/exports/xxx.pdf",
     expiresAt: "2026-01-26T10:00:00Z",
     fileSize: 1024000
   }
   ↓
7. 前端自动下载
   - 或显示"下载已准备好"
   - 提供下载按钮
   ↓
8. 完成导出
```

---

## 数据库Schema

### 新增表结构

```sql
-- 章节表
CREATE TABLE chapters (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    chapter_num INTEGER NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    word_count INTEGER DEFAULT 0,
    ai_generated_word_count INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'completed')),

    -- 生成元数据
    generated_at TIMESTAMP,
    generation_params JSONB,

    -- 版本控制
    version INTEGER DEFAULT 1,
    previous_version_id TEXT REFERENCES chapters(id),

    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 索引
    CONSTRAINT unique_chapter_project UNIQUE (project_id, chapter_num)
);

CREATE INDEX idx_chapters_project ON chapters(project_id);
CREATE INDEX idx_chapters_status ON chapters(status);
CREATE INDEX idx_chapters_created ON chapters(created_at DESC);

-- 章节版本历史表（可选）
CREATE TABLE chapter_versions (
    id TEXT PRIMARY KEY,
    chapter_id TEXT NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    content TEXT NOT NULL,
    word_count INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL, -- user or AI
    note TEXT,

    UNIQUE (chapter_id, version)
);

CREATE INDEX idx_chapter_versions_chapter ON chapter_versions(chapter_id, version);

-- 扩展projects表
ALTER TABLE projects ADD COLUMN cover_image TEXT;
ALTER TABLE projects ADD COLUMN total_words INTEGER DEFAULT 0;
ALTER TABLE projects ADD COLUMN total_chapters INTEGER DEFAULT 0;
ALTER TABLE projects ADD COLUMN current_chapter_id TEXT REFERENCES chapters(id);
ALTER TABLE projects ADD COLUMN tags TEXT;
ALTER TABLE projects ADD COLUMN is_public BOOLEAN DEFAULT FALSE;
```

### 数据关系图

```
users (用户表)
  │
  ├────── 1:N
  │
  ▼
projects (作品表)
  │
  ├────── 1:1           ┌──────────────┐
  │                       │ world_settings│
  ├────── 1:1             └──────────────┘
  │
  ├────── 1:N
  │
  ▼
chapters (章节表)
  │
  ├────── 1:N (版本历史)
  │
  ▼
chapter_versions (章节版本表)

关联关系：
- user → projects (一对多)
- project → world_setting (一对一)
- project → narrative_blueprint (一对一)
- project → chapters (一对多)
- chapter → chapter_versions (一对多)
```

---

## 技术实现细节

### 1. 编辑器实现 (Tiptap)

```typescript
// features/workspace/components/NovelEditor.tsx

import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import CharacterCount from '@tiptap/extension-character-count'
import Collaboration from '@tiptap/extension-collaboration'
import CollaborationCursor from '@tiptap/extension-collaboration-cursor'

export function NovelEditor({ content, onChange, onSave }: NovelEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        heading: {
          levels: [1, 2, 3],
        },
      }),
      Placeholder.configure({
        placeholder: '开始创作您的故事...',
      }),
      CharacterCount,
      // 未来可以添加协作
      // Collaboration.configure({
      //   document: document,
      // }),
      // CollaborationCursor.configure({
      //   provider: wsProvider,
      //   user: currentUser,
      // }),
    ],
    content,
    onUpdate: ({ editor }) => {
      const html = editor.getHTML()
      onChange(html)
    },
  })

  // 快捷键
  useEffect(() => {
    if (!editor) return

    const handleKeyDown = (e: KeyboardEvent) => {
      // Ctrl+S 保存
      if (e.ctrlKey && e.key === 's') {
        e.preventDefault()
        onSave()
      }

      // Ctrl+B 加粗
      if (e.ctrlKey && e.key === 'b') {
        e.preventDefault()
        editor.chain().focus().toggleBold().run()
      }

      // Ctrl+I 斜体
      if (e.ctrlKey && e.key === 'i') {
        e.preventDefault()
        editor.chain().focus().toggleItalic().run()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [editor, onSave])

  if (!editor) {
    return <div>加载编辑器...</div>
  }

  return (
    <div className="novel-editor">
      <EditorToolbar editor={editor} />
      <EditorContent editor={editor} />
      <EditorFooter editor={editor} />
    </div>
  )
}
```

### 2. 自动保存实现

```typescript
// features/workspace/hooks/useAutoSave.ts

import { useEffect, useRef } from 'react'
import { useEditorStore } from '@/stores/editorStore'

interface UseAutoSaveOptions {
  content: string
  onSave: () => Promise<void>
  delay?: number  // 防抖延迟（毫秒）
  enabled?: boolean
}

export function useAutoSave({
  content,
  onSave,
  delay = 5000,
  enabled = true
}: UseAutoSaveOptions) {
  const { isSaving, lastSavedAt } = useEditorStore()
  const saveTimerRef = useRef<NodeJS.Timeout>()
  const lastSavedContentRef = useRef(content)

  useEffect(() => {
    if (!enabled) return

    // 清除之前的定时器
    if (saveTimerRef.current) {
      clearTimeout(saveTimerRef.current)
    }

    // 设置新的定时器
    saveTimerRef.current = setTimeout(async () => {
      // 只在内容变化时保存
      if (content !== lastSavedContentRef.current) {
        await onSave()
        lastSavedContentRef.current = content
      }
    }, delay)

    // 清理函数
    return () => {
      if (saveTimerRef.current) {
        clearTimeout(saveTimerRef.current)
      }
    }
  }, [content, delay, enabled, onSave])

  return {
    isSaving,
    lastSavedAt,
    hasUnsavedChanges: content !== lastSavedContentRef.current,
  }
}
```

### 3. AI流式生成实现

```typescript
// features/workspace/hooks/useAIGenerate.ts

import { useState, useCallback } from 'react'

interface UseAIGenerateOptions {
  projectId: string
  chapterId: string
}

export function useAIGenerate({ projectId, chapterId }: UseAIGenerateOptions) {
  const [isGenerating, setIsGenerating] = useState(false)
  const [generatedContent, setGeneratedContent] = useState('')
  const [error, setError] = useState<string | null>(null)

  const generateContinue = useCallback(async (context: string) => {
    setIsGenerating(true)
    setGeneratedContent('')
    setError(null)

    try {
      const response = await fetch(
        `/api/v1/projects/${projectId}/chapters/${chapterId}/generate/continue`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ context }),
        }
      )

      if (!response.ok) {
        throw new Error('生成失败')
      }

      // 读取流式响应
      const reader = response.body?.getReader()
      const decoder = new TextDecoder()

      if (!reader) {
        throw new Error('无法读取响应')
      }

      let fullContent = ''

      while (true) {
        const { done, value } = await reader.read()

        if (done) break

        const chunk = decoder.decode(value)
        fullContent += chunk
        setGeneratedContent(fullContent)
      }

      return fullContent
    } catch (err) {
      setError(err instanceof Error ? err.message : '未知错误')
      throw err
    } finally {
      setIsGenerating(false)
    }
  }, [projectId, chapterId])

  return {
    isGenerating,
    generatedContent,
    error,
    generateContinue,
  }
}
```

### 4. 章节拖拽排序实现

```typescript
// features/workspace/components/ChapterList.tsx

import { DndContext, closestCenter } from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy, useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

function ChapterList({ chapters, onReorder }: ChapterListProps) {
  const handleDragEnd = async (event: any) => {
    const { active, over } = event

    if (active.id !== over.id) {
      const oldIndex = chapters.findIndex((c) => c.id === active.id)
      const newIndex = chapters.findIndex((c) => c.id === over.id)

      // 重新排序数组
      const newChapters = arrayMove(chapters, oldIndex, newIndex)

      // 更新章节序号
      const reorderedChapters = newChapters.map((chapter, index) => ({
        ...chapter,
        chapterNum: index + 1,
      }))

      // 调用API保存新顺序
      await onReorder(reorderedChapters)
    }
  }

  return (
    <DndContext
      collisionDetection={closestCenter}
      onDragEnd={handleDragEnd}
    >
      <SortableContext
        items={chapters.map(c => c.id)}
        strategy={verticalListSortingStrategy}
      >
        {chapters.map((chapter) => (
          <SortableChapter key={chapter.id} chapter={chapter} />
        ))}
      </SortableContext>
    </DndContext>
  )
}

function SortableChapter({ chapter }: { chapter: Chapter }) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
  } = useSortable({ id: chapter.id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}>
      {/* 章节内容 */}
    </div>
  )
}
```

### 5. 虚拟滚动优化（大数据量）

```typescript
// features/workspace/components/VirtualizedChapterList.tsx

import { useVirtualizer } from '@tanstack/react-virtual'

function VirtualizedChapterList({ chapters }: { chapters: Chapter[] }) {
  const parentRef = useRef<HTMLDivElement>(null)

  const rowVirtualizer = useVirtualizer({
    count: chapters.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 80,  // 预估每项高度
    overscan: 5,  // 额外渲染的项数
  })

  return (
    <div ref={parentRef} style={{ height: '600px', overflow: 'auto' }}>
      <div
        style={{
          height: `${rowVirtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {rowVirtualizer.getVirtualItems().map((virtualRow) => (
          <div
            key={virtualRow.key}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualRow.size}px`,
              transform: `translateY(${virtualRow.start}px)`,
            }}
          >
            <ChapterItem chapter={chapters[virtualRow.index]} />
          </div>
        ))}
      </div>
    </div>
  )
}
```

---

## 实施计划

### 第一阶段 (2周) - MVP
- [ ] 作品列表页
- [ ] 创建/删除作品
- [ ] 章节列表
- [ ] 基础文本编辑器
- [ ] 自动保存

### 第二阶段 (2周) - 增强
- [ ] 富文本编辑器 (Tiptap)
- [ ] AI续写功能
- [ ] 章节拖拽排序
- [ ] 作品设置页

### 第三阶段 (2周) - 高级
- [ ] 版本历史
- [ ] AI工具面板
- [ ] 导出功能
- [ ] 协作编辑

---

**文档版本**: 1.0
**创建日期**: 2026-01-25
**最后更新**: 2026-01-25
