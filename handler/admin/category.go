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

type CategoryHandler struct {
	CategoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		CategoryService: categoryService,
	}
}

func (c *CategoryHandler) GetCategoryByID(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "categoryID")
	if err != nil {
		return nil, err
	}
	category, err := c.CategoryService.GetByID(ctx.UserContext(), id)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTO(ctx.UserContext(), category)
}

func (c *CategoryHandler) ListAllCategory(ctx *fiber.Ctx) (interface{}, error) {
	categoryQuery := struct {
		*param.Sort
		More *bool `json:"more" form:"more"`
	}{}

	err := ctx.ShouldBindQuery(&categoryQuery)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if categoryQuery.Sort == nil || len(categoryQuery.Sort.Fields) == 0 {
		categoryQuery.Sort = &param.Sort{Fields: []string{"priority,asc"}}
	}
	if categoryQuery.More != nil && *categoryQuery.More {
		return c.CategoryService.ListCategoryWithPostCountDTO(ctx.UserContext(), categoryQuery.Sort)
	}
	categories, err := c.CategoryService.ListAll(ctx.UserContext(), categoryQuery.Sort)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTOs(ctx.UserContext(), categories)
}

func (c *CategoryHandler) ListAsTree(ctx *fiber.Ctx) (interface{}, error) {
	var sort param.Sort
	err := ctx.ShouldBindQuery(&sort)
	if err != nil {
		return nil, err
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "priority,asc")
	}
	return c.CategoryService.ListAsTree(ctx.UserContext(), &sort, false)
}

func (c *CategoryHandler) CreateCategory(ctx *fiber.Ctx) (interface{}, error) {
	var categoryParam param.Category
	err := util.BindAndValidate(ctx, &categoryParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	category, err := c.CategoryService.Create(ctx.UserContext(), &categoryParam)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTO(ctx.UserContext(), category)
}

func (c *CategoryHandler) UpdateCategory(ctx *fiber.Ctx) (interface{}, error) {
	var categoryParam param.Category
	err := util.BindAndValidate(ctx, &categoryParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	categoryID, err := util.ParamInt32(ctx, "categoryID")
	if err != nil {
		return nil, err
	}
	categoryParam.ID = categoryID
	category, err := c.CategoryService.Update(ctx.UserContext(), &categoryParam)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTO(ctx.UserContext(), category)
}

func (c *CategoryHandler) UpdateCategoryBatch(ctx *fiber.Ctx) (interface{}, error) {
	categoryParams := make([]*param.Category, 0)
	err := util.BindAndValidate(ctx, &categoryParams)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	categories, err := c.CategoryService.UpdateBatch(ctx.UserContext(), categoryParams)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTOs(ctx.UserContext(), categories)
}

func (c *CategoryHandler) DeleteCategory(ctx *fiber.Ctx) (interface{}, error) {
	categoryID, err := util.ParamInt32(ctx, "categoryID")
	if err != nil {
		return nil, err
	}
	return nil, c.CategoryService.Delete(ctx.UserContext(), categoryID)
}

