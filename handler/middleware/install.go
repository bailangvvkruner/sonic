package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
)

type InstallRedirectMiddleware struct {
	optionService service.OptionService
}

func NewInstallRedirectMiddleware(optionService service.OptionService) *InstallRedirectMiddleware {
	return &InstallRedirectMiddleware{
		optionService: optionService,
	}
}

func (i *InstallRedirectMiddleware) InstallRedirect() fiber.Handler {
	skipPath := map[string]struct{}{
		"/api/admin/installations":  {},
		"/api/admin/is_installed":   {},
		"/api/admin/login/precheck": {},
	}
	return func(ctx *fiber.Ctx) error {
		path := ctx.Path()
		if _, ok := skipPath[path]; ok {
			return ctx.Next()
		}
		isInstall, err := i.optionService.GetOrByDefaultWithErr(ctx.UserContext(), property.IsInstalled, false)
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).JSON(&dto.BaseDTO{
				Status:  http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			})
		}
		if !isInstall.(bool) {
			return ctx.Redirect("/admin/#install", http.StatusFound)
		}
		return ctx.Next()
	}
}
