package admin

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/handler/trans"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/service/assembler"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type SheetHandler struct {
	SheetService   service.SheetService
	PostService    service.PostService
	SheetAssembler assembler.SheetAssembler
}

func NewSheetHandler(sheetService service.SheetService, postService service.PostService, sheetAssembler assembler.SheetAssembler) *SheetHandler {
	return &SheetHandler{
		SheetService:   sheetService,
		PostService:    postService,
		SheetAssembler: sheetAssembler,
	}
}

func (s *SheetHandler) GetSheetByID(ctx *fiber.Ctx) (interface{}, error) {
	sheetID, err := util.ParamInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	sheet, err := s.SheetService.GetByPostID(ctx.UserContext(), sheetID)
	if err != nil {
		return nil, err
	}
	return s.SheetAssembler.ConvertToDetailVO(ctx.UserContext(), sheet)
}

func (s *SheetHandler) ListSheet(ctx *fiber.Ctx) (interface{}, error) {
	type SheetParam struct {
		param.Page
		Sort string `json:"sort"`
	}
	var sheetParam SheetParam
	err := ctx.BodyParser(&sheetParam)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	sheets, totalCount, err := s.SheetService.Page(ctx.UserContext(), sheetParam.Page, &param.Sort{Fields: []string{"createTime,desc"}})
	if err != nil {
		return nil, err
	}
	sheetVOs, err := s.SheetAssembler.ConvertToListVO(ctx.UserContext(), sheets)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(sheetVOs, totalCount, sheetParam.Page), nil
}

func (s *SheetHandler) IndependentSheets(ctx *fiber.Ctx) (interface{}, error) {
	return s.SheetService.ListIndependentSheets(ctx.UserContext())
}

func (s *SheetHandler) CreateSheet(ctx *fiber.Ctx) (interface{}, error) {
	var sheetParam param.Sheet
	err := util.BindAndValidate(ctx, &sheetParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	sheet, err := s.SheetService.Create(ctx.UserContext(), &sheetParam)
	if err != nil {
		return nil, err
	}
	sheetDetailVO, err := s.SheetAssembler.ConvertToDetailVO(ctx.UserContext(), sheet)
	if err != nil {
		return nil, err
	}
	return sheetDetailVO, nil
}

func (s *SheetHandler) UpdateSheet(ctx *fiber.Ctx) (interface{}, error) {
	var sheetParam param.Sheet
	err := util.BindAndValidate(ctx, &sheetParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}

	sheetID, err := util.ParamInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	postDetailVO, err := s.SheetService.Update(ctx.UserContext(), sheetID, &sheetParam)
	if err != nil {
		return nil, err
	}
	return postDetailVO, nil
}

func (s *SheetHandler) UpdateSheetStatus(ctx *fiber.Ctx) (interface{}, error) {
	sheetID, err := util.ParamInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	statusStr, err := util.ParamString(ctx, "status")
	if err != nil {
		return nil, err
	}
	status, err := consts.PostStatusFromString(statusStr)
	if err != nil {
		return nil, err
	}
	if status < consts.PostStatusPublished || status > consts.PostStatusIntimate {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("status error")
	}
	return s.SheetService.UpdateStatus(ctx.UserContext(), sheetID, status)
}

func (s *SheetHandler) UpdateSheetDraft(ctx *fiber.Ctx) (interface{}, error) {
	sheetID, err := util.ParamInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	var postContentParam param.PostContent
	err = util.BindAndValidate(ctx, &postContentParam)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("content param error")
	}
	post, err := s.SheetService.UpdateDraftContent(ctx.UserContext(), sheetID, postContentParam.Content, postContentParam.OriginalContent)
	if err != nil {
		return nil, err
	}
	return s.SheetAssembler.ConvertToDetailDTO(ctx.UserContext(), post)
}

func (s *SheetHandler) DeleteSheet(ctx *fiber.Ctx) (interface{}, error) {
	sheetID, err := util.ParamInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	return nil, s.SheetService.Delete(ctx.UserContext(), sheetID)
}

func (s *SheetHandler) PreviewSheet(ctx *fiber.Ctx) {
	sheetID, err := util.ParamInt32(ctx, "sheetID")
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		_ = ctx.Error(err)
		return
	}

	previewPath, err := s.SheetService.Preview(ctx.UserContext(), sheetID)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		_ = ctx.Error(err)
		return
	}
	ctx.Status(http.StatusOK).SendString(previewPath)
}

