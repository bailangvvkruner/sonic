package model

import (
	"context"
	"strings"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/handler/content/authentication"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/entity"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/model/vo"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/service/assembler"
	"github.com/go-sonic/sonic/service/impl"
	"github.com/go-sonic/sonic/template"
	"github.com/go-sonic/sonic/util/xerr"
)

func NewPostModel(optionService service.OptionService,
	postService service.PostService,
	themeService service.ThemeService,
	postCategoryService service.PostCategoryService,
	categoryService service.CategoryService,
	postTagService service.PostTagService,
	tagService service.TagService,
	postAssembler assembler.PostAssembler,
	metaService service.MetaService,
	postAuthentication *authentication.PostAuthentication,
	cachedPostService *impl.CachedPostService,
) *PostModel {
	return &PostModel{
		OptionService:       optionService,
		PostService:         postService,
		PostAssembler:       postAssembler,
		ThemeService:        themeService,
		PostCategoryService: postCategoryService,
		CategoryService:     categoryService,
		PostTagService:      postTagService,
		TagService:          tagService,
		MetaService:         metaService,
		PostAuthentication:  postAuthentication,
		CachedPostService:   cachedPostService,
	}
}

type PostModel struct {
	OptionService       service.OptionService
	PostService         service.PostService
	ThemeService        service.ThemeService
	PostCategoryService service.PostCategoryService
	CategoryService     service.CategoryService
	PostTagService      service.PostTagService
	TagService          service.TagService
	MetaService         service.MetaService
	PostAssembler       assembler.PostAssembler
	PostAuthentication  *authentication.PostAuthentication
	CachedPostService   *impl.CachedPostService
}

func (p *PostModel) Content(ctx context.Context, post *entity.Post, token string, model template.Model) (string, error) {
	if post == nil {
		return "", xerr.WithStatus(nil, int(xerr.StatusBadRequest)).WithMsg("查询不到文章信息")
	}
	if post.Status == consts.PostStatusRecycle || post.Status == consts.PostStatusDraft {
		return "", xerr.WithStatus(nil, xerr.StatusNotFound).WithMsg("查询不到文章信息")
	} else if post.Status == consts.PostStatusIntimate {
		if isAuthenticated, err := p.PostAuthentication.IsAuthenticated(ctx, token, post.ID); err != nil || !isAuthenticated {
			model["slug"] = post.Slug
			model["type"] = consts.EncryptTypePost.Name()
			if exist, err := p.ThemeService.TemplateExist(ctx, "post_password.tmpl"); err == nil && exist {
				return p.ThemeService.Render(ctx, "post_password")
			}
			return "common/template/post_password", nil
		}
	}

	// 使用缓存获取文章详情
	postVO, err := p.CachedPostService.GetPostDetailVO(ctx, post)
	if err != nil {
		return "", err
	}
	model["post"] = postVO

	// 使用缓存获取上下篇文章
	prevPost, nextPost, err := p.CachedPostService.GetPrevNextPosts(ctx, post)
	if err != nil {
		return "", err
	}
	if prevPost != nil {
		model["prevPost"] = prevPost
	}
	if nextPost != nil {
		model["nextPost"] = nextPost
	}

	// 使用缓存获取文章分类
	categories, err := p.CachedPostService.GetPostCategories(ctx, post.ID)
	if err != nil {
		return "", err
	}
	model["categories"], _ = p.CategoryService.ConvertToCategoryDTOs(ctx, categories)

	// 使用缓存获取文章标签
	tags, err := p.CachedPostService.GetPostTags(ctx, post.ID)
	if err != nil {
		return "", err
	}
	model["tags"], _ = p.TagService.ConvertToDTOs(ctx, tags)

	// 使用缓存获取文章 Meta
	metas, err := p.CachedPostService.GetPostMeta(ctx, post.ID)
	if err != nil {
		return "", err
	}
	model["metas"] = p.MetaService.ConvertToMetaDTOs(metas)

	if post.MetaDescription != "" {
		model["meta_description"] = post.MetaDescription
	} else {
		model["meta_description"] = post.Summary
	}
	if post.MetaKeywords != "" {
		model["meta_keywords"] = post.MetaKeywords
	} else if len(tags) > 0 {
		metaKeywords := strings.Builder{}
		metaKeywords.Write([]byte(tags[0].Name))
		for _, tag := range tags[1:] {
			metaKeywords.Write([]byte(","))
			metaKeywords.Write([]byte(tag.Name))
		}
		model["meta_keywords"] = metaKeywords.String()
	}
	model["is_post"] = true

	p.PostService.IncreaseVisit(ctx, post.ID)

	model["target"] = postVO
	model["type"] = "post"
	return p.ThemeService.Render(ctx, "post")
}

