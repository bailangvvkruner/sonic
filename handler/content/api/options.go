package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
)

type OptionHandler struct {
	OptionService service.OptionService
}

func NewOptionHandler(
	optionService service.OptionService,
) *OptionHandler {
	return &OptionHandler{
		OptionService: optionService,
	}
}

func (o *OptionHandler) Comment(ctx *fiber.Ctx) (interface{}, error) {
	result := make(map[string]interface{})

	result[property.CommentGravatarSource.KeyValue] = o.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentGravatarSource)
	result[property.CommentGravatarDefault.KeyValue] = o.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentGravatarDefault)
	result[property.CommentContentPlaceholder.KeyValue] = o.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentContentPlaceholder)
	return result, nil
}
