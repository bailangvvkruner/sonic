package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/service"
)

type LinkHandler struct {
	LinkService service.LinkService
}

func NewLinkHandler(linkService service.LinkService) *LinkHandler {
	return &LinkHandler{
		LinkService: linkService,
	}
}

type linkParam struct {
	*param.Sort
}

func (l *LinkHandler) ListLinks(ctx *fiber.Ctx) (interface{}, error) {
	p := linkParam{}
	if err := ctx.QueryParser(&p); err != nil {
		return nil, err
	}

	if p.Sort == nil || len(p.Sort.Fields) == 0 {
		p.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	links, err := l.LinkService.List(ctx.UserContext(), p.Sort)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToDTOs(ctx.UserContext(), links), nil
}

func (l *LinkHandler) LinkTeamVO(ctx *fiber.Ctx) (interface{}, error) {
	p := linkParam{}
	if err := ctx.QueryParser(&p); err != nil {
		return nil, err
	}

	if p.Sort == nil || len(p.Sort.Fields) == 0 {
		p.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	links, err := l.LinkService.List(ctx.UserContext(), p.Sort)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToLinkTeamVO(ctx.UserContext(), links), nil
}
