package admin

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"


	"github.com/go-sonic/sonic/handler/trans"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type LinkHandler struct {
	LinkService service.LinkService
}

func NewLinkHandler(linkService service.LinkService) *LinkHandler {
	return &LinkHandler{
		LinkService: linkService,
	}
}

func (l *LinkHandler) ListLinks(ctx *fiber.Ctx) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.QueryParser(&sort)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "team,desc", "priority,asc")
	} else {
		sort.Fields = append(sort.Fields, "priority,asc")
	}
	links, err := l.LinkService.List(ctx.UserContext(), &sort)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToDTOs(ctx.UserContext(), links), nil
}

func (l *LinkHandler) GetLinkByID(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	link, err := l.LinkService.GetByID(ctx.UserContext(), id)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToDTO(ctx.UserContext(), link), nil
}

func (l *LinkHandler) CreateLink(ctx *fiber.Ctx) (interface{}, error) {
	linkParam := &param.Link{}
	err := ctx.BodyParser(linkParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	link, err := l.LinkService.Create(ctx.UserContext(), linkParam)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToDTO(ctx.UserContext(), link), nil
}

func (l *LinkHandler) UpdateLink(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	linkParam := &param.Link{}
	err = ctx.BodyParser(linkParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	link, err := l.LinkService.Update(ctx.UserContext(), id, linkParam)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToDTO(ctx.UserContext(), link), nil
}

func (l *LinkHandler) DeleteLink(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return nil, l.LinkService.Delete(ctx.UserContext(), id)
}

func (l *LinkHandler) ListLinkTeams(ctx *fiber.Ctx) (interface{}, error) {
	return l.LinkService.ListTeams(ctx.UserContext())
}

