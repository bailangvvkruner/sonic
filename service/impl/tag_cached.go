package impl

import (
	"context"

	"github.com/go-sonic/sonic/cache"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/entity"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/service"
)

// CachedTagService 带缓存的标签服务
type CachedTagService struct {
	TagService   service.TagService
	CacheManager *cache.CacheManager
}

// NewCachedTagService 创建带缓存的标签服务
func NewCachedTagService(
	tagService service.TagService,
	cacheManager *cache.CacheManager,
) *CachedTagService {
	return &CachedTagService{
		TagService:   tagService,
		CacheManager: cacheManager,
	}
}

// ListAll 获取所有标签（带缓存）
func (c *CachedTagService) ListAll(ctx context.Context, sort *param.Sort) ([]*entity.Tag, error) {
	// 1. 尝试从缓存获取（仅当 sort 为空时使用缓存）
	if sort == nil {
		if cached, ok := c.CacheManager.GetTagAll(); ok {
			return cached, nil
		}
	}

	// 2. 缓存未命中，查询数据库
	tags, err := c.TagService.ListAll(ctx, sort)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存（仅当 sort 为空时）
	if sort == nil {
		_ = c.CacheManager.SetTagAll(tags)
		// 同时缓存单个标签
		for _, tag := range tags {
			_ = c.CacheManager.SetTagByID(tag)
		}
	}

	return tags, nil
}

// GetByID 根据 ID 获取标签（带缓存）
func (c *CachedTagService) GetByID(ctx context.Context, id int32) (*entity.Tag, error) {
	// 1. 尝试从缓存获取
	if cached, ok := c.CacheManager.GetTagByID(id); ok {
		return cached, nil
	}

	// 2. 缓存未命中，查询数据库
	tag, err := c.TagService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	if tag != nil {
		_ = c.CacheManager.SetTagByID(tag)
	}

	return tag, nil
}

// InvalidateTagCache 清除标签缓存
func (c *CachedTagService) InvalidateTagCache() {
	c.CacheManager.DeleteTagAll()
}

// InvalidateTagByID 清除单个标签缓存
func (c *CachedTagService) InvalidateTagByID(id int32) {
	c.CacheManager.DeleteTagByID(id)
}

// 包装原有方法，添加缓存失效逻辑

func (c *CachedTagService) Create(ctx context.Context, tagParam *param.Tag) (*entity.Tag, error) {
	tag, err := c.TagService.Create(ctx, tagParam)
	if err != nil {
		return nil, err
	}
	// 清除缓存
	c.InvalidateTagCache()
	return tag, nil
}

func (c *CachedTagService) Update(ctx context.Context, id int32, tagParam *param.Tag) (*entity.Tag, error) {
	tag, err := c.TagService.Update(ctx, id, tagParam)
	if err != nil {
		return nil, err
	}
	// 清除缓存
	c.InvalidateTagCache()
	c.InvalidateTagByID(id)
	return tag, nil
}

func (c *CachedTagService) Delete(ctx context.Context, id int32) error {
	err := c.TagService.Delete(ctx, id)
	if err != nil {
		return err
	}
	// 清除缓存
	c.InvalidateTagCache()
	c.InvalidateTagByID(id)
	return nil
}

// 委托方法

func (c *CachedTagService) GetBySlug(ctx context.Context, slug string) (*entity.Tag, error) {
	return c.TagService.GetBySlug(ctx, slug)
}

func (c *CachedTagService) ConvertToDTO(ctx context.Context, tag *entity.Tag) (*dto.Tag, error) {
	return c.TagService.ConvertToDTO(ctx, tag)
}

func (c *CachedTagService) ConvertToDTOs(ctx context.Context, tags []*entity.Tag) ([]*dto.Tag, error) {
	return c.TagService.ConvertToDTOs(ctx, tags)
}

func (c *CachedTagService) ListByIDs(ctx context.Context, tagIDs []int32) ([]*entity.Tag, error) {
	return c.TagService.ListByIDs(ctx, tagIDs)
}

func (c *CachedTagService) CountAllTag(ctx context.Context) (int64, error) {
	return c.TagService.CountAllTag(ctx)
}

func (c *CachedTagService) GetByName(ctx context.Context, name string) (*entity.Tag, error) {
	return c.TagService.GetByName(ctx, name)
}

// Ensure CachedTagService implements TagService
var _ service.TagService = (*CachedTagService)(nil)
