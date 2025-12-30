package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"sonic_fiber_v2/internal/service"
)

type AdminHandler struct {
	*BaseHandler
	userService service.UserService
}

func NewAdminHandler(
	logger *zap.Logger,
	userService service.UserService,
) *AdminHandler {
	return &AdminHandler{
		BaseHandler: NewBaseHandler(logger),
		userService: userService,
	}
}

// IsInstalled 检查是否已安装
func (h *AdminHandler) IsInstalled(c *fiber.Ctx) error {
	// 简化实现，返回true
	return h.Success(c, fiber.Map{
		"installed": true,
	})
}

// Login 登录
func (h *AdminHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		return h.Error(c, 400, "参数错误")
	}

	// 简化实现，验证用户名密码
	token, err := h.userService.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Error("登录失败", zap.Error(err))
		return h.Error(c, 401, "登录失败")
	}

	return h.Success(c, fiber.Map{
		"token": token,
	})
}

// Install 安装博客
func (h *AdminHandler) Install(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		SiteName string `json:"site_name"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		return h.Error(c, 400, "参数错误")
	}

	// 简化实现
	err := h.userService.Install(c.Context(), req.Username, req.Password, req.Email, req.SiteName)
	if err != nil {
		h.logger.Error("安装失败", zap.Error(err))
		return h.Error(c, 500, "安装失败")
	}

	return h.Success(c, nil)
}
