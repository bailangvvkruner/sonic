package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"sonic_fiber_v2/internal/service"
)

type AttachmentHandler struct {
	*BaseHandler
	attachmentService service.AttachmentService
}

func NewAttachmentHandler(
	logger *zap.Logger,
	attachmentService service.AttachmentService,
) *AttachmentHandler {
	return &AttachmentHandler{
		BaseHandler:       NewBaseHandler(logger),
		attachmentService: attachmentService,
	}
}

// Upload 上传单个文件
func (h *AttachmentHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("获取文件失败", zap.Error(err))
		return h.Error(c, 400, "文件上传失败")
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		h.logger.Error("打开文件失败", zap.Error(err))
		return h.Error(c, 500, "文件上传失败")
	}
	defer src.Close()

	attachment, err := h.attachmentService.Upload(c.Context(), file.Filename, src, file.Size)
	if err != nil {
		h.logger.Error("上传文件失败", zap.Error(err))
		return h.Error(c, 500, "文件上传失败")
	}

	return h.Success(c, attachment)
}

// List 获取附件列表
func (h *AttachmentHandler) List(c *fiber.Ctx) error {
	page := h.GetIntQuery(c, "page", 1)
	size := h.GetIntQuery(c, "size", 20)
	keyword := h.GetQuery(c, "keyword")

	attachments, total, err := h.attachmentService.List(c.Context(), page, size, keyword)
	if err != nil {
		h.logger.Error("获取附件列表失败", zap.Error(err))
		return h.Error(c, 500, "获取附件列表失败")
	}

	return h.Success(c, fiber.Map{
		"list":  attachments,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// Delete 删除附件
func (h *AttachmentHandler) Delete(c *fiber.Ctx) error {
	idStr := h.GetParam(c, "id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		return h.Error(c, 400, "附件ID不能为空")
	}

	err := h.attachmentService.Delete(c.Context(), id)
	if err != nil {
		h.logger.Error("删除附件失败", zap.Error(err), zap.Int64("id", id))
		return h.Error(c, 500, "删除附件失败")
	}

	return h.Success(c, nil)
}
