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

type PostHandler struct {
	OptionService        service.OptionService
	PostService          service.PostService
	PostCommentService   service.PostCommentService
	PostCommentAssembler assembler.PostCommentAssembler
}

func NewPostHandler(
	optionService service.OptionService,
	postService service.PostService,
	postCommentService service.PostCommentService,
	postCommentAssembler assembler.PostCommentAssembler,
) *PostHandler {
	return &PostHandler{
		OptionService:        optionService,
		PostService:          postService,
		PostCommentService:   postCommentService,
		PostCommentAssembler: postCommentAssembler,
	}
}

func (p *PostHandler) ListTopComment(ctx *fiber.Ctx) (interface{}, error) {
	postID, err := util.ParamInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	pageSize := p.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentPageSize).(int)

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
	commentQuery.ContentID = &postID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, totalCount, err := p.PostCommentService.Page(ctx.UserContext(), commentQuery, consts.CommentTypePost)
	if err != nil {
		return nil, err
	}
	_ = p.PostCommentAssembler.ClearSensitiveField(ctx.UserContext(), comments)
	commenVOs, err := p.PostCommentAssembler.ConvertToWithHasChildren(ctx.UserContext(), comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commenVOs, totalCount, commentQuery.Page), nil
}

func (p *PostHandler) ListChildren(ctx *fiber.Ctx) (interface{}, error) {
	postID, err := util.ParamInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	parentID, err := util.ParamInt32(ctx, "parentID")
	if err != nil {
		return nil, err
	}
	children, err := p.PostCommentService.GetChildren(ctx.UserContext(), parentID, postID, consts.CommentTypePost)
	if err != nil {
		return nil, err
	}
	_ = p.PostCommentAssembler.ClearSensitiveField(ctx.UserContext(), children)
	return p.PostCommentAssembler.ConvertToDTOList(ctx.UserContext(), children)
}

func (p *PostHandler) ListCommentTree(ctx *fiber.Ctx) (interface{}, error) {
	postID, err := util.ParamInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	pageSize := p.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentPageSize).(int)

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
	commentQuery.ContentID = &postID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	allComments, err := p.PostCommentService.GetByContentID(ctx.UserContext(), postID, consts.CommentTypePost, commentQuery.Sort)
	if err != nil {
		return nil, err
	}
	_ = p.PostCommentAssembler.ClearSensitiveField(ctx.UserContext(), allComments)
	commentVOs, total, err := p.PostCommentAssembler.PageConvertToVOs(ctx.UserContext(), allComments, commentQuery.Page)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentVOs, total, commentQuery.Page), nil
}

func (p *PostHandler) ListComment(ctx *fiber.Ctx) (interface{}, error) {
	postID, err := util.ParamInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	pageSize := p.OptionService.GetOrByDefault(ctx.UserContext(), property.CommentPageSize).(int)

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
	commentQuery.ContentID = &postID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, total, err := p.PostCommentService.Page(ctx.UserContext(), commentQuery, consts.CommentTypePost)
	if err != nil {
		return nil, err
	}
	_ = p.PostCommentAssembler.ClearSensitiveField(ctx.UserContext(), comments)
	result, err := p.PostCommentAssembler.ConvertToWithParentVO(ctx.UserContext(), comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(result, total, commentQuery.Page), nil
}

func (p *PostHandler) CreateComment(ctx *fiber.Ctx) (interface{}, error) {
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
	comment.CommentType = consts.CommentTypePost
	result, err := p.PostCommentService.CreateBy(ctx.UserContext(), &comment)
	if err != nil {
		return nil, err
	}
	return p.PostCommentAssembler.ConvertToDTO(ctx.UserContext(), result)
}

func (p *PostHandler) Like(ctx *fiber.Ctx) (interface{}, error) {
	postID, err := util.ParamInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	return nil, p.PostService.IncreaseLike(ctx.UserContext(), postID)
}
