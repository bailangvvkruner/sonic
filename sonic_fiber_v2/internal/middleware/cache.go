package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type CacheMiddleware struct {
	logger *zap.Logger
}

func NewCacheMiddleware(logger *zap.Logger) *CacheMiddleware {
	return &CacheMiddleware{
		logger: logger,
	}
}

// Handle 缓存中间件
func (m *CacheMiddleware) Handle(c *fiber.Ctx) error {
	// 对静态资源设置缓存头
	path := c.Path()
	if m.isStaticResource(path) {
		c.Set("Cache-Control", "public, max-age=3600") // 1小时
		c.Set("Expires", time.Now().Add(time.Hour*1).Format(time.RFC1123))
	}

	return c.Next()
}

func (m *CacheMiddleware) isStaticResource(path string) bool {
	staticExtensions := []string{".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot"}
	for _, ext := range staticExtensions {
		if len(path) >= len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}
	return false
}
