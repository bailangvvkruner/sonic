package content

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/handler/content/model"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/service/assembler"
	"github.com/go-sonic/sonic/template"
	"github.com/go-sonic/sonic/util"
)

type CategoryHandler struct {
	OptionService       service.OptionService
	PostService         service.PostService
	PostCategoryService service.PostCategoryService
	CategoryService     service.CategoryService
	PostAssembler       assembler.PostAssembler
	PostModel           *model.PostModel
	CategoryModel       *model.CategoryModel
}

func NewCategoryHandler(
	optionService service.OptionService,
	postService service.PostService,
	categoryService service.CategoryService,
	postCategoryService service.PostCategoryService,
	postAssembler assembler.PostAssembler,
	postModel *model.PostModel,
	categoryModel *model.CategoryModel,
) *CategoryHandler {
	return &CategoryHandler{
		OptionService:       optionService,
		PostService:         postService,
		PostCategoryService: postCategoryService,
		CategoryService:     categoryService,
		PostAssembler:       postAssembler,
		PostModel:           postModel,
		CategoryModel:       categoryModel,
	}
}

func (c *CategoryHandler) Categories(ctx *fiber.Ctx, model template.Model) (string, error) {
	return c.CategoryModel.ListCategories(ctx.UserContext(), model)
}

func (c *CategoryHandler) CategoryDetail(ctx *fiber.Ctx, model template.Model) (string, error) {
	slug, err := util.ParamString(ctx, "slug")
	if err != nil {
		return "", err
	}
	token := ctx.Cookies("authentication")
	return c.CategoryModel.CategoryDetail(ctx.UserContext(), model, slug, 0, token)
}

func (c *CategoryHandler) CategoryDetailPage(ctx *fiber.Ctx, model template.Model) (string, error) {
	slug, err := util.ParamString(ctx, "slug")
	if err != nil {
		return "", err
	}

	page, err := util.ParamInt32(ctx, "page")
	if err != nil {
		return "", err
	}
	token := ctx.Cookies("authentication")
	return c.CategoryModel.CategoryDetail(ctx.UserContext(), model, slug, int(page-1), token)
}
