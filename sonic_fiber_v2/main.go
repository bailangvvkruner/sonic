package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"sonic_fiber_v2/config"
	"sonic_fiber_v2/internal/core"
	"sonic_fiber_v2/internal/handler"
	"sonic_fiber_v2/internal/middleware"
	"sonic_fiber_v2/internal/service"
	"sonic_fiber_v2/pkg/database"
	"sonic_fiber_v2/pkg/log"
)

func main() {
	// 创建FX应用
	app := fx.New(
		fx.Provide(
			// 配置
			config.NewConfig,
			
			// 日志
			log.NewLogger,
			
			// 数据库
			database.NewGormDB,
			
			// 核心服务
			core.NewFiberApp,
			
			// 中间件
			middleware.NewAuthMiddleware,
			
			// 业务服务 - 使用优化后的无锁服务
			service.NewOptionService,
			func() service.PostService {
				svc := service.NewOptimizedPostService().(*service.OptimizedPostService)
				// 启动时从SQLite加载数据
				go func() {
					if err := svc.LoadFromDB(); err != nil {
						// 记录日志但不影响启动
						zap.L().Error("Failed to load posts from DB", zap.Error(err))
					} else {
						zap.L().Info("Posts loaded into memory successfully")
					}
					// 启动异步写入队列
					svc.StartWriteQueue()
				}()
				return svc
			},
			service.NewCategoryService,
			service.NewTagService,
			service.NewCommentService,
			service.NewUserService,
			service.NewThemeService,
			func(cfg *config.Config) service.AttachmentService {
				return service.NewAttachmentService(cfg.App.UploadDir)
			},
			
			// 处理器
			handler.NewIndexHandler,
			handler.NewPostHandler,
			handler.NewCategoryHandler,
			handler.NewTagHandler,
			handler.NewCommentHandler,
			handler.NewUserHandler,
			handler.NewAdminHandler,
			handler.NewAttachmentHandler,
			
			// 路由注册器
			core.NewRouteRegister,
		),
		fx.Invoke(
			// 注册路由
			core.InvokeRouteRegister,
			// 启动服务
			startServer,
		),
	)

	// 启动应用
	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}

	// 等待退出信号
	waitForShutdown()

	// 停止应用
	if err := app.Stop(context.Background()); err != nil {
		panic(err)
	}
}

func startServer(lifecycle fx.Lifecycle, app *fiber.App, logger *zap.Logger, cfg *config.Config) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			addr := cfg.Server.Host + ":" + cfg.Server.Port
			go func() {
				if err := app.Listen(addr); err != nil {
					logger.Error("Server failed to start", zap.Error(err))
					os.Exit(1)
				}
			}()
			logger.Info("Server started", zap.String("addr", addr))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down server...")
			return app.Shutdown()
		},
	})
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}
