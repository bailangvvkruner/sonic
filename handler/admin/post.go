package admin

import (
	"errors"
	"net/http"
	"strconv"

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

type PostHandler struct {
	PostService   service.PostService
	PostAssembler assembler.PostAssembler
}

func NewPostHandler(postService service.PostService, postAssembler assembler.PostAssembler) *PostHandler {
	return &PostHandler{
		PostService:   postService,
		PostAssembler: postAssembler,
	}
}

func (p *PostHandler) ListPosts(ctx *fiber.Ctx) (interface{}, error) {
	postQuery := param.PostQuery{}
	err := ctx.QueryParser(&postQuery)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if postQuery.Sort == nil {
		postQuery.Sort = &param.Sort{Fields: []string{"topPriority,desc", "createTime,desc"}}
	}
	posts, totalCount, err := p.PostService.Page(ctx.UserContext(), postQuery)
	if err != nil {
		return nil, err
	}
	if postQuery.More == nil || *postQuery.More {
		postVOs, err := p.PostAssembler.ConvertToListVO(ctx.UserContext(), posts)
		return dto.NewPage(postVOs, totalCount, postQuery.Pagination), err
	}
	postDTOs := make([]*dto.Post, 0)
	for _, post := range posts {
		postDTO, err := p.PostAssembler.ConvertToSimpleDTO(ctx.UserContext(), post)
		if err != nil {
			return nil, err
		}
		postDTOs = append(postDTOs, postDTO)
	}
	return dto.NewPage(postDTOs, totalCount, postQuery.Pagination), nil
}

func (p *PostHandler) ListLatestPosts(ctx *fiber.Ctx) (interface{}, error) {
	top, err := util.MustGetQueryInt32(ctx, "top")
	if err != nil {
		top = 10
	}
	postQuery := param.PostQuery{
		Page: param.Pagination{
			PageSize: int(top),
			PageNum:  0,
		},
		Sort: &param.Sort{
			Fields: []string{"createTime,desc"},
		},
		Keyword:    nil,
		CategoryID: nil,
		More:       util.BoolPtr(false),
	}
	posts, _, err := p.PostService.Page(ctx.UserContext(), postQuery)
	if err != nil {
		return nil, err
	}
	postMinimals := make([]*dto.PostMinimal, 0, len(posts))

	for _, post := range posts {
		postMinimal, err := p.PostAssembler.ConvertToMinimalDTO(ctx.UserContext(), post)
		if err != nil {
			return nil, err
		}
		postMinimals = append(postMinimals, postMinimal)
	}
	return postMinimals, nil
}

func (p *PostHandler) ListPostsByStatus(ctx *fiber.Ctx) (interface{}, error) {
	var postQuery param.PostQuery
	err := ctx.QueryParser(&postQuery)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if postQuery.Sort == nil {
		postQuery.Sort = &param.Sort{Fields: []string{"createTime,desc"}}
	}

	status, err := util.ParamInt32(ctx, "status")
	if err != nil {
		return nil, err
	}
	postQuery.Statuses = make([]*consts.PostStatus, 0)
	statusType := consts.PostStatus(status)
	postQuery.Statuses = append(postQuery.Statuses, &statusType)

	posts, totalCount, err := p.PostService.Page(ctx.UserContext(), postQuery)
	if err != nil {
		return nil, err
	}
	if postQuery.More == nil {
		*postQuery.More = false
	}
	if postQuery.More == nil {
		postVOs, err := p.PostAssembler.ConvertToListVO(ctx.UserContext(), posts)
		return dto.NewPage(postVOs, totalCount, postQuery.Pagination), err
	}

	postDTOs := make([]*dto.Post, 0)
	for _, post := range posts {
		postDTO, err := p.PostAssembler.ConvertToSimpleDTO(ctx.UserContext(), post)
		if err != nil {
			return nil, err
		}
		postDTOs = append(postDTOs, postDTO)
	}

	return dto.NewPage(postDTOs, totalCount, postQuery.Pagination), nil
}

func (p *PostHandler) GetByPostID(ctx *fiber.Ctx) (interface{}, error) {
	postIDStr := ctx.Params("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 32)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	post, err := p.PostService.GetByPostID(ctx.UserContext(), int32(postID))
	if err != nil {
		return nil, err
	}
	postDetailVO, err := p.PostAssembler.ConvertToDetailVO(ctx.UserContext(), post)
	if err != nil {
		return nil, err
	}
	return postDetailVO, nil
}

func (p *PostHandler) CreatePost(ctx *fiber.Ctx) (interface{}, error) {
	var postParam param.Post
	err := util.BindAndValidate(ctx, &postParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}

	post, err := p.PostService.Create(ctx.UserContext(), &postParam)
	if err != nil {
		return nil, err
	}
	return p.PostAssembler.ConvertToDetailVO(ctx.UserContext(), post)
}

func (p *PostHandler) UpdatePost(ctx *fiber.Ctx) (interface{}, error) {
	var postParam param.Post
	err := util.BindAndValidate(ctx, &postParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}

	postIDStr := ctx.Params("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 32)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}

	postDetailVO, err := p.PostService.Update(ctx.UserContext(), int32(postID), &postParam)
	if err != nil {
		return nil, err
	}
	return postDetailVO, nil
}

func (p *PostHandler) UpdatePostStatus(ctx *fiber.Ctx) (interface{}, error) {
	postID, err := util.ParamInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	statusStr, err := util.ParamString(ctx, "status")
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	status, err := consts.PostStatusFromString(statusStr)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if int32(status) < int32(consts.PostStatusPublished) || int32(status) > int32(consts.PostStatusIntimate) {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("status error")
	}
	post, err := p.PostService.UpdateStatus(ctx.UserContext(), postID, status)
	if err != nil {
		return nil, err
	}
	return p.PostAssembler.ConvertToMinimalDTO(ctx.UserContext(), post)
}

func (p *PostHandler) UpdatePostStatusBatch(ctx *fiber.Ctx) (interface{}, error) {
	statusStr, err := util.ParamString(ctx, "status")
	if err != nil {
		return nil, err
	}
	status, err := consts.PostStatusFromString(statusStr)
	if err != nil {
		return nil, err
	}
	if int32(status) < int32(consts.PostStatusPublished) || int32(status) > int32(consts.PostStatusIntimate) {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("status error")
	}
	ids := make([]int32, 0)
	err = ctx.BodyParser(&ids)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("post ids error")
	}

	return p.PostService.UpdateStatusBatch(ctx.UserContext(), status, ids)
}

func (p *PostHandler) UpdatePostDraft(ctx *fiber.Ctx) (interface{}, error) {
	postID, err := util.ParamInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	var postContentParam param.PostContent
	err = util.BindAndValidate(ctx, &postContentParam)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("content param error")
	}
	post, err := p.PostService.UpdateDraftContent(ctx.UserContext(), postID, postContentParam.Content, postContentParam.OriginalContent)
	if err != nil {
		return nil, err
	}
	return p.PostAssembler.ConvertToDetailDTO(ctx.UserContext(), post)
}

func (p *PostHandler) DeletePost(ctx *fiber.Ctx) (interface{}, error) {
	postID, err := util.ParamInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	return nil, p.PostService.Delete(ctx.UserContext(), postID)
}

func (p *PostHandler) DeletePostBatch(ctx *fiber.Ctx) (interface{}, error) {
	postIDs := make([]int32, 0)
	err := ctx.BodyParser(&postIDs)
	if err != nil {
		return nil, xerr.WithMsg(err, "postIDs error").WithStatus(xerr.StatusBadRequest)
	}
	return nil, p.PostService.DeleteBatch(ctx.UserContext(), postIDs)
}

func (p *PostHandler) PreviewPost(ctx *fiber.Ctx) error {
	postID, err := util.ParamInt32(ctx, "postID")
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return err
	}
	previewPath, err := p.PostService.Preview(ctx.UserContext(), postID)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return err
	}
	return ctx.Status(http.StatusOK).SendString(previewPath)
}

