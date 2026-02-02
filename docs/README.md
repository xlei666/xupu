# NovelFlow 叙谱 - 文档中心

欢迎来到 NovelFlow 叙谱的文档中心！这里汇集了项目的所有技术文档。

---

## 📚 快速导航

### 入门文档
- **[项目主 README](../README.md)** - 项目概述、快速开始和架构介绍
- **[前端快速启动指南](../web/QUICKSTART.md)** - 前端开发环境和快速上手
- **[前端开发指南](../web/README.md)** - 前端项目结构和开发规范

### 核心文档
- **[系统架构概览](./AI小说系统-架构概览.md)** - 快速了解系统架构和模块功能
- **[完整实现文档](./AI小说系统-完整实现文档.md)** - 详细的模块设计和实现规范
- **[开发路线图](./AI小说系统-开发路线图.md)** - 功能规划和开发进度

### 专题文档
- **[7阶段演化系统实现报告](./7阶段演化系统实现报告.md)** - 世界构建器详细说明
- **[模块联动文档](./narrator-worldbuilder-linkage.md)** - 叙事器与世界构建器联动机制
- **[LLM配置调用链追踪](../CONFIG_CALL_CHAIN.md)** - LLM API 配置和调用追踪

---

## 🏗️ 系统架构

NovelFlow 采用模块化分层架构：

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (React)                       │
│  React 18 + TypeScript + Vite + TailwindCSS + shadcn/ui    │
└────────────────────────────┬────────────────────────────────┘
                             │ REST API
                             ↓
┌─────────────────────────────────────────────────────────────┐
│                       Backend (Go)                          │
│  Gin + GORM + PostgreSQL + Zhipu AI (GLM)                 │
└────────────────────────────┬────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        ↓                    ↓                    ↓
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ World Builder│    │  Narrative   │    │    Writer    │
│  世界构建器  │    │   Engine     │    │   写作器     │
└──────────────┘    └──────────────┘    └──────────────┘
```

### 核心模块

1. **调度器 (Scheduler)** - 系统入口，负责创建和管理编排器实例
2. **编排器 (Orchestrator)** - 单部小说的编排中心，协调创作流程
3. **世界构建器 (World Builder)** - 7阶段演化系统构建虚构世界
4. **叙事引擎 (Narrative Engine)** - 基于叙事理论规划情节结构
5. **写作器 (Writer)** - 场景化文本生成，维护状态一致性
6. **知识数据服务 (Knowledge Service)** - 提供写作规则和知识素材

---

## 🚀 快速开始

### 环境要求
- **Go**: 1.23+
- **Node.js**: 18+
- **PostgreSQL**: 14+
- **智谱AI API Key**: [https://open.bigmodel.cn/](https://open.bigmodel.cn/)

### 后端启动
```bash
# 克隆项目
git clone https://github.com/xlei/xupu.git
cd xupu

# 配置环境变量（创建 .env 文件）
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=novelflow
ZHIPU_API_KEY=your_zhipu_api_key

# 安装依赖并启动
go mod download
go run cmd/server/main.go
```

### 前端启动
```bash
cd web
npm install
npm run dev
```

访问：
- 前端: http://localhost:5173
- 后端: http://localhost:8080

---

## 📖 技术文档

### 后端技术栈
- **语言**: Go 1.23+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL
- **LLM**: 智谱AI GLM-4
- **配置管理**: Viper
- **CLI框架**: Cobra

### 前端技术栈
- **框架**: React 18.2
- **语言**: TypeScript 5.0
- **构建工具**: Vite 5.0
- **状态管理**: Zustand + React Query
- **路由**: React Router v6
- **UI组件**: shadcn/ui (Radix UI + TailwindCSS)
- **表单**: React Hook Form + Zod
- **HTTP客户端**: Axios

---

## 🎯 核心功能

### 世界构建
- 7阶段演化系统：哲学思考 → 世界观 → 法则设定 → 地理环境 → 文明社会 → 历史演化 → 一致性检查
- 支持多种世界类型：奇幻、科幻、历史、都市、武侠、仙侠、混合
- 灵活的世界规模：村庄、城市、国家、大陆、星球、宇宙

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

## 🛠️ 开发指南

### 后端开发
- 参考 [完整实现文档](./AI小说系统-完整实现文档.md) 了解模块设计
- 查看 [开发路线图](./AI小说系统-开发路线图.md) 了解开发进度
- 运行 `go run cmd/cli/main.go --help` 查看 CLI 命令

### 前端开发
- 参考 [前端开发指南](../web/README.md) 了解项目结构
- 查看 [快速启动指南](../web/QUICKSTART.md) 快速上手
- 使用 `npm run dev` 启动开发服务器
- 使用 `npm run format` 格式化代码

### 添加新功能
1. 在对应模块下创建代码文件
2. 更新相关文档
3. 运行测试确保功能正常
4. 提交 PR 并描述变更内容

---

## 📋 开发进度

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

## 🤝 贡献指南

欢迎贡献代码、报告问题或提出新功能建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📞 获取帮助

- **项目主页**: [https://github.com/xlei/xupu](https://github.com/xlei/xupu)
- **问题反馈**: [GitHub Issues](https://github.com/xlei/xupu/issues)
- **功能建议**: [GitHub Discussions](https://github.com/xlei/xupu/discussions)

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](../LICENSE) 文件

---

<div align="center">

**NovelFlow 叙谱 - 智能AI驱动的小说创作平台**

Made with ❤️ by NovelFlow Team

</div>
