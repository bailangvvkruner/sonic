package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"sonic_fiber_v2/internal/service"
)

type CommentHandler struct {
	*BaseHandler
	commentService service.CommentService
}

func NewCommentHandler(
	logger *zap.Logger,
	commentService service.CommentService,
) *CommentHandler {
	return &CommentHandler{
		BaseHandler:    NewBaseHandler(logger),
		commentService: commentService,
	}
}

// Create 创建评论
func (h *CommentHandler) Create(c *fiber.Ctx) error {
	var req struct {
		PostID   int64  `json:"post_id"`
		Content  string `json:"content"`
		Author   string `json:"author"`
		Email    string `json:"email"`
		ParentID int64  `json:"parent_id"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		return h.Error(c, 400, "参数错误")
	}

	comment, err := h.commentService.Create(c.Context(), req.PostID, req.Content, req.Author, req.Email, req.ParentID)
	if err != nil {
		h.logger.Error("创建评论失败", zap.Error(err))
		return h.Error(c, 500, "创建评论失败")
	}

	return h.Success(c, comment)
}

// ListAdmin 管理后台获取评论列表
func (h *CommentHandler) ListAdmin(c *fiber.Ctx) error {
	page := h.GetIntQuery(c, "page", 1)
	size := h.GetIntQuery(c, "size", 20)
	postID := h.GetIntQuery(c, "post_id", 0)

	comments, total, err := h.commentService.ListAdmin(c.Context(), page, size, int64(postID))
	if err != nil {
		h.logger.Error("获取评论列表失败", zap.Error(err))
		return h.Error(c, 500, "获取评论列表失败")
	}

	return h.Success(c, fiber.Map{
		"list":  comments,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// UpdateStatus 更新评论状态
func (h *CommentHandler) UpdateStatus(c *fiber.Ctx) error {
	idStr := h.GetParam(c, "id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		return h.Error(c, 400, "评论ID不能为空")
	}

	status := h.GetParam(c, "status")
	if status == "" {
		return h.Error(c, 400, "状态不能为空")
	}

	err := h.commentService.UpdateStatus(c.Context(), id, status)
	if err != nil {
		h.logger.Error("更新评论状态失败", zap.Error(err), zap.Int64("id", id))
		return h.Error(c, 500, "更新评论状态失败")
	}

	return h.Success(c, nil)
}

// Delete 删除评论
func (h *CommentHandler) Delete(c *fiber.Ctx) error {
	idStr := h.GetParam(c, "id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		return h.Error(c, 400, "评论ID不能为空")
	}

	err := h.commentService.Delete(c.Context(), id)
	if err != nil {
		h.logger.Error("删除评论失败", zap.Error(err), zap.Int64("id", id))
		return h.Error(c, 500, "删除评论失败")
	}

	return h.Success(c, nil)
}
