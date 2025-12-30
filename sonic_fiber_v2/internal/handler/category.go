package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"sonic_fiber_v2/internal/service"
)

type CategoryHandler struct {
	*BaseHandler
	categoryService service.CategoryService
}

func NewCategoryHandler(
	logger *zap.Logger,
	categoryService service.CategoryService,
) *CategoryHandler {
	return &CategoryHandler{
		BaseHandler:     NewBaseHandler(logger),
		categoryService: categoryService,
	}
}

// List 获取分类列表
func (h *CategoryHandler) List(c *fiber.Ctx) error {
	categories, err := h.categoryService.List(c.Context())
	if err != nil {
		h.logger.Error("获取分类列表失败", zap.Error(err))
		return h.Error(c, 500, "获取分类列表失败")
	}

	return h.Success(c, categories)
}

// GetBySlug 根据别名获取分类
func (h *CategoryHandler) GetBySlug(c *fiber.Ctx) error {
	slug := h.GetParam(c, "slug")
	if slug == "" {
		return h.Error(c, 400, "分类别名不能为空")
	}

	category, err := h.categoryService.GetBySlug(c.Context(), slug)
	if err != nil {
		h.logger.Error("获取分类失败", zap.Error(err), zap.String("slug", slug))
		return h.Error(c, 404, "分类不存在")
	}

	return h.Success(c, category)
}

// GetPosts 获取分类下的文章
func (h *CategoryHandler) GetPosts(c *fiber.Ctx) error {
	slug := h.GetParam(c, "slug")
	if slug == "" {
		return h.Error(c, 400, "分类别名不能为空")
	}

	page := h.GetIntQuery(c, "page", 1)
	size := h.GetIntQuery(c, "size", 10)

	posts, err := h.categoryService.GetPosts(c.Context(), slug, page, size)
	if err != nil {
		h.logger.Error("获取分类文章失败", zap.Error(err), zap.String("slug", slug))
		return h.Error(c, 500, "获取分类文章失败")
	}

	return h.Success(c, posts)
}

// ListAdmin 管理后台获取分类列表
func (h *CategoryHandler) ListAdmin(c *fiber.Ctx) error {
	categories, err := h.categoryService.ListAdmin(c.Context())
	if err != nil {
		h.logger.Error("获取管理分类列表失败", zap.Error(err))
		return h.Error(c, 500, "获取分类列表失败")
	}

	return h.Success(c, categories)
}

// Create 创建分类
func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		ParentID    int64  `json:"parent_id"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		return h.Error(c, 400, "参数错误")
	}

	category, err := h.categoryService.Create(c.Context(), req.Name, req.Slug, req.Description, req.ParentID)
	if err != nil {
		h.logger.Error("创建分类失败", zap.Error(err))
		return h.Error(c, 500, "创建分类失败")
	}

	return h.Success(c, category)
}

// Update 更新分类
func (h *CategoryHandler) Update(c *fiber.Ctx) error {
	idStr := h.GetParam(c, "id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		return h.Error(c, 400, "分类ID不能为空")
	}

	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		ParentID    int64  `json:"parent_id"`
	}

	if err := h.BindJSON(c, &req); err != nil {
		return h.Error(c, 400, "参数错误")
	}

	category, err := h.categoryService.Update(c.Context(), id, req.Name, req.Slug, req.Description, req.ParentID)
	if err != nil {
		h.logger.Error("更新分类失败", zap.Error(err), zap.Int64("id", id))
		return h.Error(c, 500, "更新分类失败")
	}

	return h.Success(c, category)
}

// Delete 删除分类
func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	idStr := h.GetParam(c, "id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		return h.Error(c, 400, "分类ID不能为空")
	}

	err := h.categoryService.Delete(c.Context(), id)
	if err != nil {
		h.logger.Error("删除分类失败", zap.Error(err), zap.Int64("id", id))
		return h.Error(c, 500, "删除分类失败")
	}

	return h.Success(c, nil)
}
