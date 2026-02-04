// Package api API服务器
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/handlers"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/orchestrator"
)

// Server API服务器
type Server struct {
	engine       *gin.Engine
	orchestrator *orchestrator.Orchestrator
}

// NewServer 创建API服务器
func NewServer() *Server {
	// 初始化数据库
	_ = db.Get()

	// 创建编排器（内部已初始化数据库）
	orc, _ := orchestrator.New()

	// 创建Gin引擎
	engine := gin.New()

	return &Server{
		engine:       engine,
		orchestrator: orc,
	}
}

// Use 注册中间件
func (s *Server) Use(middleware ...gin.HandlerFunc) {
	s.engine.Use(middleware...)
}

// RegisterRoutes 注册路由
func (s *Server) RegisterRoutes(
	projectHandler *handlers.ProjectHandler,
	worldHandler *handlers.WorldHandler,
	narrativeHandler *handlers.NarrativeHandler,
	exportHandler *handlers.ExportHandler,
	authHandler *handlers.AuthHandler,
	chapterHandler *handlers.ChapterHandler,
	narrativeNodeHandler *handlers.NarrativeNodeHandler,
	worldSettingHandler *handlers.WorldSettingHandler,
	characterHandler *handlers.CharacterHandler,
	synopsisHandler *handlers.SynopsisHandler,
	writerHandler *handlers.WriterHandler,
	externalRankHandler *handlers.ExternalRankHandler,
	adminHandler *handlers.AdminHandler,
) {
	// 同时创建任务处理器
	taskHandler := handlers.NewTaskHandler()

	// 健康检查
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "xupu-api",
		})
	})

	// API v1
	v1 := s.engine.Group("/api/v1")
	{
		// 认证路由（无需认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// 用户路由（需要认证）
		users := v1.Group("/users")
		{
			users.GET("/me", authHandler.GetCurrentUser)
			users.PUT("/me/password", authHandler.ChangePassword)
		}

		// 项目管理（需要认证）
		projects := v1.Group("/projects")
		projects.Use(authHandler.AuthMiddleware()) // 应用认证中间件
		{
			projects.POST("", projectHandler.CreateProject)
			projects.POST("/import", projectHandler.ImportProject)
			projects.GET("", projectHandler.ListProjects)
			projects.GET("/:projectId", projectHandler.GetProject)
			projects.DELETE("/:projectId", projectHandler.DeleteProject)
			projects.POST("/:projectId/generate", projectHandler.GenerateChapter)
			projects.POST("/:projectId/intervene", projectHandler.Intervene)
			projects.POST("/:projectId/pause", projectHandler.PauseGeneration)
			projects.POST("/:projectId/resume", projectHandler.ResumeGeneration)
			projects.GET("/:projectId/progress", projectHandler.GetProgress)
			projects.POST("/:projectId/blueprint/apply", narrativeHandler.ApplyBlueprint)

			// 章节管理（使用 :projectId 作为项目ID）
			projects.GET("/:projectId/chapters", chapterHandler.ListChapters)
			projects.GET("/:projectId/chapters/:chapterId", chapterHandler.GetChapter)
			projects.POST("/:projectId/chapters", chapterHandler.CreateChapter)
			projects.PUT("/:projectId/chapters/:chapterId", chapterHandler.UpdateChapter)
			projects.DELETE("/:projectId/chapters/:chapterId", chapterHandler.DeleteChapter)
			projects.POST("/:projectId/chapters/:chapterId/continue", writerHandler.ContinueChapter)
			projects.GET("/:projectId/chapters/:chapterId/outline", writerHandler.GenerateChapterOutline)

			// 叙事节点管理
			projects.GET("/:projectId/narrative-nodes", narrativeNodeHandler.GetNodeTree)
			projects.POST("/:projectId/narrative-nodes", narrativeNodeHandler.CreateNode)
			projects.PUT("/:projectId/narrative-nodes/:nodeId", narrativeNodeHandler.UpdateNode)
			projects.DELETE("/:projectId/narrative-nodes/:nodeId", narrativeNodeHandler.DeleteNode)
			projects.POST("/:projectId/narrative-nodes/:nodeId/branches", narrativeNodeHandler.GenerateBranches)
			projects.POST("/:projectId/narrative-nodes/:nodeId/branches/:branchId/select", narrativeNodeHandler.SelectBranch)
			projects.POST("/:projectId/narrative-nodes/:nodeId/merge", narrativeNodeHandler.MergeToChapter)

			// 世界设定管理（7个阶段）
			projects.GET("/:projectId/world-stages", worldSettingHandler.GetWorldStages)
			projects.POST("/:projectId/world-stages", worldSettingHandler.SaveWorldStages)
			projects.POST("/:projectId/world-stages/:stage/generate", worldSettingHandler.GenerateWorldStage)

			// 使用 world-gacha 避免与 :stage 路由冲突
			projects.POST("/:projectId/world-gacha", worldSettingHandler.GachaWorldSettings)

			// 角色设定管理
			projects.POST("/:projectId/characters/gacha", characterHandler.GachaCharacters)

			// 简介设定管理
			projects.POST("/:projectId/synopsis/gacha", synopsisHandler.GachaSynopsis)
		}

		// 世界设定
		worlds := v1.Group("/worlds")
		{
			worlds.POST("", worldHandler.CreateWorld)
			worlds.GET("", worldHandler.ListWorlds)
			worlds.GET("/:id", worldHandler.GetWorld)
			worlds.DELETE("/:id", worldHandler.DeleteWorld)
		}

		// 叙事蓝图
		blueprints := v1.Group("/blueprints")
		{
			blueprints.POST("", narrativeHandler.CreateBlueprint)
			blueprints.GET("/:id", narrativeHandler.GetBlueprint)
			blueprints.GET("/:id/export", narrativeHandler.ExportBlueprint)
		}

		// 导出
		export := v1.Group("/export")
		{
			export.GET("/project/:id", exportHandler.ExportProject)
			export.GET("/world/:id", exportHandler.ExportWorld)
			export.GET("/blueprint/:id", exportHandler.ExportBlueprint)
		}

		// 异步任务
		tasks := v1.Group("/tasks")
		{
			tasks.POST("/project", taskHandler.CreateAsyncProject)
			tasks.GET("/:id", taskHandler.GetTaskStatus)
			tasks.POST("/:id/cancel", taskHandler.CancelTask)
			tasks.POST("/:id/pause", taskHandler.PauseTask)
			tasks.GET("/stats", taskHandler.GetSchedulerStats)
			tasks.GET("/project/:id", taskHandler.ListProjectTasks)
			tasks.GET("/:id/wait", taskHandler.WaitForTask)
		}

		// 外部数据源
		external := v1.Group("/external")
		{
			// 排行榜
			external.GET("/ranks/fanqie", externalRankHandler.GetFanqieRank)

			// 番茄小说详细API
			fanqie := external.Group("/fanqie")
			{
				fanqie.GET("/books/:bookId", externalRankHandler.GetFanqieBookDetail)
				fanqie.GET("/books/:bookId/chapters", externalRankHandler.GetFanqieChapterList)
				fanqie.GET("/chapters/:chapterId", externalRankHandler.GetFanqieChapterContent)
			}
		}

		// 管理后台
		admin := v1.Group("/admin")
		// admin.Use(authHandler.AdminMiddleware()) // TODO: 添加管理员认证
		{
			// 系统配置
			admin.GET("/configs", adminHandler.GetConfigs)
			admin.PUT("/configs/:key", adminHandler.UpdateConfig)

			// 提示词管理
			admin.GET("/prompts", adminHandler.GetPrompts)
			admin.GET("/prompts/:key", adminHandler.GetPrompt)
			admin.PUT("/prompts/:key", adminHandler.UpdatePrompt)
			admin.POST("/sync", adminHandler.SyncFromConfig)

			// 结构模板
			admin.GET("/structures", adminHandler.GetStructures)
			admin.PUT("/structures/:id", adminHandler.UpdateStructure)
		}
	}
}

// Engine 获取Gin引擎
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
