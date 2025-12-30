package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"sonic_fiber_v2/internal/service"
)

type UserHandler struct {
	*BaseHandler
	userService service.UserService
}

func NewUserHandler(
	logger *zap.Logger,
	userService service.UserService,
) *UserHandler {
	return &UserHandler{
		BaseHandler: NewBaseHandler(logger),
		userService: userService,
	}
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	// 这里应该从token中获取用户ID
	userID := int64(1) // 简化实现

	user, err := h.userService.GetByID(c.Context(), userID)
	if err != nil {
		h.logger.Error("获取用户资料失败", zap.Error(err))
		return h.Error(c, 500, "获取用户资料失败")
	}

	return h.Success(c, user)
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Avatar   string `json:"avatar"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		return h.Error(c, 400, "参数错误")
	}

	// 简化实现，使用固定ID
	userID := int64(1)
	user, err := h.userService.UpdateProfile(c.Context(), userID, req.Nickname, req.Email, req.Avatar)
	if err != nil {
		h.logger.Error("更新用户资料失败", zap.Error(err))
		return h.Error(c, 500, "更新用户资料失败")
	}

	return h.Success(c, user)
}

// UpdatePassword 更新密码
func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		return h.Error(c, 400, "参数错误")
	}

	// 简化实现，使用固定ID
	userID := int64(1)
	err := h.userService.UpdatePassword(c.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		h.logger.Error("更新密码失败", zap.Error(err))
		return h.Error(c, 500, "更新密码失败")
	}

	return h.Success(c, nil)
}
