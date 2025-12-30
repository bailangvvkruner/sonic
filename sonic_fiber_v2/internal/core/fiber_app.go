package core

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"

	"sonic_fiber_v2/config"
	"sonic_fiber_v2/internal/middleware"
)

// NewFiberApp 创建Fiber应用实例
func NewFiberApp(cfg *config.Config, log *zap.Logger) *fiber.App {
	// 创建缓存中间件实例
	cacheMiddleware := middleware.NewCacheMiddleware(log)
	// 配置Fiber
	fiberConfig := fiber.Config{
		DisableStartupMessage: true,
		// 优化性能
		Concurrency:        256 * 1024,
		DisableKeepalive:   false,
		ReadTimeout:        30 * 1000 * 1000 * 1000, // 30秒
		WriteTimeout:       30 * 1000 * 1000 * 1000, // 30秒
		IdleTimeout:        30 * 1000 * 1000 * 1000, // 30秒
		ReduceMemoryUsage:  true,
	}

	app := fiber.New(fiberConfig)

	// 全局中间件
	app.Use(recover.New())

	// 日志中间件
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${method} | ${path} | ${latency}\n",
		DisableColors: !cfg.IsDev(),
	}))

	// CORS中间件（开发环境）
	if cfg.IsDev() {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     "*",
			AllowMethods:     "GET, POST, PUT, DELETE, PATCH, OPTIONS",
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
			AllowCredentials: true,
		}))
	}

	// 缓存中间件
	app.Use(cacheMiddleware.Handle)

	// 健康检查路由
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "sonic-fiber-v2",
		})
	})

	// Ping测试
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	return app
}
