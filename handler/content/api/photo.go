package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util"
)

type PhotoHandler struct {
	PhotoService service.PhotoService
}

func NewPhotoHandler(photoService service.PhotoService) *PhotoHandler {
	return &PhotoHandler{
		PhotoService: photoService,
	}
}

func (p *PhotoHandler) Like(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "photoID")
	if err != nil {
		return nil, err
	}
	return nil, p.PhotoService.IncreaseLike(ctx.UserContext(), id)
}