func (p *PostModel) List(ctx context.Context, page int, model template.Model) (string, error) {
	pageSize := p.OptionService.GetIndexPageSize(ctx)
	sort := p.OptionService.GetPostSort(ctx)
	postQuery := param.PostQuery{
		Pagination: param.Pagination{
			PageNum:  page,
			PageSize: pageSize,
		},
		Sort:     sort,
		Statuses: []consts.PostStatus{consts.PostStatusPublished},
	}
	posts, totalCount, err := p.PostService.Page(ctx, postQuery)
	if err != nil {
		return "", err
	}

	// 批量获取文章 ID
	postIDs := make([]int32, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}

	// 使用缓存批量获取标签和分类
	tagMap, _ := p.CachedPostService.GetPostTagsMap(ctx, postIDs)
	catMap, _ := p.CachedPostService.GetPostCategoriesMap(ctx, postIDs)

	// 转换为 VO
	postVOs, err := p.convertToListVOWithCache(ctx, posts, tagMap, catMap)
	if err != nil {
		return "", err
	}

	postPage := dto.NewPage(postVOs, totalCount, param.Pagination{
		PageNum:  page,
		PageSize: pageSize,
	})
	seoKeyWords := p.OptionService.GetOrByDefault(ctx, property.SeoKeywords)
	seoDescription := p.OptionService.GetOrByDefault(ctx, property.SeoDescription)

	model["is_index"] = true
	model["posts"] = postPage
	model["meta_keywords"] = seoKeyWords
	model["meta_description"] = seoDescription
	return p.ThemeService.Render(ctx, "index")
}

// convertToListVOWithCache 使用缓存数据转换文章列表 VO
func (p *PostModel) convertToListVOWithCache(
	ctx context.Context,
	posts []*entity.Post,
	tagMap map[int32][]*entity.Tag,
	catMap map[int32][]*entity.Category,
) ([]*vo.Post, error) {
	postVOs := make([]*vo.Post, 0, len(posts))

	for _, post := range posts {
		postVO := &vo.Post{}

		// 从缓存获取分类
		if categories, ok := catMap[post.ID]; ok {
			categoryDTOs, err := p.CategoryService.ConvertToCategoryDTOs(ctx, categories)
			if err != nil {
				return nil, err
			}
			postVO.Categories = categoryDTOs
		}

		// 从缓存获取标签
		if tags, ok := tagMap[post.ID]; ok {
			tagDTOs, err := p.TagService.ConvertToDTOs(ctx, tags)
			if err != nil {
				return nil, err
			}
			postVO.Tags = tagDTOs
		}

		// 转换基础 DTO
		postDTO, err := p.PostAssembler.ConvertToSimpleDTO(ctx, post)
		if err != nil {
			return nil, err
		}
		postVO.Post = *postDTO
		postVOs = append(postVOs, postVO)
	}

	return postVOs, nil
}

