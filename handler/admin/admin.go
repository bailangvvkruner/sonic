package admin

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"

	"github.com/go-sonic/sonic/handler/trans"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type AdminHandler struct {
	OptionService       service.OptionService
	AdminService        service.AdminService
	TwoFactorMFAService service.TwoFactorTOTPMFAService
}

func NewAdminHandler(optionService service.OptionService, adminService service.AdminService, twoFactorMFA service.TwoFactorTOTPMFAService) *AdminHandler {
	return &AdminHandler{
		OptionService:       optionService,
		AdminService:        adminService,
		TwoFactorMFAService: twoFactorMFA,
	}
}

func (a *AdminHandler) IsInstalled(ctx *fiber.Ctx) (interface{}, error) {
	return a.OptionService.GetOrByDefaultWithErr(ctx.UserContext(), property.IsInstalled, false)
}

func (a *AdminHandler) AuthPreCheck(ctx *fiber.Ctx) (interface{}, error) {
	var loginParam param.LoginParam
	err := util.BindAndValidate(ctx, &loginParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.BadParam.Wrapf(err, "")
	}

	user, err := a.AdminService.Authenticate(ctx.UserContext(), loginParam)
	if err != nil {
		return nil, err
	}
	return &dto.LoginPreCheckDTO{NeedMFACode: a.TwoFactorMFAService.UseMFA(user.MfaType)}, nil
}

func (a *AdminHandler) Auth(ctx *fiber.Ctx) (interface{}, error) {
	var loginParam param.LoginParam
	err := util.BindAndValidate(ctx, &loginParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.BadParam.Wrapf(err, "").WithStatus(xerr.StatusBadRequest)
	}

	return a.AdminService.Auth(ctx.UserContext(), loginParam)
}

func (a *AdminHandler) LogOut(ctx *fiber.Ctx) (interface{}, error) {
	err := a.AdminService.ClearToken(ctx.UserContext())
	return nil, err
}

func (a *AdminHandler) SendResetCode(ctx *fiber.Ctx) (interface{}, error) {
	var resetPasswordParam param.ResetPasswordParam
	err := util.BindAndValidate(ctx, &resetPasswordParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.BadParam.Wrapf(err, "").WithStatus(xerr.StatusBadRequest)
	}
	return nil, a.AdminService.SendResetPasswordCode(ctx.UserContext(), resetPasswordParam)
}

func (a *AdminHandler) RefreshToken(ctx *fiber.Ctx) (interface{}, error) {
	refreshToken := ctx.Params("refreshToken")
	if refreshToken == "" {
		return nil, xerr.BadParam.New("refreshToken参数为空").WithStatus(xerr.StatusBadRequest).
			WithMsg("refreshToken 参数不能为空")
	}
	return a.AdminService.RefreshToken(ctx.UserContext(), refreshToken)
}

func (a *AdminHandler) GetEnvironments(ctx *fiber.Ctx) (interface{}, error) {
	return a.AdminService.GetEnvironments(ctx.UserContext()), nil
}

func (a *AdminHandler) GetLogFiles(ctx *fiber.Ctx) (interface{}, error) {
	lines, err := util.MustGetQueryInt64(ctx, "lines")
	if err != nil {
		return nil, err
	}
	return a.AdminService.GetLogFiles(ctx.UserContext(), lines)
}

