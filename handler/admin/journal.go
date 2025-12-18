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

type JournalHandler struct {
	JournalService service.JournalService
}

func NewJournalHandler(journalService service.JournalService) *JournalHandler {
	return &JournalHandler{
		JournalService: journalService,
	}
}

func (j *JournalHandler) ListJournal(ctx *fiber.Ctx) (interface{}, error) {
	var journalQuery param.JournalQuery
	err := ctx.QueryParser(&journalQuery)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	journalQuery.Sort = &param.Sort{
		Fields: []string{"createTime,desc"},
	}
	journals, totalCount, err := j.JournalService.ListJournal(ctx.UserContext(), journalQuery)
	if err != nil {
		return nil, err
	}
	journalDTOs, err := j.JournalService.ConvertToWithCommentDTOList(ctx.UserContext(), journals)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(journalDTOs, totalCount, journalQuery.Pagination), nil
}

func (j *JournalHandler) ListLatestJournal(ctx *fiber.Ctx) (interface{}, error) {
	top, err := util.MustGetQueryInt(ctx, "top")
	if err != nil {
		top = 10
	}
	journalQuery := param.JournalQuery{
		Sort: &param.Sort{Fields: []string{"createTime,desc"}},
		Page: param.Pagination{PageNum: 0, PageSize: top},
	}
	journals, _, err := j.JournalService.ListJournal(ctx.UserContext(), journalQuery)
	if err != nil {
		return nil, err
	}
	return j.JournalService.ConvertToWithCommentDTOList(ctx.UserContext(), journals)
}

func (j *JournalHandler) CreateJournal(ctx *fiber.Ctx) (interface{}, error) {
	var journalParam param.Journal
	err := util.BindAndValidate(ctx, &journalParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	if journalParam.Content == "" {
		journalParam.Content = journalParam.SourceContent
	}
	journal, err := j.JournalService.Create(ctx.UserContext(), &journalParam)
	if err != nil {
		return nil, err
	}
	return j.JournalService.ConvertToDTO(journal), nil
}

func (j *JournalHandler) UpdateJournal(ctx *fiber.Ctx) (interface{}, error) {
	var journalParam param.Journal
	err := util.BindAndValidate(ctx, &journalParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}

	journalID, err := util.ParamInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	return j.JournalService.Update(ctx.UserContext(), journalID, &journalParam)
}

func (j *JournalHandler) DeleteJournal(ctx *fiber.Ctx) (interface{}, error) {
	journalID, err := util.ParamInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	return nil, j.JournalService.Delete(ctx.UserContext(), journalID)
}

