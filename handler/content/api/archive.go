package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/service/assembler"
)

type ArchiveHandler struct {
	PostService   service.PostService
	PostAssembler assembler.PostAssembler
}

func NewArchiveHandler(postService service.PostService, postAssemeber assembler.PostAssembler) *ArchiveHandler {
	return &ArchiveHandler{
		PostService:   postService,
		PostAssembler: postAssemeber,
	}
}

func (a *ArchiveHandler) ListYearArchives(ctx *fiber.Ctx) (interface{}, error) {
	posts, err := a.PostService.GetByStatus(ctx.Context(), []consts.PostStatus{consts.PostStatusPublished}, consts.PostTypePost, nil)
	if err != nil {
		return nil, err
	}
	return a.PostAssembler.ConvertToArchiveYearVOs(ctx.Context(), posts)
}

func (a *ArchiveHandler) ListMonthArchives(ctx *fiber.Ctx) (interface{}, error) {
	posts, err := a.PostService.GetByStatus(ctx.Context(), []consts.PostStatus{consts.PostStatusPublished}, consts.PostTypePost, nil)
	if err != nil {
		return nil, err
	}
	return a.PostAssembler.ConvertTOArchiveMonthVOs(ctx.Context(), posts)
}
