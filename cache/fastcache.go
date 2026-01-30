package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/VictoriaMetrics/fastcache"
)

const (
	// 默认缓存大小 512MB
	DefaultCacheSize = 512 * 1024 * 1024
	// 最大条目大小 64KB
	MaxEntrySize = 64 * 1024
)

// FastCache 基于 fastcache 的高性能缓存实现
type FastCache struct {
	cache *fastcache.Cache
}

// NewFastCache 创建 FastCache 实例
func NewFastCache(maxBytes int) *FastCache {
	if maxBytes <= 0 {
		maxBytes = DefaultCacheSize
	}
	return &FastCache{
		cache: fastcache.New(maxBytes),
	}
}

// Set 存储数据到缓存
func (f *FastCache) Set(key string, value []byte) {
	f.cache.Set([]byte(key), value)
}

// SetWithTTL 存储数据到缓存（fastcache 本身不支持 TTL，需要在外层处理）
func (f *FastCache) SetWithTTL(key string, value []byte, ttl time.Duration) {
	// fastcache 不支持原生 TTL，存储时附带过期时间戳
	item := &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}
	data, _ := json.Marshal(item)
	f.cache.Set([]byte(key), data)
}

// Get 从缓存获取数据
func (f *FastCache) Get(key string) ([]byte, bool) {
	data := f.cache.Get(nil, []byte(key))
	if data == nil {
		return nil, false
	}

	// 尝试解析带 TTL 的缓存项
	var item CacheItem
	if err := json.Unmarshal(data, &item); err == nil && item.ExpiresAt > 0 {
		if time.Now().Unix() > item.ExpiresAt {
			f.cache.Del([]byte(key))
			return nil, false
		}
		return item.Value, true
	}

	// 原始数据（不带 TTL）
	return data, true
}

// Delete 删除缓存
func (f *FastCache) Delete(key string) {
	f.cache.Del([]byte(key))
}

// Reset 清空缓存
func (f *FastCache) Reset() {
	f.cache.Reset()
}

// Has 检查 key 是否存在
func (f *FastCache) Has(key string) bool {
	return f.cache.Has([]byte(key))
}

// CacheItem 带过期时间的缓存项
type CacheItem struct {
	Value     []byte `json:"v"`
	ExpiresAt int64  `json:"e,omitempty"`
}

// CacheKey 缓存 key 生成器
type CacheKey struct {
	prefix string
}

// NewCacheKey 创建缓存 key 生成器
func NewCacheKey(prefix string) *CacheKey {
	return &CacheKey{prefix: prefix}
}

// Build 构建缓存 key
func (c *CacheKey) Build(parts ...interface{}) string {
	key := c.prefix
	for _, part := range parts {
		key += fmt.Sprintf(":%v", part)
	}
	return key
}

// PostCacheKey 文章缓存 key
type PostCacheKey struct{}

func (p PostCacheKey) Detail(postID int32) string {
	return fmt.Sprintf("post:detail:%d", postID)
}

func (p PostCacheKey) List(page, size int, sort string) string {
	return fmt.Sprintf("post:list:%d:%d:%s", page, size, sort)
}

func (p PostCacheKey) Archives(page int) string {
	return fmt.Sprintf("post:archives:%d", page)
}

func (p PostCacheKey) PrevNext(postID int32) string {
	return fmt.Sprintf("post:prevnext:%d", postID)
}

// CategoryCacheKey 分类缓存 key
type CategoryCacheKey struct{}

func (c CategoryCacheKey) All() string {
	return "category:all"
}

func (c CategoryCacheKey) ByID(id int32) string {
	return fmt.Sprintf("category:id:%d", id)
}

func (c CategoryCacheKey) BySlug(slug string) string {
	return fmt.Sprintf("category:slug:%s", slug)
}

func (c CategoryCacheKey) Tree() string {
	return "category:tree"
}

// TagCacheKey 标签缓存 key
type TagCacheKey struct{}

func (t TagCacheKey) All() string {
	return "tag:all"
}

func (t TagCacheKey) ByID(id int32) string {
	return fmt.Sprintf("tag:id:%d", id)
}

func (t TagCacheKey) BySlug(slug string) string {
	return fmt.Sprintf("tag:slug:%s", slug)
}

// MetaCacheKey Meta 缓存 key
type MetaCacheKey struct{}

func (m MetaCacheKey) ByPostID(postID int32) string {
	return fmt.Sprintf("meta:post:%d", postID)
}

// PostTagCacheKey 文章标签关联缓存 key
type PostTagCacheKey struct{}

func (p PostTagCacheKey) ByPostID(postID int32) string {
	return fmt.Sprintf("posttag:post:%d", postID)
}

func (p PostTagCacheKey) ByPostIDs(postIDs string) string {
	return fmt.Sprintf("posttag:posts:%s", postIDs)
}

// PostCategoryCacheKey 文章分类关联缓存 key
type PostCategoryCacheKey struct{}

func (p PostCategoryCacheKey) ByPostID(postID int32) string {
	return fmt.Sprintf("postcat:post:%d", postID)
}

func (p PostCategoryCacheKey) ByPostIDs(postIDs string) string {
	return fmt.Sprintf("postcat:posts:%s", postIDs)
}
