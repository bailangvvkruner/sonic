package impl

import (
	"context"

	"github.com/go-sonic/sonic/cache"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/entity"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/model/vo"
	"github.com/go-sonic/sonic/service"
)

// CachedCategoryService 带缓存的分类服务
type CachedCategoryService struct {
	CategoryService service.CategoryService
	CacheManager    *cache.CacheManager
}

// NewCachedCategoryService 创建带缓存的分类服务
func NewCachedCategoryService(
	categoryService service.CategoryService,
	cacheManager *cache.CacheManager,
) *CachedCategoryService {
	return &CachedCategoryService{
		CategoryService: categoryService,
		CacheManager:    cacheManager,
	}
}

// ListAll 获取所有分类（带缓存）
func (c *CachedCategoryService) ListAll(ctx context.Context, sort *param.Sort) ([]*entity.Category, error) {
	// 1. 尝试从缓存获取（仅当 sort 为空时使用缓存）
	if sort == nil {
		if cached, ok := c.CacheManager.GetCategoryAll(); ok {
			return cached, nil
		}
	}

	// 2. 缓存未命中，查询数据库
	categories, err := c.CategoryService.ListAll(ctx, sort)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存（仅当 sort 为空时）
	if sort == nil {
		_ = c.CacheManager.SetCategoryAll(categories)
		// 同时缓存单个分类
		for _, cat := range categories {
			_ = c.CacheManager.SetCategoryByID(cat)
		}
	}

	return categories, nil
}

// GetByID 根据 ID 获取分类（带缓存）
func (c *CachedCategoryService) GetByID(ctx context.Context, id int32) (*entity.Category, error) {
	// 1. 尝试从缓存获取
	if cached, ok := c.CacheManager.GetCategoryByID(id); ok {
		return cached, nil
	}

	// 2. 缓存未命中，查询数据库
	category, err := c.CategoryService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	if category != nil {
		_ = c.CacheManager.SetCategoryByID(category)
	}

	return category, nil
}

// ListAsTree 获取分类树（带缓存）
func (c *CachedCategoryService) ListAsTree(ctx context.Context, sort *param.Sort, fillPassword bool) ([]*vo.CategoryVO, error) {
	// 1. 尝试从缓存获取（仅当 sort 为空且 fillPassword 为 false 时）
	if sort == nil && !fillPassword {
		if cached, ok := c.CacheManager.GetCategoryTree(); ok {
			return cached, nil
		}
	}

	// 2. 缓存未命中，查询数据库
	tree, err := c.CategoryService.ListAsTree(ctx, sort, fillPassword)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	if sort == nil && !fillPassword {
		_ = c.CacheManager.SetCategoryTree(tree)
	}

	return tree, nil
}

// InvalidateCategoryCache 清除分类缓存
func (c *CachedCategoryService) InvalidateCategoryCache() {
	c.CacheManager.DeleteCategoryAll()
}

// InvalidateCategoryByID 清除单个分类缓存
func (c *CachedCategoryService) InvalidateCategoryByID(id int32) {
	c.CacheManager.DeleteCategoryByID(id)
}

// 包装原有方法，添加缓存失效逻辑

func (c *CachedCategoryService) Create(ctx context.Context, categoryParam *param.Category) (*entity.Category, error) {
	category, err := c.CategoryService.Create(ctx, categoryParam)
	if err != nil {
		return nil, err
	}
	// 清除缓存
	c.InvalidateCategoryCache()
	return category, nil
}

func (c *CachedCategoryService) Update(ctx context.Context, categoryParam *param.Category) (*entity.Category, error) {
	category, err := c.CategoryService.Update(ctx, categoryParam)
	if err != nil {
		return nil, err
	}
	// 清除缓存
	c.InvalidateCategoryCache()
	c.InvalidateCategoryByID(categoryParam.ID)
	return category, nil
}

func (c *CachedCategoryService) Delete(ctx context.Context, categoryID int32) error {
	err := c.CategoryService.Delete(ctx, categoryID)
	if err != nil {
		return err
	}
	// 清除缓存
	c.InvalidateCategoryCache()
	c.InvalidateCategoryByID(categoryID)
	return nil
}

// 委托方法

func (c *CachedCategoryService) GetBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	return c.CategoryService.GetBySlug(ctx, slug)
}

func (c *CachedCategoryService) GetByName(ctx context.Context, name string) (*entity.Category, error) {
	return c.CategoryService.GetByName(ctx, name)
}

func (c *CachedCategoryService) ListCategoryWithPostCountDTO(ctx context.Context, sort *param.Sort) ([]*dto.CategoryWithPostCount, error) {
	return c.CategoryService.ListCategoryWithPostCountDTO(ctx, sort)
}

func (c *CachedCategoryService) ConvertToCategoryDTO(ctx context.Context, e *entity.Category) (*dto.CategoryDTO, error) {
	return c.CategoryService.ConvertToCategoryDTO(ctx, e)
}

func (c *CachedCategoryService) ConvertToCategoryDTOs(ctx context.Context, categories []*entity.Category) ([]*dto.CategoryDTO, error) {
	return c.CategoryService.ConvertToCategoryDTOs(ctx, categories)
}

func (c *CachedCategoryService) UpdateBatch(ctx context.Context, categoryParams []*param.Category) ([]*entity.Category, error) {
	categories, err := c.CategoryService.UpdateBatch(ctx, categoryParams)
	if err != nil {
		return nil, err
	}
	// 清除缓存
	c.InvalidateCategoryCache()
	for _, param := range categoryParams {
		c.InvalidateCategoryByID(param.ID)
	}
	return categories, nil
}

func (c *CachedCategoryService) ListByIDs(ctx context.Context, categoryIDs []int32) ([]*entity.Category, error) {
	return c.CategoryService.ListByIDs(ctx, categoryIDs)
}

func (c *CachedCategoryService) IsCategoriesEncrypt(ctx context.Context, categoryIDs ...int32) (bool, error) {
	return c.CategoryService.IsCategoriesEncrypt(ctx, categoryIDs...)
}

func (c *CachedCategoryService) Count(ctx context.Context) (int64, error) {
	return c.CategoryService.Count(ctx)
}

func (c *CachedCategoryService) GetChildCategory(ctx context.Context, parentCategoryID int32) ([]*entity.Category, error) {
	return c.CategoryService.GetChildCategory(ctx, parentCategoryID)
}

// Ensure CachedCategoryService implements CategoryService
var _ service.CategoryService = (*CachedCategoryService)(nil)
