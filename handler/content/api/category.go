package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/handler/content/authentication"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/service/assembler"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type CategoryHandler struct {
	PostService            service.PostService
	CategoryService        service.CategoryService
	CategoryAuthentication authentication.CategoryAuthentication
	PostAssembler          assembler.PostAssembler
}

func NewCategoryHandler(postService service.PostService, categoryService service.CategoryService, categoryAuthentication *authentication.CategoryAuthentication, postAssembler assembler.PostAssembler) *CategoryHandler {
	return &CategoryHandler{
		PostService:            postService,
		CategoryService:        categoryService,
		CategoryAuthentication: *categoryAuthentication,
		PostAssembler:          postAssembler,
	}
}

func (c *CategoryHandler) ListCategories(ctx *fiber.Ctx) (interface{}, error) {
	categoryQuery := struct {
		*param.Sort
		More *bool `json:"more" query:"more" form:"more"`
	}{}

	err := ctx.QueryParser(&categoryQuery)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if categoryQuery.Sort == nil || len(categoryQuery.Sort.Fields) == 0 {
		categoryQuery.Sort = &param.Sort{Fields: []string{"updateTime,desc"}}
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

func (c *CategoryHandler) ListPosts(ctx *fiber.Ctx) (interface{}, error) {
	slug, err := util.ParamString(ctx, "slug")
	if err != nil {
		return nil, err
	}
	category, err := c.CategoryService.GetBySlug(ctx.UserContext(), slug)
	if err != nil {
		return nil, err
	}
	postQuery := param.PostQuery{}
	err = ctx.QueryParser(&postQuery)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if postQuery.Sort == nil {
		postQuery.Sort = &param.Sort{Fields: []string{"topPriority,desc", "updateTime,desc"}}
	}
	password := ctx.Query("password")

	if category.Type == consts.CategoryTypeIntimate {
		token := ctx.Cookies("authentication")
		if authenticated, _ := c.CategoryAuthentication.IsAuthenticated(ctx.UserContext(), token, category.ID); !authenticated {
			token, err := c.CategoryAuthentication.Authenticate(ctx.UserContext(), token, category.ID, password)
			if err != nil {
				return nil, err
			}
			ctx.Cookie(&fiber.Cookie{
				Name:     "authentication",
				Value:    token,
				MaxAge:   1800,
				Path:     "/",
				HTTPOnly: true,
				Secure:   false,
			})
		}
	}
	postQuery.WithPassword = util.BoolPtr(false)
	postQuery.Statuses = []*consts.PostStatus{consts.PostStatusPublished.Ptr(), consts.PostStatusIntimate.Ptr()}
	posts, totalCount, err := c.PostService.Page(ctx.UserContext(), postQuery)
	if err != nil {
		return nil, err
	}
	postVOs, err := c.PostAssembler.ConvertToListVO(ctx.UserContext(), posts)
	return dto.NewPage(postVOs, totalCount, postQuery.Page), err
}
