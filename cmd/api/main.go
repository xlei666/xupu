// Package main Xupu AI小说创作系统 - API服务
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/api"
	"github.com/xlei/xupu/internal/handlers"
	"github.com/xlei/xupu/internal/middleware"
	"github.com/xlei/xupu/pkg/config"
	"github.com/xlei/xupu/pkg/db"
	"github.com/xlei/xupu/pkg/llm"
	"github.com/xlei/xupu/pkg/orchestrator"
	"github.com/xlei/xupu/pkg/worldbuilder"
)

// 静态文件将在运行时从文件系统加载
// TODO: 调整目录结构后使用embed
// //go:embed static
// var staticFiles embed.FS

func main() {
	// 初始化数据库（首次调用会自动初始化）
	_ = db.Get()

	// 初始化全局调度器
	if err := orchestrator.InitScheduler(); err != nil {
		log.Fatalf("Failed to initialize scheduler: %v", err)
	}
	defer orchestrator.StopScheduler()

	// 初始化编排器
	orc, err := orchestrator.New()
	if err != nil {
		log.Fatalf("Failed to initialize orchestrator: %v", err)
	}

	// 初始化配置
	cfg, err := config.LoadDefault()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化 LLM 客户端（用于叙事引擎）
	llmClient, _, err := llm.NewClientForModule("narrative_engine")
	if err != nil {
		log.Fatalf("Failed to initialize LLM client: %v", err)
	}

	// 初始化 WorldBuilder
	worldBuilder, err := worldbuilder.New()
	if err != nil {
		log.Fatalf("Failed to initialize world builder: %v", err)
	}

	// 创建服务器
	server := api.NewServer()

	// 注册中间件
	server.Use(middleware.Logger())
	server.Use(middleware.Recovery())
	server.Use(middleware.CORS())
	server.Use(middleware.RequestID())

	// 注册处理器
	projectHandler := handlers.NewProjectHandler(orc)
	worldHandler := handlers.NewWorldHandler(nil)
	narrativeHandler := handlers.NewNarrativeHandler(nil)
	exportHandler := handlers.NewExportHandler()
	authHandler := handlers.NewAuthHandler(getEnv("JWT_SECRET", "your-secret-key-change-in-production"))
	chapterHandler := handlers.NewChapterHandler()
	narrativeNodeHandler := handlers.NewNarrativeNodeHandler(db.Get(), llmClient, cfg)
	worldSettingHandler := handlers.NewWorldSettingHandler(db.Get(), worldBuilder)
	characterHandler := handlers.NewCharacterHandler(db.Get())
	synopsisHandler := handlers.NewSynopsisHandler(db.Get())
	writerHandler := handlers.NewWriterHandler(db.Get())
	externalRankHandler := handlers.NewExternalRankHandler()
	adminHandler := handlers.NewAdminHandler(db.Get())

	// 注册路由
	server.RegisterRoutes(projectHandler, worldHandler, narrativeHandler, exportHandler, authHandler, chapterHandler, narrativeNodeHandler, worldSettingHandler, characterHandler, synopsisHandler, writerHandler, externalRankHandler, adminHandler)

	// 配置静态文件服务（从文件系统加载，禁用JS缓存便于开发）
	server.Engine().Use(func(c *gin.Context) {
		if len(c.Request.URL.Path) > 3 && c.Request.URL.Path[len(c.Request.URL.Path)-3:] == ".js" {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})
	server.Engine().Static("/static", "./static")
	// Allow accessing the test page from root for convenience
	server.Engine().StaticFile("/fanqie_test.html", "./static/fanqie_test.html")

	// SPA路由支持 - 所有未匹配的路由返回index.html
	server.Engine().NoRoute(func(c *gin.Context) {
		// API路由返回404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		// 其他路由返回index.html（SPA）
		c.File("./static/index.html")
	})

	// 获取配置
	port := getEnv("PORT", "80")
	addr := ":" + port

	// 启动服务器
	log.Printf("Starting Xupu API server on %s", addr)
	log.Printf("Static files served from ./static")

	srv := &http.Server{
		Addr:    addr,
		Handler: server.Engine(),
	}

	// 启动goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// getEnv 获取环境变量，支持默认值
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
