package admin

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/handler/trans"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type ThemeHandler struct {
	ThemeService  service.ThemeService
	OptionService service.OptionService
}

func NewThemeHandler(l service.ThemeService, o service.OptionService) *ThemeHandler {
	return &ThemeHandler{
		ThemeService:  l,
		OptionService: o,
	}
}

func (t *ThemeHandler) GetThemeByID(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeByID(ctx.UserContext(), themeID)
}

func (t *ThemeHandler) ListAllThemes(ctx *fiber.Ctx) (interface{}, error) {
	return t.ThemeService.ListAllTheme(ctx.UserContext())
}

func (t *ThemeHandler) ListActivatedThemeFile(ctx *fiber.Ctx) (interface{}, error) {
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(ctx.UserContext())
	if err != nil {
		return nil, err
	}
	return t.ThemeService.ListThemeFiles(ctx.UserContext(), activatedThemeID)
}

func (t *ThemeHandler) ListThemeFileByID(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.ListThemeFiles(ctx.UserContext(), themeID)
}

func (t *ThemeHandler) GetThemeFileContent(ctx *fiber.Ctx) (interface{}, error) {
	path, err := util.MustGetQueryString(ctx, "path")
	if err != nil {
		return nil, err
	}
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(ctx.UserContext())
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeFileContent(ctx.UserContext(), activatedThemeID, path)
}

func (t *ThemeHandler) GetThemeFileContentByID(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	path, err := util.MustGetQueryString(ctx, "path")
	if err != nil {
		return nil, err
	}

	return t.ThemeService.GetThemeFileContent(ctx.UserContext(), themeID, path)
}

func (t *ThemeHandler) UpdateThemeFile(ctx *fiber.Ctx) (interface{}, error) {
	themeParam := &param.ThemeContent{}
	err := ctx.ShouldBindJSON(themeParam)
	if err != nil {
		if err != nil {
			e := validator.ValidationErrors{}
			if errors.As(err, &e) {
				return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
			}
			return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
		}
	}
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(ctx.UserContext())
	if err != nil {
		return nil, err
	}
	return nil, t.ThemeService.UpdateThemeFile(ctx.UserContext(), activatedThemeID, themeParam.Path, themeParam.Content)
}

func (t *ThemeHandler) UpdateThemeFileByID(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	themeParam := &param.ThemeContent{}
	err = ctx.ShouldBindJSON(themeParam)
	if err != nil {
		if err != nil {
			e := validator.ValidationErrors{}
			if errors.As(err, &e) {
				return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
			}
			return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
		}
	}
	return nil, t.ThemeService.UpdateThemeFile(ctx.UserContext(), themeID, themeParam.Path, themeParam.Content)
}

func (t *ThemeHandler) ListCustomSheetTemplate(ctx *fiber.Ctx) (interface{}, error) {
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(ctx.UserContext())
	if err != nil {
		return nil, err
	}
	return t.ThemeService.ListCustomTemplates(ctx.UserContext(), activatedThemeID, consts.ThemeCustomSheetPrefix)
}

func (t *ThemeHandler) ListCustomPostTemplate(ctx *fiber.Ctx) (interface{}, error) {
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(ctx.UserContext())
	if err != nil {
		return nil, err
	}
	return t.ThemeService.ListCustomTemplates(ctx.UserContext(), activatedThemeID, consts.ThemeCustomPostPrefix)
}

func (t *ThemeHandler) ActivateTheme(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.ActivateTheme(ctx.UserContext(), themeID)
}

func (t *ThemeHandler) GetActivatedTheme(ctx *fiber.Ctx) (interface{}, error) {
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(ctx.UserContext())
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeByID(ctx.UserContext(), activatedThemeID)
}

func (t *ThemeHandler) GetActivatedThemeConfig(ctx *fiber.Ctx) (interface{}, error) {
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(ctx.UserContext())
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeConfig(ctx.UserContext(), activatedThemeID)
}

func (t *ThemeHandler) GetThemeConfigByID(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeConfig(ctx.UserContext(), themeID)
}

