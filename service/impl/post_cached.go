package impl

import (
	"context"

	"go.uber.org/zap"

	"github.com/go-sonic/sonic/cache"
	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/log"
	"github.com/go-sonic/sonic/model/entity"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/model/vo"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/service/assembler"
)

// CachedPostService 带缓存的文章服务
type CachedPostService struct {
	PostService       service.PostService
	CacheManager      *cache.CacheManager
	PostAssembler     assembler.PostAssembler
	PostTagService    service.PostTagService
	PostCategoryService service.PostCategoryService
	TagService        service.TagService
	CategoryService   service.CategoryService
	MetaService       service.MetaService
}

// NewCachedPostService 创建带缓存的文章服务
func NewCachedPostService(
	postService service.PostService,
	cacheManager *cache.CacheManager,
	postAssembler assembler.PostAssembler,
	postTagService service.PostTagService,
	postCategoryService service.PostCategoryService,
	tagService service.TagService,
	categoryService service.CategoryService,
	metaService service.MetaService,
) *CachedPostService {
	return &CachedPostService{
		PostService:         postService,
		CacheManager:        cacheManager,
		PostAssembler:       postAssembler,
		PostTagService:      postTagService,
		PostCategoryService: postCategoryService,
		TagService:          tagService,
		CategoryService:     categoryService,
		MetaService:         metaService,
	}
}

// GetPostDetailVO 获取文章详情（带缓存）
func (c *CachedPostService) GetPostDetailVO(ctx context.Context, post *entity.Post) (*vo.PostDetailVO, error) {
	if post == nil {
		return nil, nil
	}

	// 1. 尝试从缓存获取
	if cached, ok := c.CacheManager.GetPostDetail(post.ID); ok {
		log.Info("[CACHE HIT] PostDetail", zap.Int32("postID", post.ID))
		return cached, nil
	}
	log.Info("[CACHE MISS] PostDetail", zap.Int32("postID", post.ID))

	// 2. 缓存未命中，从数据库组装
	postVO, err := c.PostAssembler.ConvertToDetailVO(ctx, post)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	_ = c.CacheManager.SetPostDetail(postVO)
	log.Info("[CACHE SET] PostDetail", zap.Int32("postID", post.ID))

	return postVO, nil
}

// GetPrevNextPosts 获取上下篇文章（带缓存）
func (c *CachedPostService) GetPrevNextPosts(ctx context.Context, post *entity.Post) (prev, next *vo.Post, err error) {
	if post == nil {
		return nil, nil, nil
	}

	// 1. 尝试从缓存获取
	if prevCached, nextCached, ok := c.CacheManager.GetPrevNextPosts(post.ID); ok {
		return prevCached, nextCached, nil
	}

	// 2. 缓存未命中，查询数据库
	prevPosts, err := c.PostService.GetPrevPosts(ctx, post, 1)
	if err != nil {
		return nil, nil, err
	}

	nextPosts, err := c.PostService.GetNextPosts(ctx, post, 1)
	if err != nil {
		return nil, nil, err
	}

	// 3. 转换为 VO
	var prevVO, nextVO *vo.Post
	if len(prevPosts) > 0 {
		prevVOs, err := c.PostAssembler.ConvertToListVO(ctx, []*entity.Post{prevPosts[0]})
		if err != nil {
			return nil, nil, err
		}
		if len(prevVOs) > 0 {
			prevVO = prevVOs[0]
		}
	}

	if len(nextPosts) > 0 {
		nextVOs, err := c.PostAssembler.ConvertToListVO(ctx, []*entity.Post{nextPosts[0]})
		if err != nil {
			return nil, nil, err
		}
		if len(nextVOs) > 0 {
			nextVO = nextVOs[0]
		}
	}

	// 4. 写入缓存
	_ = c.CacheManager.SetPrevNextPosts(post.ID, prevVO, nextVO)

	return prevVO, nextVO, nil
}

// GetPostTags 获取文章标签（带缓存）
func (c *CachedPostService) GetPostTags(ctx context.Context, postID int32) ([]*entity.Tag, error) {
	// 1. 尝试从缓存获取
	if cached, ok := c.CacheManager.GetPostTags(postID); ok {
		return cached, nil
	}

	// 2. 缓存未命中，查询数据库
	tags, err := c.PostTagService.ListTagByPostID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	_ = c.CacheManager.SetPostTags(postID, tags)

	return tags, nil
}

