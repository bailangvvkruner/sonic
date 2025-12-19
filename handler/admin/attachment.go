package admin

import (
	"github.com/gofiber/fiber/v2"


	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type AttachmentHandler struct {
	AttachmentService service.AttachmentService
}

func NewAttachmentHandler(attachmentService service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{
		AttachmentService: attachmentService,
	}
}

func (a *AttachmentHandler) QueryAttachment(ctx *fiber.Ctx) (interface{}, error) {
	queryParam := &param.AttachmentQuery{}
	err := ctx.QueryParser(queryParam)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("param error ")
	}
	attachments, totalCount, err := a.AttachmentService.Page(ctx.UserContext(), queryParam)
	if err != nil {
		return nil, err
	}
	attachmentDTOs, err := a.AttachmentService.ConvertToDTOs(ctx.UserContext(), attachments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(attachmentDTOs, totalCount, queryParam.Pagination), nil
}

func (a *AttachmentHandler) GetAttachmentByID(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	if id < 0 {
		return nil, xerr.BadParam.New("id < 0").WithStatus(xerr.StatusBadRequest).WithMsg("param error")
	}
	return a.AttachmentService.GetAttachment(ctx.UserContext(), id)
}

func (a *AttachmentHandler) UploadAttachment(ctx *fiber.Ctx) (interface{}, error) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return nil, xerr.WithMsg(err, "上传文件错误").WithStatus(xerr.StatusBadRequest)
	}
	return a.AttachmentService.Upload(ctx.UserContext(), fileHeader)
}

func (a *AttachmentHandler) UploadAttachments(ctx *fiber.Ctx) (interface{}, error) {
	form, _ := ctx.MultipartForm()
	if len(form.File) == 0 {
		return nil, xerr.BadParam.New("empty files").WithStatus(xerr.StatusBadRequest).WithMsg("empty files")
	}
	files := form.File["files"]
	attachmentDTOs := make([]*dto.AttachmentDTO, 0)
	for _, file := range files {
		attachment, err := a.AttachmentService.Upload(ctx.UserContext(), file)
		if err != nil {
			return nil, err
		}
		attachmentDTOs = append(attachmentDTOs, attachment)
	}
	return attachmentDTOs, nil
}

func (a *AttachmentHandler) UpdateAttachment(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}

	updateParam := &param.AttachmentUpdate{}
	err = ctx.BodyParser(updateParam)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("param error ")
	}
	return a.AttachmentService.Update(ctx.UserContext(), id, updateParam)
}

func (a *AttachmentHandler) DeleteAttachment(ctx *fiber.Ctx) (interface{}, error) {
	id, err := util.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return a.AttachmentService.Delete(ctx.UserContext(), id)
}

func (a *AttachmentHandler) DeleteAttachmentInBatch(ctx *fiber.Ctx) (interface{}, error) {
	ids := make([]int32, 0)
	err := ctx.BodyParser(&ids)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	return a.AttachmentService.DeleteBatch(ctx.UserContext(), ids)
}

func (a *AttachmentHandler) GetAllMediaType(ctx *fiber.Ctx) (interface{}, error) {
	return a.AttachmentService.GetAllMediaTypes(ctx.UserContext())
}

func (a *AttachmentHandler) GetAllTypes(ctx *fiber.Ctx) (interface{}, error) {
	attachmentTypes, err := a.AttachmentService.GetAllTypes(ctx.UserContext())
	if err != nil {
		return nil, err
	}
	return attachmentTypes, nil
}

