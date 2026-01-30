package listener

import (
	"context"

	"go.uber.org/zap"

	"github.com/go-sonic/sonic/cache"
	"github.com/go-sonic/sonic/event"
	"github.com/go-sonic/sonic/log"
	"github.com/go-sonic/sonic/model/entity"
	"github.com/go-sonic/sonic/service"
)

// CacheWarmupListener 缓存预热监听器
type CacheWarmupListener struct {
	CacheManager    *cache.CacheManager
	CategoryService service.CategoryService
	TagService      service.TagService
}

// NewCacheWarmupListener 创建缓存预热监听器
func NewCacheWarmupListener(
	cacheManager *cache.CacheManager,
	categoryService service.CategoryService,
	tagService service.TagService,
) *CacheWarmupListener {
	return &CacheWarmupListener{
		CacheManager:    cacheManager,
		CategoryService: categoryService,
		TagService:      tagService,
	}
}

// HandleStartEvent 处理启动事件，进行缓存预热
func (c *CacheWarmupListener) HandleStartEvent(ctx context.Context, startEvent *event.StartEvent) error {
	log.Info("开始缓存预热...")

	// 预热分类和标签缓存
	err := c.CacheManager.WarmUpCache(
		func() ([]*entity.Category, error) {
			return c.CategoryService.ListAll(ctx, nil)
		},
		func() ([]*entity.Tag, error) {
			return c.TagService.ListAll(ctx, nil)
		},
	)

	if err != nil {
		log.Error("缓存预热失败", zap.Error(err))
		return err
	}

	// 获取缓存统计
	stats := c.CacheManager.Stats()
	log.Info("缓存预热完成",
		zap.Uint64("entries_count", stats.EntriesCount),
		zap.Uint64("bytes_size", stats.BytesSize),
	)

	return nil
}

// RegisterCacheWarmupListener 注册缓存预热监听器
func RegisterCacheWarmupListener(eventBus event.Bus, listener *CacheWarmupListener) {
	eventBus.Subscribe(event.StartEventName, func(ctx context.Context, e event.Event) error {
		if startEvent, ok := e.(*event.StartEvent); ok {
			return listener.HandleStartEvent(ctx, startEvent)
		}
		return nil
	})
}
