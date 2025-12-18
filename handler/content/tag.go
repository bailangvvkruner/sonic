package content

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/handler/content/model"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/template"
	"github.com/go-sonic/sonic/util"
)

type TagHandler struct {
	OptionService  service.OptionService
	TagService     service.TagService
	TagModel       *model.TagModel
	PostTagService service.PostTagService
}

func NewTagHandler(
	optionService service.OptionService,
	tagService service.TagService,
	tagModel *model.TagModel,
	postTagService service.PostTagService,
) *TagHandler {
	return &TagHandler{
		OptionService:  optionService,
		TagService:     tagService,
		TagModel:       tagModel,
		PostTagService: postTagService,
	}
}

func (t *TagHandler) Tags(ctx *fiber.Ctx, model template.Model) (string, error) {
	return t.TagModel.Tags(ctx, model)
}

func (t *TagHandler) TagPost(ctx *fiber.Ctx, model template.Model) (string, error) {
	slug, err := util.ParamString(ctx, "slug")
	if err != nil {
		return "", err
	}
	return t.TagModel.TagPosts(ctx, model, slug, 0)
}

func (t *TagHandler) TagPostPage(ctx *fiber.Ctx, model template.Model) (string, error) {
	slug, err := util.ParamString(ctx, "slug")
	if err != nil {
		return "", err
	}
	page, err := util.ParamInt32(ctx, "page")
	if err != nil {
		return "", err
	}
	return t.TagModel.TagPosts(ctx, model, slug, int(page-1))
}
