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

type MenuHandler struct {
	MenuService service.MenuService
}

func NewMenuHandler(menuService service.MenuService) *MenuHandler {
	return &MenuHandler{
		MenuService: menuService,
	}
}

func (m *MenuHandler) ListMenus(ctx *fiber.Ctx) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.QueryParser(&sort)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "team,desc", "priority,asc")
	} else {
		sort.Fields = append(sort.Fields, "priority,asc")
	}
	menus, err := m.MenuService.List(ctx.UserContext(), &sort)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTOs(ctx.UserContext(), menus), nil
}

func (m *MenuHandler) ListMenusAsTree(ctx *fiber.Ctx) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.QueryParser(&sort)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "team,desc", "priority,asc")
	} else {
		sort.Fields = append(sort.Fields, "priority,asc")
	}
	menus, err := m.MenuService.ListAsTree(ctx.UserContext(), &sort)
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (m *MenuHandler) ListMenusAsTreeByTeam(ctx *fiber.Ctx) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.QueryParser(&sort)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "priority,asc")
	}
	team, _ := util.MustGetQueryString(ctx, "team")
	if team == "" {
		menus, err := m.MenuService.ListAsTree(ctx.UserContext(), &sort)
		if err != nil {
			return nil, err
		}
		return menus, nil
	}
	menus, err := m.MenuService.ListAsTreeByTeam(ctx.UserContext(), team, &sort)
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (m *MenuHandler) GetMenuByID(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	menu, err := m.MenuService.GetByID(ctx.UserContext(), id)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTO(ctx.UserContext(), menu), nil
}

func (m *MenuHandler) CreateMenu(ctx *fiber.Ctx) (interface{}, error) {
	menuParam := &param.Menu{}
	err := ctx.BodyParser(menuParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	menu, err := m.MenuService.Create(ctx.UserContext(), menuParam)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTO(ctx.UserContext(), menu), nil
}

func (m *MenuHandler) CreateMenuBatch(ctx *fiber.Ctx) (interface{}, error) {
	menuParams := make([]*param.Menu, 0)
	err := util.BindAndValidate(ctx, &menuParams)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	menus, err := m.MenuService.CreateBatch(ctx.UserContext(), menuParams)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTOs(ctx.UserContext(), menus), nil
}

func (m *MenuHandler) UpdateMenu(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	menuParam := &param.Menu{}
	err = ctx.BodyParser(menuParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	menu, err := m.MenuService.Update(ctx.UserContext(), id, menuParam)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTO(ctx.UserContext(), menu), nil
}

func (m *MenuHandler) UpdateMenuBatch(ctx *fiber.Ctx) (interface{}, error) {
	menuParams := make([]*param.Menu, 0)
	err := util.BindAndValidate(ctx, &menuParams)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	menus, err := m.MenuService.UpdateBatch(ctx.UserContext(), menuParams)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTOs(ctx.UserContext(), menus), nil
}

func (m *MenuHandler) DeleteMenu(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return nil, m.MenuService.Delete(ctx.UserContext(), id)
}

func (m *MenuHandler) DeleteMenuBatch(ctx *fiber.Ctx) (interface{}, error) {
	menuIDs := make([]int32, 0)
	err := ctx.BodyParser(&menuIDs)
	if err != nil {
		return nil, xerr.WithMsg(err, "menuIDs error").WithStatus(xerr.StatusBadRequest)
	}
	return nil, m.MenuService.DeleteBatch(ctx.UserContext(), menuIDs)
}

func (m *MenuHandler) ListMenuTeams(ctx *fiber.Ctx) (interface{}, error) {
	return m.MenuService.ListTeams(ctx.UserContext())
}

