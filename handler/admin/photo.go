package admin

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"

	"github.com/go-sonic/sonic/handler/trans"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type PhotoHandler struct {
	PhotoService service.PhotoService
}

func NewPhotoHandler(photoService service.PhotoService) *PhotoHandler {
	return &PhotoHandler{
		PhotoService: photoService,
	}
}

func (p *PhotoHandler) ListPhoto(ctx *fiber.Ctx) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.ShouldBindQuery(&sort)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "createTime,desc")
	}
	photos, err := p.PhotoService.List(ctx.UserContext(), &sort)
	if err != nil {
		return nil, err
	}
	return p.PhotoService.ConvertToDTOs(ctx.UserContext(), photos), nil
}

func (p *PhotoHandler) PagePhotos(ctx *fiber.Ctx) (interface{}, error) {
	type Param struct {
		param.Page
		param.Sort
	}
	param := Param{}
	err := ctx.ShouldBindQuery(&param)
	if err != nil {
		return nil, xerr.WithMsg(err, "parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(param.Fields) == 0 {
		param.Fields = append(param.Fields, "createTime,desc")
	}
	photos, totalCount, err := p.PhotoService.Page(ctx.UserContext(), param.Page, &param.Sort)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(p.PhotoService.ConvertToDTOs(ctx.UserContext(), photos), totalCount, param.Page), nil
}

func (p *PhotoHandler) GetPhotoByID(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	photo, err := p.PhotoService.GetByID(ctx.UserContext(), id)
	if err != nil {
		return nil, err
	}
	return p.PhotoService.ConvertToDTO(ctx.UserContext(), photo), nil
}

func (p *PhotoHandler) CreatePhoto(ctx *fiber.Ctx) (interface{}, error) {
	photoParam := &param.Photo{}
	err := ctx.ShouldBindJSON(photoParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	photo, err := p.PhotoService.Create(ctx.UserContext(), photoParam)
	if err != nil {
		return nil, err
	}
	return p.PhotoService.ConvertToDTO(ctx.UserContext(), photo), nil
}

func (p *PhotoHandler) CreatePhotoBatch(ctx *fiber.Ctx) (interface{}, error) {
	photosParam := make([]*param.Photo, 0)
	err := util.BindAndValidate(ctx, &photosParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	photos, err := p.PhotoService.CreateBatch(ctx.UserContext(), photosParam)
	if err != nil {
		return nil, err
	}
	return p.PhotoService.ConvertToDTOs(ctx.UserContext(), photos), nil
}

func (p *PhotoHandler) UpdatePhoto(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	photoParam := &param.Photo{}
	err = ctx.ShouldBindJSON(photoParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	photo, err := p.PhotoService.Update(ctx.UserContext(), id, photoParam)
	if err != nil {
		return nil, err
	}
	return p.PhotoService.ConvertToDTO(ctx.UserContext(), photo), nil
}

func (p *PhotoHandler) DeletePhoto(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return nil, p.PhotoService.Delete(ctx.UserContext(), id)
}

func (p *PhotoHandler) DeletePhotoBatch(ctx *fiber.Ctx) (interface{}, error) {
	photosParam := make([]int32, 0)
	err := util.BindAndValidate(ctx, &photosParam)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	for _, id := range photosParam {
		err := p.PhotoService.Delete(ctx.UserContext(), id)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (p *PhotoHandler) ListPhotoTeams(ctx *fiber.Ctx) (interface{}, error) {
	return p.PhotoService.ListTeams(ctx.UserContext())
}

