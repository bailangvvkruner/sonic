package handler

import (
	"fmt"
	
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"sonic_fiber_v2/internal/service"
)

type IndexHandler struct {
	*BaseHandler
	postService service.PostService
}

func NewIndexHandler(
	logger *zap.Logger,
	postService service.PostService,
) *IndexHandler {
	return &IndexHandler{
		BaseHandler: NewBaseHandler(logger),
		postService: postService,
	}
}

// Home 首页
func (h *IndexHandler) Home(c *fiber.Ctx) error {
	page := h.GetIntQuery(c, "page", 1)
	size := h.GetIntQuery(c, "size", 10)

	posts, err := h.postService.GetRecentPosts(c.Context(), page, size)
	if err != nil {
		h.logger.Error("获取文章列表失败", zap.Error(err))
		return h.Error(c, 500, "获取文章列表失败")
	}

	return h.Success(c, posts)
}

// Page 分页首页
func (h *IndexHandler) Page(c *fiber.Ctx) error {
	pageStr := h.GetParam(c, "page")
	page := 1
	if pageStr != "" {
		fmt.Sscanf(pageStr, "%d", &page)
	}

	posts, err := h.postService.GetRecentPosts(c.Context(), page, 10)
	if err != nil {
		h.logger.Error("获取分页文章失败", zap.Error(err))
		return h.Error(c, 500, "获取分页文章失败")
	}

	return h.Success(c, posts)
}

// Search 搜索
func (h *IndexHandler) Search(c *fiber.Ctx) error {
	keyword := h.GetQuery(c, "q")
	if keyword == "" {
		return h.Error(c, 400, "搜索关键词不能为空")
	}

	page := h.GetIntQuery(c, "page", 1)
	size := h.GetIntQuery(c, "size", 10)

	results, err := h.postService.Search(c.Context(), keyword, page, size)
	if err != nil {
		h.logger.Error("搜索失败", zap.Error(err))
		return h.Error(c, 500, "搜索失败")
	}

	return h.Success(c, results)
}

// Archives 归档
func (h *IndexHandler) Archives(c *fiber.Ctx) error {
	archives, err := h.postService.GetArchives(c.Context())
	if err != nil {
		h.logger.Error("获取归档失败", zap.Error(err))
		return h.Error(c, 500, "获取归档失败")
	}

	return h.Success(c, archives)
}
