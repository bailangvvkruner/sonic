package content

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/handler/content/authentication"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/template"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type ViewHandler struct {
	OptionService          service.OptionService
	UserService            service.UserService
	CategoryService        service.CategoryService
	PostService            service.PostService
	ThemeService           service.ThemeService
	CategoryAuthentication *authentication.CategoryAuthentication
	PostAuthentication     *authentication.PostAuthentication
}

func NewViewHandler(
	optionService service.OptionService,
	userService service.UserService,
	categoryService service.CategoryService,
	postService service.PostService,
	themeService service.ThemeService,
	categoryAuthentication *authentication.CategoryAuthentication,
	postAuthentication *authentication.PostAuthentication,
) *ViewHandler {
	return &ViewHandler{
		OptionService:          optionService,
		UserService:            userService,
		CategoryService:        categoryService,
		PostService:            postService,
		ThemeService:           themeService,
		CategoryAuthentication: categoryAuthentication,
		PostAuthentication:     postAuthentication,
	}
}

func (v *ViewHandler) Admin(ctx *fiber.Ctx) (interface{}, error) {
	// TODO
	return nil, nil
}

func (v *ViewHandler) Version(ctx *fiber.Ctx) (interface{}, error) {
	return consts.SonicVersion, nil
}

func (v *ViewHandler) Install(ctx *fiber.Ctx) error {
	isInstall := v.OptionService.GetOrByDefault(ctx.UserContext(), property.IsInstalled).(bool)
	if isInstall {
		return nil
	}
	adminURLPath, _ := v.OptionService.GetAdminURLPath(ctx.UserContext())
	return ctx.Redirect(adminURLPath+"/#install", http.StatusTemporaryRedirect)
}

func (v *ViewHandler) Logo(ctx *fiber.Ctx) (interface{}, error) {
	logo := v.OptionService.GetOrByDefault(ctx.UserContext(), property.BlogLogo).(string)
	if logo != "" {
		return nil, ctx.Redirect(logo, http.StatusTemporaryRedirect)
	}
	return nil, nil
}

func (v *ViewHandler) Favicon(ctx *fiber.Ctx) (interface{}, error) {
	favicon := v.OptionService.GetOrByDefault(ctx.UserContext(), property.BlogFavicon).(string)
	if favicon != "" {
		return nil, ctx.Redirect(favicon, http.StatusTemporaryRedirect)
	}
	return nil, nil
}

func (v *ViewHandler) Authenticate(ctx *fiber.Ctx, model template.Model) (string, error) {
	contentType, err := util.ParamString(ctx, "type")
	if err != nil {
		return v.authenticateErr(ctx, model, contentType, "", err)
	}
	slug, err := util.ParamString(ctx, "slug")
	if err != nil {
		return v.authenticateErr(ctx, model, contentType, slug, err)
	}

	var authenticationParam param.Authentication
	err = util.BindAndValidate(ctx, &authenticationParam)
	if err != nil {
		return v.authenticateErr(ctx, model, "post", slug, err)
	}
	if authenticationParam.Password == "" {
		return v.authenticateErr(ctx, model, "post", slug, xerr.WithMsg(nil, "密码为空"))
	}

	token := ctx.Cookies("authentication")

	switch contentType {
	case consts.EncryptTypeCategory.Name():
		token, err = v.authenticateCategory(ctx, slug, authenticationParam.Password, token)
	case consts.EncryptTypePost.Name():
		token, err = v.authenticatePost(ctx, slug, authenticationParam.Password, token)
	default:
		return v.authenticateErr(ctx, model, "post", slug, xerr.WithStatus(nil, xerr.StatusBadRequest))
	}
	if err != nil {
		return v.authenticateErr(ctx, model, contentType, slug, err)
	}
	ctx.Cookie(&fiber.Cookie{
		Name:     "authentication",
		Value:    token,
		MaxAge:   1800,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
	})
	return "", nil
}

func (v *ViewHandler) authenticateCategory(ctx *fiber.Ctx, slug, password, token string) (string, error) {
	category, err := v.CategoryService.GetBySlug(ctx.UserContext(), slug)
	if err != nil {
		return "", err
	}
	categoryDTO, err := v.CategoryService.ConvertToCategoryDTO(ctx.UserContext(), category)
	if err != nil {
		return "", err
	}

	token, err = v.CategoryAuthentication.Authenticate(ctx.UserContext(), token, category.ID, password)
	if err != nil {
		return "", err
	}

	ctx.Redirect(categoryDTO.FullPath, http.StatusFound)
	return token, nil
}

func (v *ViewHandler) authenticatePost(ctx *fiber.Ctx, slug, password, token string) (string, error) {
	post, err := v.PostService.GetBySlug(ctx.UserContext(), slug)
	if err != nil {
		return "", err
	}
	fullPath, err := v.PostService.BuildFullPath(ctx.UserContext(), post)
	if err != nil {
		return "", err
	}
	token, err = v.PostAuthentication.Authenticate(ctx.UserContext(), token, post.ID, password)
	if err != nil {
		return "", err
	}

	ctx.Redirect(fullPath, http.StatusFound)
	return token, nil
}

func (v *ViewHandler) authenticateErr(ctx *fiber.Ctx, model template.Model, aType string, slug string, err error) (string, error) {
	model["type"] = aType
	model["slug"] = slug
	model["errorMsg"] = xerr.GetMessage(err)
	if exist, err := v.ThemeService.TemplateExist(ctx.UserContext(), "post_password.tmpl"); err == nil && exist {
		return v.ThemeService.Render(ctx.UserContext(), "post_password")
	}
	return "common/template/post_password", nil
}
