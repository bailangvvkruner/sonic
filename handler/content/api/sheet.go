package api

import (
	"html/template"

	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/service/assembler"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type SheetHandler struct {
	OptionService         service.OptionService
	SheetService          service.SheetService
	SheetCommentService   service.SheetCommentService
	SheetCommentAssembler assembler.SheetCommentAssembler
}

func NewSheetHandler(
	optionService service.OptionService,
	sheetService service.SheetService,
	sheetCommentService service.SheetCommentService,
	sheetCommentAssembler assembler.SheetCommentAssembler,
) *SheetHandler {
	return &SheetHandler{
		OptionService:         optionService,
		SheetService:          sheetService,
		SheetCommentService:   sheetCommentService,
		SheetCommentAssembler: sheetCommentAssembler,
	}
}

func (s *SheetHandler) ListTopComment(ctx *fiber.Ctx) (interface{}, error) {
	sheetID, err := util.ParamInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	pageSize := s.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentPageSize).(int)

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
	commentQuery.ContentID = &sheetID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, totalCount, err := s.SheetCommentService.Page(ctx.UserContext(), commentQuery, consts.CommentTypeSheet)
	if err != nil {
		return nil, err
	}
	_ = s.SheetCommentAssembler.ClearSensitiveField(ctx.UserContext(), comments)
	commenVOs, err := s.SheetCommentAssembler.ConvertToWithHasChildren(ctx.UserContext(), comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commenVOs, totalCount, commentQuery.Pagination), nil
}

func (s *SheetHandler) ListChildren(ctx *fiber.Ctx) (interface{}, error) {
	sheetID, err := util.ParamInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	parentID, err := util.ParamInt32(ctx, "parentID")
	if err != nil {
		return nil, err
	}
	children, err := s.SheetCommentService.GetChildren(ctx.UserContext(), parentID, sheetID, consts.CommentTypeSheet)
	if err != nil {
		return nil, err
	}
	_ = s.SheetCommentAssembler.ClearSensitiveField(ctx.UserContext(), children)
	return s.SheetCommentAssembler.ConvertToDTOList(ctx.UserContext(), children)
}

func (s *SheetHandler) ListCommentTree(ctx *fiber.Ctx) (interface{}, error) {
	sheetID, err := util.ParamInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	pageSize := s.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentPageSize).(int)

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
	commentQuery.ContentID = &sheetID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	allComments, err := s.SheetCommentService.GetByContentID(ctx.UserContext(), sheetID, consts.CommentTypeSheet, commentQuery.Sort)
	if err != nil {
		return nil, err
	}
	_ = s.SheetCommentAssembler.ClearSensitiveField(ctx.UserContext(), allComments)
	commentVOs, total, err := s.SheetCommentAssembler.PageConvertToVOs(ctx.UserContext(), allComments, commentQuery.Pagination)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentVOs, total, commentQuery.Pagination), nil
}

func (s *SheetHandler) ListComment(ctx *fiber.Ctx) (interface{}, error) {
	sheetID, err := util.ParamInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	pageSize := s.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentPageSize).(int)

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
	commentQuery.ContentID = &sheetID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, total, err := s.SheetCommentService.Page(ctx.UserContext(), commentQuery, consts.CommentTypeSheet)
	if err != nil {
		return nil, err
	}
	_ = s.SheetCommentAssembler.ClearSensitiveField(ctx.UserContext(), comments)
	result, err := s.SheetCommentAssembler.ConvertToWithParentVO(ctx.UserContext(), comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(result, total, commentQuery.Pagination), nil
}

func (s *SheetHandler) CreateComment(ctx *fiber.Ctx) (interface{}, error) {
	comment := param.Comment{}
	err := util.BindAndValidate(ctx, &comment)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if comment.AuthorURL != "" {
		err = util.Validate.Var(comment.AuthorURL, "http_url")
		if err != nil {
			return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
		}
	}
	comment.Author = template.HTMLEscapeString(comment.Author)
	comment.AuthorURL = template.HTMLEscapeString(comment.AuthorURL)
	comment.Content = template.HTMLEscapeString(comment.Content)
	comment.Email = template.HTMLEscapeString(comment.Email)
	comment.CommentType = consts.CommentTypeSheet
	result, err := s.SheetCommentService.CreateBy(ctx.UserContext(), &comment)
	if err != nil {
		return nil, err
	}
	return s.SheetCommentAssembler.ConvertToDTO(ctx.UserContext(), result)
}
