package content

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/handler/content/model"
	"github.com/go-sonic/sonic/template"
	"github.com/go-sonic/sonic/util"
)

type IndexHandler struct {
	PostModel *model.PostModel
}

func NewIndexHandler(postModel *model.PostModel) *IndexHandler {
	return &IndexHandler{
		PostModel: postModel,
	}
}

func (h *IndexHandler) Index(ctx *fiber.Ctx, model template.Model) (string, error) {
	return h.PostModel.List(ctx.UserContext(), 0, model)
}

func (h *IndexHandler) IndexPage(ctx *fiber.Ctx, model template.Model) (string, error) {
	page, err := util.ParamInt32(ctx, "page")
	if err != nil {
		return "", err
	}
	return h.PostModel.List(ctx.UserContext(), int(page)-1, model)
}
