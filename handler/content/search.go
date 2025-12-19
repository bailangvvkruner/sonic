package content

import (
	"html"

	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/service/assembler"
	"github.com/go-sonic/sonic/template"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type SearchHandler struct {
	PostAssembler assembler.PostAssembler
	PostService   service.PostService
	OptionService service.OptionService
	ThemeService  service.ThemeService
}

func NewSearchHandler(
	postAssembler assembler.PostAssembler,
	postService service.PostService,
	optionService service.OptionService,
	themeService service.ThemeService,
) *SearchHandler {
	return &SearchHandler{
		PostAssembler: postAssembler,
		PostService:   postService,
		OptionService: optionService,
		ThemeService:  themeService,
	}
}

func (s *SearchHandler) Search(ctx *fiber.Ctx, model template.Model) (string, error) {
	return s.search(ctx, 0, model)
}

func (s *SearchHandler) PageSearch(ctx *fiber.Ctx, model template.Model) (string, error) {
	page, err := util.ParamInt32(ctx, "page")
	if err != nil {
		return "", err
	}
	return s.search(ctx, int(page)-1, model)
}

func (s *SearchHandler) search(ctx *fiber.Ctx, pageNum int, model template.Model) (string, error) {
	keyword, err := util.MustGetQueryString(ctx, "keyword")
	if err != nil {
		return "", err
	}
	sort := param.Sort{}
	err = util.BindAndValidate(ctx, &sort)
	if err != nil {
		return "", xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if len(sort.Fields) == 0 {
		sort = s.OptionService.GetPostSort(ctx.UserContext())
	}
	defaultPageSize := s.OptionService.GetIndexPageSize(ctx.UserContext())
	page := param.Pagination{
		PageNum:  pageNum,
		PageSize: defaultPageSize,
	}
	postQuery := param.PostQuery{
		Pagination: page,
		Sort:       sort,
		Keyword:    &keyword,
		Statuses:   []consts.PostStatus{consts.PostStatusPublished},
	}
	posts, total, err := s.PostService.Page(ctx.UserContext(), postQuery)
	if err != nil {
		return "", err
	}
	postVOs, err := s.PostAssembler.ConvertToListVO(ctx.UserContext(), posts)
	if err != nil {
		return "", err
	}
	model["is_search"] = true
	model["keyword"] = html.EscapeString(keyword)
	model["posts"] = dto.NewPage(postVOs, total, page)
	model["meta_keywords"] = s.OptionService.GetOrByDefault(ctx.UserContext(), property.SeoKeywords)
	model["meta_description"] = s.OptionService.GetOrByDefault(ctx.UserContext(), property.SeoDescription)
	return s.ThemeService.Render(ctx.UserContext(), "search")
}