// GetPostCategories 获取文章分类（带缓存）
func (c *CachedPostService) GetPostCategories(ctx context.Context, postID int32) ([]*entity.Category, error) {
	// 1. 尝试从缓存获取
	if cached, ok := c.CacheManager.GetPostCategories(postID); ok {
		return cached, nil
	}

	// 2. 缓存未命中，查询数据库
	categories, err := c.PostCategoryService.ListCategoryByPostID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	_ = c.CacheManager.SetPostCategories(postID, categories)

	return categories, nil
}

// GetPostMeta 获取文章 Meta（带缓存）
func (c *CachedPostService) GetPostMeta(ctx context.Context, postID int32) ([]*entity.Meta, error) {
	// 1. 尝试从缓存获取
	if cached, ok := c.CacheManager.GetPostMeta(postID); ok {
		return cached, nil
	}

	// 2. 缓存未命中，查询数据库
	metas, err := c.MetaService.GetPostMeta(ctx, postID)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	_ = c.CacheManager.SetPostMeta(postID, metas)

	return metas, nil
}

// GetPostTagsMap 批量获取文章标签映射（带缓存）
func (c *CachedPostService) GetPostTagsMap(ctx context.Context, postIDs []int32) (map[int32][]*entity.Tag, error) {
	if len(postIDs) == 0 {
		return make(map[int32][]*entity.Tag), nil
	}

	// 1. 尝试从缓存获取
	if cached, ok := c.CacheManager.GetPostTagsMap(postIDs); ok {
		return cached, nil
	}

	// 2. 缓存未命中，查询数据库
	tagMap, err := c.PostTagService.ListTagMapByPostID(ctx, postIDs)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	_ = c.CacheManager.SetPostTagsMap(postIDs, tagMap)

	return tagMap, nil
}

// GetPostCategoriesMap 批量获取文章分类映射（带缓存）
func (c *CachedPostService) GetPostCategoriesMap(ctx context.Context, postIDs []int32) (map[int32][]*entity.Category, error) {
	if len(postIDs) == 0 {
		return make(map[int32][]*entity.Category), nil
	}

	// 1. 尝试从缓存获取
	if cached, ok := c.CacheManager.GetPostCategoriesMap(postIDs); ok {
		return cached, nil
	}

	// 2. 缓存未命中，查询数据库
	catMap, err := c.PostCategoryService.ListCategoryMapByPostID(ctx, postIDs)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	_ = c.CacheManager.SetPostCategoriesMap(postIDs, catMap)

	return catMap, nil
}

// InvalidatePostCache 清除文章缓存
func (c *CachedPostService) InvalidatePostCache(postID int32) {
	c.CacheManager.InvalidatePostCache(postID)
}

// InvalidateAllPostList 清除所有文章列表缓存
func (c *CachedPostService) InvalidateAllPostList() {
	c.CacheManager.InvalidateAllPostList()
}

// 包装原有方法，添加缓存失效逻辑

func (c *CachedPostService) Create(ctx context.Context, postParam *param.Post) (*entity.Post, error) {
	post, err := c.PostService.Create(ctx, postParam)
	if err != nil {
		return nil, err
	}
	// 清除列表缓存
	c.InvalidateAllPostList()
	return post, nil
}

func (c *CachedPostService) Update(ctx context.Context, postID int32, postParam *param.Post) (*entity.Post, error) {
	post, err := c.PostService.Update(ctx, postID, postParam)
	if err != nil {
		return nil, err
	}
	// 清除该文章的所有缓存
	c.InvalidatePostCache(postID)
	// 清除列表缓存
	c.InvalidateAllPostList()
	return post, nil
}

func (c *CachedPostService) Delete(ctx context.Context, postID int32) error {
	err := c.PostService.(interface{ Delete(context.Context, int32) error }).Delete(ctx, postID)
	if err != nil {
		return err
	}
	// 清除该文章的所有缓存
	c.InvalidatePostCache(postID)
	// 清除列表缓存
	c.InvalidateAllPostList()
	return nil
}

// 委托方法

func (c *CachedPostService) Page(ctx context.Context, postQuery param.PostQuery) ([]*entity.Post, int64, error) {
	return c.PostService.Page(ctx, postQuery)
}

func (c *CachedPostService) IncreaseLike(ctx context.Context, postID int32) error {
	return c.PostService.IncreaseLike(ctx, postID)
}

func (c *CachedPostService) GetPrevPosts(ctx context.Context, post *entity.Post, size int) ([]*entity.Post, error) {
	return c.PostService.GetPrevPosts(ctx, post, size)
}