func (p *PostModel) Archives(ctx context.Context, page int, model template.Model) (string, error) {
	pageSize := p.OptionService.GetOrByDefault(ctx, property.ArchivePageSize).(int)
	postQuery := param.PostQuery{
		Pagination: param.Pagination{
			PageNum:  page,
			PageSize: pageSize,
		},
		Sort: param.Sort{
			Fields: []string{"createTime,desc"},
		},
		Statuses: []consts.PostStatus{consts.PostStatusPublished},
	}
	posts, totalPage, err := p.PostService.Page(ctx, postQuery)
	if err != nil {
		return "", err
	}

	// 使用缓存批量获取标签和分类
	postIDs := make([]int32, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}
	tagMap, _ := p.CachedPostService.GetPostTagsMap(ctx, postIDs)
	catMap, _ := p.CachedPostService.GetPostCategoriesMap(ctx, postIDs)

	postVOs, err := p.convertToListVOWithCache(ctx, posts, tagMap, catMap)
	if err != nil {
		return "", err
	}

	postPage := dto.NewPage(postVOs, totalPage, param.Pagination{
		PageNum:  page,
		PageSize: pageSize,
	})
	archives, err := p.PostAssembler.ConvertToArchiveYearVOs(ctx, posts)
	if err != nil {
		return "", err
	}
	seoKeyWords := p.OptionService.GetOrByDefault(ctx, property.SeoKeywords)
	seoDescription := p.OptionService.GetOrByDefault(ctx, property.SeoDescription)

	model["is_archives"] = true
	model["posts"] = postPage
	model["archives"] = archives
	model["meta_keywords"] = seoKeyWords
	model["meta_description"] = seoDescription
	return p.ThemeService.Render(ctx, "archives")
}

func (p *PostModel) AdminPreview(ctx context.Context, post *entity.Post, model template.Model) (string, error) {
	if post == nil {
		return "", xerr.WithStatus(nil, int(xerr.StatusBadRequest)).WithMsg("查询不到文章信息")
	}

	// 预览不使用缓存，直接查询
	postVO, err := p.PostAssembler.ConvertToDetailVO(ctx, post)
	if err != nil {
		return "", err
	}
	model["post"] = postVO

	prevPosts, err := p.PostService.GetPrevPosts(ctx, post, 1)
	if err != nil {
		return "", err
	}
	nextPosts, err := p.PostService.GetNextPosts(ctx, post, 1)
	if err != nil {
		return "", err
	}
	if len(prevPosts) > 0 {
		prePost, err := p.PostAssembler.ConvertToDetailVO(ctx, prevPosts[0])
		if err != nil {
			return "", err
		}
		model["prevPost"] = prePost
	}
	if len(nextPosts) > 0 {
		nextPost, err := p.PostAssembler.ConvertToDetailVO(ctx, nextPosts[0])
		if err != nil {
			return "", err
		}
		model["nextPost"] = nextPost
	}

	categories, err := p.PostCategoryService.ListCategoryByPostID(ctx, post.ID)
	if err != nil {
		return "", err
	}
	model["categories"], _ = p.CategoryService.ConvertToCategoryDTOs(ctx, categories)

	tags, err := p.PostTagService.ListTagByPostID(ctx, post.ID)
	if err != nil {
		return "", err
	}
	model["tags"], _ = p.TagService.ConvertToDTOs(ctx, tags)

	metas, err := p.MetaService.GetPostMeta(ctx, post.ID)
	if err != nil {
		return "", err
	}
	model["metas"] = p.MetaService.ConvertToMetaDTOs(metas)

	if post.MetaDescription != "" {
		model["meta_description"] = post.MetaDescription
	} else {
		model["meta_description"] = post.Summary
	}
	if post.MetaKeywords != "" {
		model["meta_keywords"] = post.MetaKeywords
	} else if len(tags) > 0 {
		metaKeywords := strings.Builder{}
		metaKeywords.Write([]byte(tags[0].Name))
		for _, tag := range tags[1:] {
			metaKeywords.Write([]byte(","))
			metaKeywords.Write([]byte(tag.Name))
		}
		model["meta_keywords"] = metaKeywords.String()
	}
	model["is_post"] = true

	model["target"] = postVO
	model["type"] = "post"
	return p.ThemeService.Render(ctx, "post")
}