func (t *ThemeHandler) GetThemeConfigByGroup(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	group, err := util.ParamString(ctx, "group")
	if err != nil {
		return nil, err
	}
	themeSettings, err := t.ThemeService.GetThemeConfig(ctx.UserContext(), themeID)
	if err != nil {
		return nil, err
	}
	for _, setting := range themeSettings {
		if setting.Name == group {
			return setting.Items, nil
		}
	}
	return nil, nil
}

func (t *ThemeHandler) GetThemeConfigGroupNames(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	themeSettings, err := t.ThemeService.GetThemeConfig(ctx.UserContext(), themeID)
	if err != nil {
		return nil, err
	}
	groupNames := make([]string, len(themeSettings))
	for index, setting := range themeSettings {
		groupNames[index] = setting.Name
	}
	return groupNames, nil
}

func (t *ThemeHandler) GetActivatedThemeSettingMap(ctx *fiber.Ctx) (interface{}, error) {
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(ctx.UserContext())
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeSettingMap(ctx.UserContext(), activatedThemeID)
}

func (t *ThemeHandler) GetThemeSettingMapByID(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeSettingMap(ctx.UserContext(), themeID)
}

func (t *ThemeHandler) GetThemeSettingMapByGroupAndThemeID(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	group, err := util.ParamString(ctx, "group")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeGroupSettingMap(ctx.UserContext(), themeID, group)
}

func (t *ThemeHandler) SaveActivatedThemeSetting(ctx *fiber.Ctx) (interface{}, error) {
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(ctx.UserContext())
	if err != nil {
		return nil, err
	}
	settings := make(map[string]interface{})
	err = util.BindAndValidate(ctx, &settings)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	return nil, t.ThemeService.SaveThemeSettings(ctx.UserContext(), activatedThemeID, settings)
}

func (t *ThemeHandler) SaveThemeSettingByID(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	settings := make(map[string]interface{})
	err = util.BindAndValidate(ctx, &settings)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	return nil, t.ThemeService.SaveThemeSettings(ctx.UserContext(), themeID, settings)
}

func (t *ThemeHandler) DeleteThemeByID(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	isDeleteSetting, err := util.GetQueryBool(ctx, "deleteSettings", false)
	if err != nil {
		return nil, err
	}
	return nil, t.ThemeService.DeleteTheme(ctx.UserContext(), themeID, isDeleteSetting)
}

func (t *ThemeHandler) UploadTheme(ctx *fiber.Ctx) (interface{}, error) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return nil, xerr.WithMsg(err, "upload theme error").WithStatus(xerr.StatusBadRequest)
	}
	return t.ThemeService.UploadTheme(ctx.UserContext(), fileHeader)
}

func (t *ThemeHandler) UpdateThemeByUpload(ctx *fiber.Ctx) (interface{}, error) {
	themeID, err := util.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return nil, xerr.WithMsg(err, "upload theme error").WithStatus(xerr.StatusBadRequest)
	}
	return t.ThemeService.UpdateThemeByUpload(ctx.UserContext(), themeID, fileHeader)
}

func (t *ThemeHandler) FetchTheme(ctx *fiber.Ctx) (interface{}, error) {
	uri, _ := util.MustGetQueryString(ctx, "uri")
	return t.ThemeService.Fetch(ctx.UserContext(), uri)
}

func (t *ThemeHandler) UpdateThemeByFetching(ctx *fiber.Ctx) (interface{}, error) {
	return nil, xerr.WithMsg(nil, "not support").WithStatus(xerr.StatusInternalServerError)
}

func (t *ThemeHandler) ReloadTheme(ctx *fiber.Ctx) (interface{}, error) {
	return nil, t.ThemeService.ReloadTheme(ctx.UserContext())
}

func (t *ThemeHandler) TemplateExist(ctx *fiber.Ctx) (interface{}, error) {
	template, err := util.MustGetQueryString(ctx, "template")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.TemplateExist(ctx.UserContext(), template)
}

