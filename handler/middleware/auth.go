package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/cache"
	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util/xerr"
)

type AuthMiddleware struct {
	OptionService       service.OptionService
	OneTimeTokenService service.OneTimeTokenService
	UserService         service.UserService
	Cache               cache.Cache
}

func NewAuthMiddleware(optionService service.OptionService, oneTimeTokenService service.OneTimeTokenService, cache cache.Cache, userService service.UserService) *AuthMiddleware {
	authMiddleware := &AuthMiddleware{
		OptionService:       optionService,
		OneTimeTokenService: oneTimeTokenService,
		Cache:               cache,
		UserService:         userService,
	}
	return authMiddleware
}

func (a *AuthMiddleware) GetWrapHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		isInstalled, err := a.OptionService.GetOrByDefaultWithErr(ctx.Context(), property.IsInstalled, false)
		if err != nil {
			return abortWithStatusJSON(ctx, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		if !isInstalled.(bool) {
			return abortWithStatusJSON(ctx, http.StatusBadRequest, "Blog is not initialized")
		}

		oneTimeToken := ctx.Query(consts.OneTimeTokenQueryName)
		if oneTimeToken != "" {
			allowedURL, ok := a.OneTimeTokenService.Get(oneTimeToken)
			if !ok {
				return abortWithStatusJSON(ctx, http.StatusBadRequest, "OneTimeToken is not exist or expired")
			}
			currentURL := ctx.Path()
			if currentURL != allowedURL {
				return abortWithStatusJSON(ctx, http.StatusBadRequest, "The one-time token does not correspond the request uri")
			}
			return ctx.Next()
		}

		token := ctx.Get(consts.AdminTokenHeaderName)
		if token == "" {
			return abortWithStatusJSON(ctx, http.StatusUnauthorized, "未登录，请登录后访问")
		}
		userID, ok := a.Cache.Get(cache.BuildTokenAccessKey(token))

		if !ok || userID == nil {
			return abortWithStatusJSON(ctx, http.StatusUnauthorized, "Token 已过期或不存在")
		}

		user, err := a.UserService.GetByID(ctx.Context(), userID.(int32))
		if xerr.GetType(err) == xerr.NoRecord {
			return abortWithStatusJSON(ctx, http.StatusUnauthorized, "用户不存在")
		}
		if err != nil {
			return abortWithStatusJSON(ctx, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		ctx.Locals(consts.AuthorizedUser, user)
		return ctx.Next()
	}
}

func abortWithStatusJSON(ctx *fiber.Ctx, status int, message string) error {
	ctx.Status(status)
	return ctx.JSON(&dto.BaseDTO{
		Status:  status,
		Message: message,
	})
}
