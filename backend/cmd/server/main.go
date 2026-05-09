// main.go - Claw Native Company 后端服务入口
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"claw/internal/config"
	"claw/internal/database"
	"claw/internal/handler"
	"claw/internal/logger"
	"claw/internal/middleware"
	"claw/internal/repository"
	"claw/internal/service"
	"claw/internal/websocket"
	"claw/migrations"
	"claw/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfgPath := os.Getenv("CLAW_CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

	if err := config.Init(cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}
	cfg := config.Get()

	// 确保数据目录存在
	if err := cfg.EnsureDataDir(); err != nil {
		fmt.Fprintf(os.Stderr, "创建数据目录失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		fmt.Fprintf(os.Stderr, "日志初始化失败: %v\n", err)
		os.Exit(1)
	}

	log := logger.Get()
	log.Info("服务启动中",
		"version", "1.0.0",
		"mode", cfg.Server.Mode,
		"port", cfg.Server.Port,
	)

	// 初始化数据库
	if err := database.Init(cfg); err != nil {
		log.Error("数据库初始化失败", "error", err)
		os.Exit(1)
	}

	// 执行数据库迁移
	if err := migrations.Migrate(database.GetDB()); err != nil {
		log.Error("数据库迁移失败", "error", err)
		os.Exit(1)
	}

	// 获取数据库实例
	db := database.GetDB()

	// 初始化 WebSocket 管理器
	websocket.Init()

	// 设置 Gin 模式
	if cfg.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Repository 实例
	employeeRepo := repository.NewEmployeeRepository()
	taskRepo := repository.NewTaskRepository()
	channelRepo := repository.NewChannelRepository()
	workflowRepo := repository.NewWorkflowRepository()
	gatewayConfigRepo := repository.NewGatewayConfigRepository(db)

	// 创建 Service 实例
	dashboardService := service.NewDashboardService(employeeRepo, taskRepo, channelRepo, workflowRepo)
	gatewayConfigService := service.NewGatewayConfigService(gatewayConfigRepo)

	// 创建 Handler 实例
	authHandler := handler.NewAuthHandler()
	employeeHandler := handler.NewEmployeeHandler()
	channelHandler := handler.NewChannelHandler()
	messageHandler := handler.NewMessageHandler()
	workflowHandler := handler.NewWorkflowHandler()
	taskHandler := handler.NewTaskHandler()
	agentHandler := handler.NewAgentHandler()
	webhookHandler := handler.NewWebhookHandler()
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	gatewayConfigHandler := handler.NewGatewayConfigHandler(gatewayConfigService)
	documentHandler := handler.NewDocumentHandler() // 新增文档Handler
	wsManager := websocket.GetManager()

	// 创建 Gin 引擎
	r := gin.New()

	// 注册中间件（注意顺序）
	r.Use(middleware.CORS())         // 跨域
	r.Use(middleware.ErrorHandler()) // 错误处理
	r.Use(gin.Recovery())            // Gin 内置恢复

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 健康检查
		handler.RegisterHealthRoutes(api)

		// 认证路由（无需认证）
		authHandler.RegisterRoutes(api.Group("/auth"))

		// 需要认证的路由
		protected := api.Group("")
		protected.Use(middleware.DualAuth())
		{
			// 员工管理路由
			employeeHandler.RegisterRoutes(protected)

			// 频道管理路由
			channelHandler.RegisterRoutes(protected)

			// 文档管理路由
			documentHandler.RegisterRoutes(protected)

			// 消息路由
			messageHandler.RegisterRoutes(protected)

			// 工作流路由
			workflowHandler.RegisterRoutes(protected)

			// 任务路由
			taskHandler.RegisterRoutes(protected)

			// Gateway 配置路由
			gatewayConfigHandler.RegisterRoutes(protected)

			// Dashboard 路由
			protected.GET("/dashboard/stats", dashboardHandler.GetStats)
		}

		// Agent 路由（使用 API Key 认证）
		agentRoutes := api.Group("")
		agentRoutes.Use(middleware.APIKeyAuth())
		{
			agentHandler.RegisterRoutes(agentRoutes)
		}
	}

	// Webhook 路由（无需认证）
	webhookHandler.RegisterRoutes(r.Group(""))

	// WebSocket 路由（认证在 Handler 内部处理）
	r.GET("/ws", wsManager.HandleWebSocket)

	// 404 处理
	r.NoRoute(func(c *gin.Context) {
		utils.Error(c, http.StatusNotFound, "请求的资源不存在")
	})
	r.NoMethod(func(c *gin.Context) {
		utils.Error(c, http.StatusMethodNotAllowed, "请求方法不允许")
	})

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		log.Info("收到关闭信号", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("服务关闭失败", "error", err)
		}
	}()

	// 启动服务
	log.Info("服务启动成功", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("服务启动失败", "error", err)
		os.Exit(1)
	}

	log.Info("服务已停止")

	// 关闭数据库
	if err := database.Close(); err != nil {
		log.Error("数据库关闭失败", "error", err)
	}
}
