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

type TagHandler struct {
	PostTagService service.PostTagService
	TagService     service.TagService
}

func NewTagHandler(postTagService service.PostTagService, tagService service.TagService) *TagHandler {
	return &TagHandler{
		PostTagService: postTagService,
		TagService:     tagService,
	}
}

func (t *TagHandler) ListTags(ctx *fiber.Ctx) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.ShouldBindQuery(&sort)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "createTime,desc")
	}
	more, _ := util.MustGetQueryBool(ctx, "more")
	if more {
		return t.PostTagService.ListAllTagWithPostCount(ctx.UserContext(), &sort)
	}
	tags, err := t.TagService.ListAll(ctx.UserContext(), &sort)
	if err != nil {
		return nil, err
	}
	return t.TagService.ConvertToDTOs(ctx.UserContext(), tags)
}

func (t *TagHandler) GetTagByID(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	tag, err := t.TagService.GetByID(ctx.UserContext(), id)
	if err != nil {
		return nil, err
	}
	return t.TagService.ConvertToDTO(ctx.UserContext(), tag)
}

func (t *TagHandler) CreateTag(ctx *fiber.Ctx) (interface{}, error) {
	tagParam := &param.Tag{}
	err := ctx.ShouldBindJSON(tagParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	tag, err := t.TagService.Create(ctx.UserContext(), tagParam)
	if err != nil {
		return nil, err
	}
	return t.TagService.ConvertToDTO(ctx.UserContext(), tag)
}

func (t *TagHandler) UpdateTag(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	tagParam := &param.Tag{}
	err = ctx.ShouldBindJSON(tagParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	tag, err := t.TagService.Update(ctx.UserContext(), id, tagParam)
	if err != nil {
		return nil, err
	}
	return t.TagService.ConvertToDTO(ctx.UserContext(), tag)
}

func (t *TagHandler) DeleteTag(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return nil, t.TagService.Delete(ctx.UserContext(), id)
}

