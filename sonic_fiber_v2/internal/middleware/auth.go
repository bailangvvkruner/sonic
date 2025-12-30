package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"sonic_fiber_v2/internal/service"
)

type AuthMiddleware struct {
	userService service.UserService
	logger      *zap.Logger
}

func NewAuthMiddleware(logger *zap.Logger) *AuthMiddleware {
	// 简化实现，暂时不依赖UserService
	return &AuthMiddleware{
		userService: nil,
		logger:      logger,
	}
}

// Handle 认证中间件
func (m *AuthMiddleware) Handle(c *fiber.Ctx) error {
	// 从请求头获取token
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "未授权",
			"code":    401,
		})
	}

	// 简化实现，直接返回成功
	// 在实际项目中，这里应该验证JWT token
	return c.Next()
}

// SkipAuth 跳过认证的中间件（用于测试）
func SkipAuth(c *fiber.Ctx) error {
	return c.Next()
}
