package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util"
)

type CommentHandler struct {
	BaseCommentService service.BaseCommentService
}

func NewCommentHandler(baseCommentService service.BaseCommentService) *CommentHandler {
	return &CommentHandler{
		BaseCommentService: baseCommentService,
	}
}

func (c *CommentHandler) Like(ctx *fiber.Ctx) (interface{}, error) {
	commentID, err := util.ParamInt32(ctx, "commentID")
	if err != nil {
		return nil, err
	}
	return nil, c.BaseCommentService.IncreaseLike(ctx.UserContext(), commentID)
}
