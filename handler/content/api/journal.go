package api

import (
	"html/template"

	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/entity"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/service/assembler"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type JournalHandler struct {
	JournalService          service.JournalService
	JournalCommentService   service.JournalCommentService
	OptionService           service.ClientOptionService
	JournalCommentAssembler assembler.JournalCommentAssembler
}

func NewJournalHandler(
	journalService service.JournalService,
	journalCommentService service.JournalCommentService,
	optionService service.ClientOptionService,
	journalCommentAssembler assembler.JournalCommentAssembler,
) *JournalHandler {
	return &JournalHandler{
		JournalService:          journalService,
		JournalCommentService:   journalCommentService,
		OptionService:           optionService,
		JournalCommentAssembler: journalCommentAssembler,
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
	journalQuery.JournalType = consts.JournalTypePublic.Ptr()
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

func (j *JournalHandler) GetJournal(ctx *fiber.Ctx) (interface{}, error) {
	journalID, err := util.ParamInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	journals, err := j.JournalService.GetByJournalIDs(ctx.UserContext(), []int32{journalID})
	if err != nil {
		return nil, err
	}
	if len(journals) == 0 {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest)
	}
	journalDTOs, err := j.JournalService.ConvertToWithCommentDTOList(ctx.UserContext(), []*entity.Journal{journals[journalID]})
	if err != nil {
		return nil, err
	}
	return journalDTOs[0], nil
}

func (j *JournalHandler) ListTopComment(ctx *fiber.Ctx) (interface{}, error) {
	journalID, err := util.ParamInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	pageSize := j.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentPageSize).(int)

	commentQuery := param.CommentQuery{}
	err = ctx.QueryParser(&commentQuery)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if commentQuery.Sort != nil && len(commentQuery.Fields) > 0 {
		commentQuery.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	commentQuery.ContentID = &journalID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, totalCount, err := j.JournalCommentService.Page(ctx.UserContext(), commentQuery, consts.CommentTypeJournal)
	if err != nil {
		return nil, err
	}
	_ = j.JournalCommentAssembler.ClearSensitiveField(ctx.UserContext(), comments)
	commenVOs, err := j.JournalCommentAssembler.ConvertToWithHasChildren(ctx.UserContext(), comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commenVOs, totalCount, commentQuery.Pagination), nil
}

func (j *JournalHandler) ListChildren(ctx *fiber.Ctx) (interface{}, error) {
	journalID, err := util.ParamInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	parentID, err := util.ParamInt32(ctx, "parentID")
	if err != nil {
		return nil, err
	}
	children, err := j.JournalCommentService.GetChildren(ctx.UserContext(), parentID, journalID, consts.CommentTypeJournal)
	if err != nil {
		return nil, err
	}
	_ = j.JournalCommentAssembler.ClearSensitiveField(ctx.UserContext(), children)
	return j.JournalCommentAssembler.ConvertToDTOList(ctx.UserContext(), children)
}

func (j *JournalHandler) ListCommentTree(ctx *fiber.Ctx) (interface{}, error) {
	journalID, err := util.ParamInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	pageSize := j.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentPageSize).(int)

	commentQuery := param.CommentQuery{}
	err = ctx.QueryParser(&commentQuery)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if commentQuery.Sort != nil && len(commentQuery.Fields) > 0 {
		commentQuery.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	commentQuery.ContentID = &journalID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	allComments, err := j.JournalCommentService.GetByContentID(ctx.UserContext(), journalID, consts.CommentTypeJournal, commentQuery.Sort)
	if err != nil {
		return nil, err
	}
	_ = j.JournalCommentAssembler.ClearSensitiveField(ctx.UserContext(), allComments)
	commentVOs, total, err := j.JournalCommentAssembler.PageConvertToVOs(ctx.UserContext(), allComments, commentQuery.Pagination)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentVOs, total, commentQuery.Pagination), nil
}

func (j *JournalHandler) ListComment(ctx *fiber.Ctx) (interface{}, error) {
	journalID, err := util.ParamInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	pageSize := j.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentPageSize).(int)

	commentQuery := param.CommentQuery{}
	err = ctx.QueryParser(&commentQuery)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if commentQuery.Sort != nil && len(commentQuery.Fields) > 0 {
		commentQuery.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	commentQuery.ContentID = &journalID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, total, err := j.JournalCommentService.Page(ctx.UserContext(), commentQuery, consts.CommentTypeJournal)
	if err != nil {
		return nil, err
	}
	_ = j.JournalCommentAssembler.ClearSensitiveField(ctx.UserContext(), comments)
	result, err := j.JournalCommentAssembler.ConvertToWithParentVO(ctx.UserContext(), comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(result, total, commentQuery.Pagination), nil
}

func (j *JournalHandler) CreateComment(ctx *fiber.Ctx) (interface{}, error) {
	p := param.Comment{}
	err := util.BindAndValidate(ctx, &p)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if p.AuthorURL != "" {
		err = util.Validate.Var(p.AuthorURL, "http_url")
		if err != nil {
			return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
		}
	}
	p.Author = template.HTMLEscapeString(p.Author)
	p.AuthorURL = template.HTMLEscapeString(p.AuthorURL)
	p.Content = template.HTMLEscapeString(p.Content)
	p.Email = template.HTMLEscapeString(p.Email)
	p.CommentType = consts.CommentTypeJournal
	result, err := j.JournalCommentService.CreateBy(ctx.UserContext(), &p)
	if err != nil {
		return nil, err
	}
	return j.JournalCommentAssembler.ConvertToDTO(ctx.UserContext(), result)
}

func (j *JournalHandler) Like(ctx *fiber.Ctx) (interface{}, error) {
	journalID, err := util.ParamInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	err = j.JournalService.IncreaseLike(ctx.UserContext(), journalID)
	if err != nil {
		return nil, err
	}
	return nil, err
}
