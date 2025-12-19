package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/go-sonic/sonic/cache"
	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/dal"
	"github.com/go-sonic/sonic/model/entity"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util"
)

type postCategoryServiceImpl struct {
	CategoryService service.CategoryService
	OptionService   service.OptionService
	Cache           cache.Cache
}

func NewPostCategoryService(categoryService service.CategoryService, optionService service.OptionService, cache cache.Cache) service.PostCategoryService {
	return &postCategoryServiceImpl{
		CategoryService: categoryService,
		OptionService:   optionService,
		Cache:           cache,
	}
}

func (p *postCategoryServiceImpl) ListByPostIDs(ctx context.Context, postIDs []int32) ([]*entity.PostCategory, error) {
	postCategoryDAL := dal.GetQueryByCtx(ctx).PostCategory
	postCategories, err := postCategoryDAL.WithContext(ctx).Where(postCategoryDAL.PostID.In(postIDs...)).Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return postCategories, nil
}

func (p *postCategoryServiceImpl) ListCategoryMapByPostID(ctx context.Context, postIDs []int32) (map[int32][]*entity.Category, error) {
	result := make(map[int32][]*entity.Category, 0)
	if len(postIDs) == 0 {
		return result, nil
	}

	// Generate cache key
	sortedIDs := make([]int, len(postIDs))
	for i, id := range postIDs {
		sortedIDs[i] = int(id)
	}
	sort.Ints(sortedIDs)
	
	sb := strings.Builder{}
	for _, id := range sortedIDs {
		sb.WriteString(fmt.Sprintf("%d,", id))
	}
	cacheKey := "post_category:map:" + util.Md5Hex(sb.String())

	if cached, ok := p.Cache.Get(cacheKey); ok {
		if err := json.Unmarshal(cached.([]byte), &result); err == nil {
			return result, nil
		}
	}

	postCategories, err := p.ListByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, err
	}
	if len(postCategories) == 0 {
		return result, nil
	}
	categoryIDs := make([]int32, 0)
	for _, postCategory := range postCategories {
		categoryIDs = append(categoryIDs, postCategory.CategoryID)
	}
	categories, err := p.CategoryService.ListByIDs(ctx, categoryIDs)
	if err != nil {
		return nil, err
	}
	categoryIDMap := make(map[int32]*entity.Category, 0)
	for _, category := range categories {
		categoryIDMap[category.ID] = category
	}
	for _, postCategory := range postCategories {
		category, ok := categoryIDMap[postCategory.CategoryID]
		if !ok {
			continue
		}
		result[postCategory.PostID] = append(result[postCategory.PostID], category)
	}

	if bytes, err := json.Marshal(result); err == nil {
		duration, _ := p.OptionService.GetCacheDuration(ctx)
		if duration > 0 {
			p.Cache.Set(cacheKey, bytes, duration)
		}
	}
	return result, nil
}

func (p *postCategoryServiceImpl) ListCategoryByPostID(ctx context.Context, postID int32) ([]*entity.Category, error) {
	categoryMap, err := p.ListCategoryMapByPostID(ctx, []int32{postID})
	if err != nil {
		return nil, err
	}
	categories, ok := categoryMap[postID]
	if !ok {
		return make([]*entity.Category, 0), nil
	}
	return categories, nil
}

func (p *postCategoryServiceImpl) ListByCategoryID(ctx context.Context, categoryID int32, status consts.PostStatus) ([]*entity.Post, error) {
	postCategoryDAL := dal.GetQueryByCtx(ctx).PostCategory
	postDAL := dal.GetQueryByCtx(ctx).Post

	postIDsQuery := postCategoryDAL.WithContext(ctx).Where(postCategoryDAL.CategoryID.Eq(categoryID)).Select(postCategoryDAL.PostID)
	posts, err := postDAL.WithContext(ctx).Where(postDAL.WithContext(ctx).Columns(postDAL.ID).In(postIDsQuery), postDAL.Status.Eq(status)).Find()
	return posts, WrapDBErr(err)
}
