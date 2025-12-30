package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"sonic_fiber_v2/internal/service"
)

type TagHandler struct {
	*BaseHandler
	tagService service.TagService
}

func NewTagHandler(
	logger *zap.Logger,
	tagService service.TagService,
) *TagHandler {
	return &TagHandler{
		BaseHandler: NewBaseHandler(logger),
		tagService:  tagService,
	}
}

// List 获取标签列表
func (h *TagHandler) List(c *fiber.Ctx) error {
	tags, err := h.tagService.List(c.Context())
	if err != nil {
		h.logger.Error("获取标签列表失败", zap.Error(err))
		return h.Error(c, 500, "获取标签列表失败")
	}

	return h.Success(c, tags)
}

// GetBySlug 根据别名获取标签
func (h *TagHandler) GetBySlug(c *fiber.Ctx) error {
	slug := h.GetParam(c, "slug")
	if slug == "" {
		return h.Error(c, 400, "标签别名不能为空")
	}

	tag, err := h.tagService.GetBySlug(c.Context(), slug)
	if err != nil {
		h.logger.Error("获取标签失败", zap.Error(err), zap.String("slug", slug))
		return h.Error(c, 404, "标签不存在")
	}

	return h.Success(c, tag)
}

// GetPosts 获取标签下的文章
func (h *TagHandler) GetPosts(c *fiber.Ctx) error {
	slug := h.GetParam(c, "slug")
	if slug == "" {
		return h.Error(c, 400, "标签别名不能为空")
	}

	page := h.GetIntQuery(c, "page", 1)
	size := h.GetIntQuery(c, "size", 10)

	posts, err := h.tagService.GetPosts(c.Context(), slug, page, size)
	if err != nil {
		h.logger.Error("获取标签文章失败", zap.Error(err), zap.String("slug", slug))
		return h.Error(c, 500, "获取标签文章失败")
	}

	return h.Success(c, posts)
}

// ListAdmin 管理后台获取标签列表
func (h *TagHandler) ListAdmin(c *fiber.Ctx) error {
	tags, err := h.tagService.ListAdmin(c.Context())
	if err != nil {
		h.logger.Error("获取管理标签列表失败", zap.Error(err))
		return h.Error(c, 500, "获取标签列表失败")
	}

	return h.Success(c, tags)
}

// Create 创建标签
func (h *TagHandler) Create(c *fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		return h.Error(c, 400, "参数错误")
	}

	tag, err := h.tagService.Create(c.Context(), req.Name, req.Slug)
	if err != nil {
		h.logger.Error("创建标签失败", zap.Error(err))
		return h.Error(c, 500, "创建标签失败")
	}

	return h.Success(c, tag)
}

// Update 更新标签
func (h *TagHandler) Update(c *fiber.Ctx) error {
	idStr := h.GetParam(c, "id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		return h.Error(c, 400, "标签ID不能为空")
	}

	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		return h.Error(c, 400, "参数错误")
	}

	tag, err := h.tagService.Update(c.Context(), id, req.Name, req.Slug)
	if err != nil {
		h.logger.Error("更新标签失败", zap.Error(err), zap.Int64("id", id))
		return h.Error(c, 500, "更新标签失败")
	}

	return h.Success(c, tag)
}

// Delete 删除标签
func (h *TagHandler) Delete(c *fiber.Ctx) error {
	idStr := h.GetParam(c, "id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		return h.Error(c, 400, "标签ID不能为空")
	}

	err := h.tagService.Delete(c.Context(), id)
	if err != nil {
		h.logger.Error("删除标签失败", zap.Error(err), zap.Int64("id", id))
		return h.Error(c, 500, "删除标签失败")
	}

	return h.Success(c, nil)
}