func (c *CachedPostService) GetNextPosts(ctx context.Context, post *entity.Post, size int) ([]*entity.Post, error) {
	return c.PostService.GetNextPosts(ctx, post, size)
}

func (c *CachedPostService) CountByStatus(ctx context.Context, status consts.PostStatus) (int64, error) {
	return c.PostService.CountByStatus(ctx, status)
}

func (c *CachedPostService) CountVisit(ctx context.Context) (int64, error) {
	return c.PostService.CountVisit(ctx)
}

func (c *CachedPostService) Preview(ctx context.Context, postID int32) (string, error) {
	return c.PostService.Preview(ctx, postID)
}

func (c *CachedPostService) CountLike(ctx context.Context) (int64, error) {
	return c.PostService.CountLike(ctx)
}

// 包装 BasePostService 方法

func (c *CachedPostService) GetByStatus(ctx context.Context, status []consts.PostStatus, postType consts.PostType, sort *param.Sort) ([]*entity.Post, error) {
	return c.PostService.GetByStatus(ctx, status, postType, sort)
}

func (c *CachedPostService) BuildFullPath(ctx context.Context, post *entity.Post) (string, error) {
	return c.PostService.BuildFullPath(ctx, post)
}

func (c *CachedPostService) GetByPostIDs(ctx context.Context, postIDs []int32) (map[int32]*entity.Post, error) {
	return c.PostService.GetByPostIDs(ctx, postIDs)
}

func (c *CachedPostService) GetBySlug(ctx context.Context, slug string) (*entity.Post, error) {
	return c.PostService.GetBySlug(ctx, slug)
}

func (c *CachedPostService) GetByPostID(ctx context.Context, postID int32) (*entity.Post, error) {
	return c.PostService.GetByPostID(ctx, postID)
}

func (c *CachedPostService) GenerateSummary(ctx context.Context, htmlContent string) string {
	return c.PostService.GenerateSummary(ctx, htmlContent)
}

func (c *CachedPostService) UpdateStatus(ctx context.Context, postID int32, status consts.PostStatus) (*entity.Post, error) {
	post, err := c.PostService.UpdateStatus(ctx, postID, status)
	if err != nil {
		return nil, err
	}
	// 清除缓存
	c.InvalidatePostCache(postID)
	c.InvalidateAllPostList()
	return post, nil
}

func (c *CachedPostService) UpdateStatusBatch(ctx context.Context, status consts.PostStatus, postIDs []int32) ([]*entity.Post, error) {
	posts, err := c.PostService.(interface {
		UpdateStatusBatch(context.Context, consts.PostStatus, []int32) ([]*entity.Post, error)
	}).UpdateStatusBatch(ctx, status, postIDs)
	if err != nil {
		return nil, err
	}
	// 清除缓存
	for _, id := range postIDs {
		c.InvalidatePostCache(id)
	}
	c.InvalidateAllPostList()
	return posts, nil
}

func (c *CachedPostService) UpdateDraftContent(ctx context.Context, postID int32, content, originalContent string) (*entity.Post, error) {
	post, err := c.PostService.(interface {
		UpdateDraftContent(context.Context, int32, string, string) (*entity.Post, error)
	}).UpdateDraftContent(ctx, postID, content, originalContent)
	if err != nil {
		return nil, err
	}
	// 清除缓存
	c.InvalidatePostCache(postID)
	return post, nil
}

func (c *CachedPostService) IncreaseVisit(ctx context.Context, postID int32) {
	c.PostService.IncreaseVisit(ctx, postID)
}

func (c *CachedPostService) CreateOrUpdate(ctx context.Context, post *entity.Post, categoryIDs, tagIDs []int32, metas []param.Meta) (*entity.Post, error) {
	return c.PostService.(interface {
		CreateOrUpdate(context.Context, *entity.Post, []int32, []int32, []param.Meta) (*entity.Post, error)
	}).CreateOrUpdate(ctx, post, categoryIDs, tagIDs, metas)
}

func (c *CachedPostService) DeleteBatch(ctx context.Context, postIDs []int32) error {
	err := c.PostService.(interface {
		DeleteBatch(context.Context, []int32) error
	}).DeleteBatch(ctx, postIDs)
	if err != nil {
		return err
	}
	// 清除缓存
	for _, id := range postIDs {
		c.InvalidatePostCache(id)
	}
	c.InvalidateAllPostList()
	return nil
}

// Ensure CachedPostService implements PostService
var _ service.PostService = (*CachedPostService)(nil)
