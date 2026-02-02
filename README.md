# NovelFlow 叙谱

<div align="center">

**智能AI驱动的小说创作平台**

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![JavaScript](https://img.shields.io/badge/JavaScript-ES6+-F7DF1E?style=flat&logo=javascript)](https://developer.mozilla.org/zh-CN/docs/Web/JavaScript)
[![Bootstrap](https://img.shields.io/badge/Bootstrap-5.3-7952B3?style=flat&logo=bootstrap)](https://getbootstrap.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

## 项目简介

**NovelFlow 叙谱** 是一个商业化AI小说创作平台，采用单一Go服务架构，通过智能AI技术帮助作者从创意到成稿完成小说创作。

### 核心特性

- **🌍 世界构建器** - 7阶段演化系统构建有机、自洽的虚构世界
- **📖 叙事引擎** - 基于叙事理论规划情节结构和冲突弧光
- **✍️ 智能写作** - 场景化文本生成，维护角色和世界观一致性
- **🎼 多种创作模式** - 规划生成、干预生成、随机生成、故事核、短篇、剧本
- **⚙️ 异步调度** - 支持多项目并发的任务调度系统

### 技术架构

```
┌─────────────────────────────────────────────────────────────┐
│                  Frontend (纯JavaScript)                     │
│       HTML5 + CSS3 + ES6+ + Bootstrap 5 (SPA)              │
└────────────────────────────┬────────────────────────────────┘
                             │ REST API
                             ↓
┌─────────────────────────────────────────────────────────────┐
│                       Backend (Go)                          │
│  Gin + GORM + PostgreSQL + Zhipu AI (GLM)                 │
│  + 静态文件服务 (单一服务部署)                              │
└────────────────────────────┬────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        ↓                    ↓                    ↓
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ World Builder│    │  Narrative   │    │    Writer    │
│  世界构建器  │    │   Engine     │    │   写作器     │
└──────────────┘    └──────────────┘    └──────────────┘
```

---

## 目录结构

```
novelflow/
├── cmd/                    # CLI工具和命令
│   └── api/               # API服务入口
├── internal/               # 内部包
│   ├── api/               # HTTP API处理器
│   ├── handlers/          # 请求处理器
│   ├── middleware/        # 中间件
│   └── models/            # 数据模型
├── pkg/                    # 公共包
│   ├── worldbuilder/      # 世界构建器
│   ├── narrative/         # 叙事引擎
│   ├── orchestrator/      # 编排器
│   ├── writer/            # 写作器
│   ├── llm/               # LLM客户端
│   └── db/                # 数据库层
├── static/                 # 前端静态文件
│   ├── index.html         # 主HTML文件
│   ├── css/               # 样式文件
│   └── js/                # JavaScript文件
│       ├── app.js         # 应用入口
│       ├── router.js      # 路由系统
│       ├── store.js       # 状态管理
│       ├── api.js         # API服务
│       ├── components/    # 组件
│       └── pages/         # 页面
├── docs/                   # 文档
│   ├── AI小说系统-架构概览.md
│   ├── AI小说系统-完整实现文档.md
│   └── AI小说系统-开发路线图.md
├── config/                 # 配置文件
├── bin/                    # 编译输出
├── go.mod
└── README.md
```

---

## 快速开始

### 环境要求

- **Go**: 1.23+
- **PostgreSQL**: 14+
- **智谱AI API Key**: [https://open.bigmodel.cn/](https://open.bigmodel.cn/)

### 启动服务

1. **克隆项目**

```bash
git clone https://github.com/xlei/xupu.git
cd xupu
```

2. **配置环境变量**

创建 `.env` 文件：

```env
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=novelflow

# 智谱AI配置
ZHIPU_API_KEY=your_zhipu_api_key
```

3. **编译并启动**

```bash
# 安装Go依赖
go mod download

# 编译
go build -o bin/api cmd/api/main.go

# 启动服务（包含前端静态文件服务）
./bin/api
```

服务将运行在 `http://localhost:8080`

**单一服务优势**：
- ✅ 无需单独启动前端服务器
- ✅ 无需安装Node.js和npm依赖
- ✅ 无需前端构建步骤
- ✅ 单一端口访问，简化部署

---

## 系统架构

NovelFlow 采用模块化分层架构，核心组件包括：

### 1. 调度器 (Scheduler)
系统入口，负责创建和管理编排器实例，支持多项目并发。

### 2. 编排器 (Orchestrator)
单部小说的编排中心，协调世界构建、叙事规划和写作生成。

**6种创作模式：**
- **规划生成**: 完整流程，从参数搜集到成稿
- **干预生成**: 支持用户在任意节点干预
- **随机生成**: 基于参数随机生成创意
- **故事核**: 从一句话创意扩写
- **短篇模式**: 一次性生成短篇小说
- **剧本模式**: 按剧本格式输出

### 3. 世界构建器 (World Builder)
7阶段演化系统构建虚构世界：
1. 哲学思考
2. 世界观推导
3. 法则设定（物理/魔法体系）
4. 地理环境
5. 文明社会
6. 历史演化
7. 一致性检查

### 4. 叙事引擎 (Narrative Engine)
基于叙事理论规划情节：
- 生成整体大纲（三幕结构、英雄之旅）
- 细化章节规划
- 生成场景序列
- 角色弧光规划

### 5. 写作器 (Writer)
场景化文本生成：
- 维护角色动态状态
- 生成对话和描写
- 应用视角过滤
- 保持风格一致性

### 6. 知识数据服务 (Knowledge Service)
提供写作规则和知识素材，支持自我扩充。

---

## 核心功能

### 世界构建
- 支持多种世界类型：奇幻、科幻、历史、都市、武侠、仙侠、混合
- 灵活的世界规模：村庄、城市、国家、大陆、星球、宇宙
- 自动推导哲学→世界观→法则→地理→文明→历史的依赖链

### 叙事规划
- 整合成熟叙事理论（三幕结构、英雄之旅、救猫咪等）
- 自动生成章节大纲和场景序列
- 角色弧光规划和冲突设计

### 智能写作
- 场景化内容生成，保持上下文一致性
- 角色状态实时维护和更新
- 对话、动作、描写的智能组合
- 支持多种视角和写作风格

---

## 文档

- **[架构概览](docs/AI小说系统-架构概览.md)** - 快速了解系统架构和模块功能
- **[完整实现文档](docs/AI小说系统-完整实现文档.md)** - 详细的模块设计和实现规范
- **[开发路线图](docs/AI小说系统-开发路线图.md)** - 功能规划和开发进度
- **[7阶段演化系统](docs/7阶段演化系统实现报告.md)** - 世界构建器详细说明

---

## 技术栈

### 后端
- **语言**: Go 1.23+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL
- **LLM**: 智谱AI GLM-4
- **配置管理**: YAML
- **静态文件**: 嵌入式服务

### 前端
- **语言**: JavaScript (ES6+)
- **UI框架**: Bootstrap 5
- **编辑器**: Quill.js
- **架构**: SPA (单页应用)
- **状态管理**: 自定义Store
- **路由**: 自定义Router
- **HTTP**: Fetch API

---

## 开发路线

### ✅ 已完成
- [x] 基础架构搭建
- [x] 7阶段世界构建系统
- [x] 叙事引擎基础框架
- [x] 写作器核心功能
- [x] 前端UI框架和组件库
- [x] 用户认证系统

### 🚧 开发中
- [ ] 完整的编排器实现
- [ ] 前端工作台功能
- [ ] AI工具集（拆书分析、黄金开头等）
- [ ] 异步任务调度优化

### 📋 计划中
- [ ] 订阅和支付系统
- [ ] 后台管理面板
- [ ] 多LLM模型支持
- [ ] 移动端适配

---

## 贡献指南

欢迎贡献代码、报告问题或提出新功能建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 联系方式

- **项目主页**: [https://github.com/xlei/xupu](https://github.com/xlei/xupu)
- **问题反馈**: [GitHub Issues](https://github.com/xlei/xupu/issues)

---

<div align="center">

**Made with ❤️ by NovelFlow Team**

</div>
