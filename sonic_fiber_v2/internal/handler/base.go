package handler

import (
	"fmt"
	
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// BaseHandler 基础处理器，提供通用功能
type BaseHandler struct {
	logger *zap.Logger
}

func NewBaseHandler(logger *zap.Logger) *BaseHandler {
	return &BaseHandler{
		logger: logger,
	}
}

// Success 成功响应
func (h *BaseHandler) Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// Error 错误响应
func (h *BaseHandler) Error(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(fiber.Map{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}

// BindJSON 绑定JSON请求体
func (h *BaseHandler) BindJSON(c *fiber.Ctx, v interface{}) error {
	if err := c.BodyParser(v); err != nil {
		h.logger.Error("解析请求体失败", zap.Error(err))
		return err
	}
	return nil
}

// GetParam 获取路径参数
func (h *BaseHandler) GetParam(c *fiber.Ctx, key string) string {
	return c.Params(key)
}

// GetQuery 获取查询参数
func (h *BaseHandler) GetQuery(c *fiber.Ctx, key string) string {
	return c.Query(key)
}

// GetIntQuery 获取整型查询参数
func (h *BaseHandler) GetIntQuery(c *fiber.Ctx, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
