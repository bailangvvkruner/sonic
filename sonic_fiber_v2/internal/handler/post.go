package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"sonic_fiber_v2/internal/service"
)

type PostHandler struct {
	*BaseHandler
	postService service.PostService
}

func NewPostHandler(
	logger *zap.Logger,
	postService service.PostService,
) *PostHandler {
	return &PostHandler{
		BaseHandler: NewBaseHandler(logger),
		postService: postService,
	}
}

// GetBySlug 根据别名获取文章
func (h *PostHandler) GetBySlug(c *fiber.Ctx) error {
	slug := h.GetParam(c, "slug")
	if slug == "" {
		return h.Error(c, 400, "文章别名不能为空")
	}

	post, err := h.postService.GetBySlug(c.Context(), slug)
	if err != nil {
		h.logger.Error("获取文章失败", zap.Error(err), zap.String("slug", slug))
		return h.Error(c, 404, "文章不存在")
	}

	return h.Success(c, post)
}

// GetByID 根据ID获取文章
func (h *PostHandler) GetByID(c *fiber.Ctx) error {
	idStr := h.GetParam(c, "id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		return h.Error(c, 400, "文章ID不能为空")
	}

	post, err := h.postService.GetByID(c.Context(), id)
	if err != nil {
		h.logger.Error("获取文章失败", zap.Error(err), zap.Int64("id", id))
		return h.Error(c, 404, "文章不存在")
	}

	return h.Success(c, post)
}

// ListAdmin 管理后台获取文章列表
func (h *PostHandler) ListAdmin(c *fiber.Ctx) error {
	page := h.GetIntQuery(c, "page", 1)
	size := h.GetIntQuery(c, "size", 20)
	status := h.GetQuery(c, "status")
	keyword := h.GetQuery(c, "keyword")

	posts, total, err := h.postService.ListAdmin(c.Context(), page, size, status, keyword)
	if err != nil {
		h.logger.Error("获取文章列表失败", zap.Error(err))
		return h.Error(c, 500, "获取文章列表失败")
	}

	return h.Success(c, fiber.Map{
		"list":  posts,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// Create 创建文章
func (h *PostHandler) Create(c *fiber.Ctx) error {
	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Slug     string `json:"slug"`
		Status   string `json:"status"`
		Category int64  `json:"category"`
		Tags     []int64 `json:"tags"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		return h.Error(c, 400, "参数错误")
	}

	post, err := h.postService.Create(c.Context(), req.Title, req.Content, req.Slug, req.Status, req.Category, req.Tags)
	if err != nil {
		h.logger.Error("创建文章失败", zap.Error(err))
		return h.Error(c, 500, "创建文章失败")
	}

	return h.Success(c, post)
}

// Update 更新文章
func (h *PostHandler) Update(c *fiber.Ctx) error {
	idStr := h.GetParam(c, "id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		return h.Error(c, 400, "文章ID不能为空")
	}

	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Slug     string `json:"slug"`
		Status   string `json:"status"`
		Category int64  `json:"category"`
		Tags     []int64 `json:"tags"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		return h.Error(c, 400, "参数错误")
	}

	post, err := h.postService.Update(c.Context(), id, req.Title, req.Content, req.Slug, req.Status, req.Category, req.Tags)
	if err != nil {
		h.logger.Error("更新文章失败", zap.Error(err), zap.Int64("id", id))
		return h.Error(c, 500, "更新文章失败")
	}

	return h.Success(c, post)
}

// Delete 删除文章
func (h *PostHandler) Delete(c *fiber.Ctx) error {
	idStr := h.GetParam(c, "id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		return h.Error(c, 400, "文章ID不能为空")
	}

	err := h.postService.Delete(c.Context(), id)
	if err != nil {
		h.logger.Error("删除文章失败", zap.Error(err), zap.Int64("id", id))
		return h.Error(c, 500, "删除文章失败")
	}

	return h.Success(c, nil)
}

// GetByArchive 根据归档获取文章
func (h *PostHandler) GetByArchive(c *fiber.Ctx) error {
	slug := h.GetParam(c, "slug")
	if slug == "" {
		return h.Error(c, 400, "归档别名不能为空")
	}

	posts, err := h.postService.GetByArchive(c.Context(), slug)
	if err != nil {
		h.logger.Error("获取归档文章失败", zap.Error(err), zap.String("slug", slug))
		return h.Error(c, 500, "获取归档文章失败")
	}

	return h.Success(c, posts)
}
