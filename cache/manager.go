package cache

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/go-sonic/sonic/model/entity"
	"github.com/go-sonic/sonic/model/vo"
)

const (
	// // 文章详情缓存时间 24小时
	// PostDetailTTL = time.Hour * 24
	// // 文章列表缓存时间 5分钟
	// PostListTTL = time.Minute * 5
	// // 分类缓存时间 7天
	// CategoryTTL = time.Hour * 24 * 7
	// // 标签缓存时间 7天
	// TagTTL = time.Hour * 24 * 7
	// // Meta 缓存时间 24小时
	// MetaTTL = time.Hour * 24
	// // 上下篇文章缓存时间 1小时
	// PrevNextTTL = time.Hour

	// 文章详情缓存时间 24小时
	PostDetailTTL = time.Minute * 5
	// 文章列表缓存时间 5分钟
	PostListTTL = time.Minute * 5
	// 分类缓存时间 7天
	CategoryTTL = time.Minute * 5
	// 标签缓存时间 7天
	TagTTL = time.Minute * 5
	// Meta 缓存时间 24小时
	MetaTTL = time.Minute * 5
	// 上下篇文章缓存时间 1小时
	PrevNextTTL = time.Minute * 5
)

// CacheManager 缓存管理器
type CacheManager struct {
	cache *FastCache
	// Key 生成器
	PostKey      PostCacheKey
	CategoryKey  CategoryCacheKey
	TagKey       TagCacheKey
	MetaKey      MetaCacheKey
	PostTagKey   PostTagCacheKey
	PostCatKey   PostCategoryCacheKey
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(cache *FastCache) *CacheManager {
	return &CacheManager{
		cache: cache,
	}
}

// ==================== 文章缓存 ====================

// SetPostDetail 缓存文章详情
func (c *CacheManager) SetPostDetail(post *vo.PostDetailVO) error {
	if post == nil {
		return nil
	}
	key := c.PostKey.Detail(post.ID)
	data, err := json.Marshal(post)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, PostDetailTTL)
	return nil
}

// GetPostDetail 获取文章详情缓存
func (c *CacheManager) GetPostDetail(postID int32) (*vo.PostDetailVO, bool) {
	key := c.PostKey.Detail(postID)
	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var post vo.PostDetailVO
	if err := json.Unmarshal(data, &post); err != nil {
		return nil, false
	}
	return &post, true
}

// DeletePostDetail 删除文章详情缓存
func (c *CacheManager) DeletePostDetail(postID int32) {
	key := c.PostKey.Detail(postID)
	c.cache.Delete(key)
}

// SetPostList 缓存文章列表
func (c *CacheManager) SetPostList(page, size int, sort string, posts []*vo.Post) error {
	key := c.PostKey.List(page, size, sort)
	data, err := json.Marshal(posts)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, PostListTTL)
	return nil
}

// GetPostList 获取文章列表缓存
func (c *CacheManager) GetPostList(page, size int, sort string) ([]*vo.Post, bool) {
	key := c.PostKey.List(page, size, sort)
	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var posts []*vo.Post
	if err := json.Unmarshal(data, &posts); err != nil {
		return nil, false
	}
	return posts, true
}

// DeletePostList 删除所有文章列表缓存
func (c *CacheManager) DeletePostList() {
	// fastcache 不支持前缀删除，这里通过重置整个缓存实现
	// 或者可以维护一个列表缓存 key 的索引
	c.cache.Reset()
}

// SetPrevNextPosts 缓存上下篇文章
func (c *CacheManager) SetPrevNextPosts(postID int32, prev, next *vo.Post) error {
	key := c.PostKey.PrevNext(postID)
	type prevNext struct {
		Prev *vo.Post `json:"prev"`
		Next *vo.Post `json:"next"`
	}
	pn := prevNext{Prev: prev, Next: next}
	data, err := json.Marshal(pn)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, PrevNextTTL)
	return nil
}

// GetPrevNextPosts 获取上下篇文章缓存
func (c *CacheManager) GetPrevNextPosts(postID int32) (prev, next *vo.Post, ok bool) {
	key := c.PostKey.PrevNext(postID)
	data, found := c.cache.Get(key)
	if !found {
		return nil, nil, false
	}
	type prevNext struct {
		Prev *vo.Post `json:"prev"`
		Next *vo.Post `json:"next"`
	}
	var pn prevNext
	if err := json.Unmarshal(data, &pn); err != nil {
		return nil, nil, false
	}
	return pn.Prev, pn.Next, true
}

// ==================== 分类缓存 ====================

// SetCategoryAll 缓存所有分类
func (c *CacheManager) SetCategoryAll(categories []*entity.Category) error {
	key := c.CategoryKey.All()
	data, err := json.Marshal(categories)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, CategoryTTL)
	return nil
}

// GetCategoryAll 获取所有分类缓存
func (c *CacheManager) GetCategoryAll() ([]*entity.Category, bool) {
	key := c.CategoryKey.All()
	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var categories []*entity.Category
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, false
	}
	return categories, true
}

// SetCategoryByID 缓存单个分类
func (c *CacheManager) SetCategoryByID(category *entity.Category) error {
	if category == nil {
		return nil
	}
	key := c.CategoryKey.ByID(category.ID)
	data, err := json.Marshal(category)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, CategoryTTL)
	return nil
}

// GetCategoryByID 获取单个分类缓存
func (c *CacheManager) GetCategoryByID(id int32) (*entity.Category, bool) {
	key := c.CategoryKey.ByID(id)
	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var category entity.Category
	if err := json.Unmarshal(data, &category); err != nil {
		return nil, false
	}
	return &category, true
}

// SetCategoryTree 缓存分类树
func (c *CacheManager) SetCategoryTree(tree []*vo.CategoryVO) error {
	key := c.CategoryKey.Tree()
	data, err := json.Marshal(tree)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, CategoryTTL)
	return nil
}

// GetCategoryTree 获取分类树缓存
func (c *CacheManager) GetCategoryTree() ([]*vo.CategoryVO, bool) {
	key := c.CategoryKey.Tree()
	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var tree []*vo.CategoryVO
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, false
	}
	return tree, true
}

// DeleteCategoryAll 删除所有分类缓存
func (c *CacheManager) DeleteCategoryAll() {
	// 由于 fastcache 不支持前缀删除，这里只删除已知 key
	c.cache.Delete(c.CategoryKey.All())
	c.cache.Delete(c.CategoryKey.Tree())
}

// DeleteCategoryByID 删除单个分类缓存
func (c *CacheManager) DeleteCategoryByID(id int32) {
	c.cache.Delete(c.CategoryKey.ByID(id))
}

// ==================== 标签缓存 ====================

// SetTagAll 缓存所有标签
func (c *CacheManager) SetTagAll(tags []*entity.Tag) error {
	key := c.TagKey.All()
	data, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, TagTTL)
	return nil
}

// GetTagAll 获取所有标签缓存
func (c *CacheManager) GetTagAll() ([]*entity.Tag, bool) {
	key := c.TagKey.All()
	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var tags []*entity.Tag
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, false
	}
	return tags, true
}

// SetTagByID 缓存单个标签
func (c *CacheManager) SetTagByID(tag *entity.Tag) error {
	if tag == nil {
		return nil
	}
	key := c.TagKey.ByID(tag.ID)
	data, err := json.Marshal(tag)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, TagTTL)
	return nil
}

// GetTagByID 获取单个标签缓存
func (c *CacheManager) GetTagByID(id int32) (*entity.Tag, bool) {
	key := c.TagKey.ByID(id)
	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var tag entity.Tag
	if err := json.Unmarshal(data, &tag); err != nil {
		return nil, false
	}
	return &tag, true
}

// DeleteTagAll 删除所有标签缓存
func (c *CacheManager) DeleteTagAll() {
	c.cache.Delete(c.TagKey.All())
}

// DeleteTagByID 删除单个标签缓存
func (c *CacheManager) DeleteTagByID(id int32) {
	c.cache.Delete(c.TagKey.ByID(id))
}

// ==================== Meta 缓存 ====================

// SetPostMeta 缓存文章 Meta
func (c *CacheManager) SetPostMeta(postID int32, metas []*entity.Meta) error {
	key := c.MetaKey.ByPostID(postID)
	data, err := json.Marshal(metas)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, MetaTTL)
	return nil
}

// GetPostMeta 获取文章 Meta 缓存
func (c *CacheManager) GetPostMeta(postID int32) ([]*entity.Meta, bool) {
	key := c.MetaKey.ByPostID(postID)
	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var metas []*entity.Meta
	if err := json.Unmarshal(data, &metas); err != nil {
		return nil, false
	}
	return metas, true
}

// DeletePostMeta 删除文章 Meta 缓存
func (c *CacheManager) DeletePostMeta(postID int32) {
	key := c.MetaKey.ByPostID(postID)
	c.cache.Delete(key)
}

// ==================== 文章标签关联缓存 ====================

// SetPostTags 缓存文章标签
func (c *CacheManager) SetPostTags(postID int32, tags []*entity.Tag) error {
	key := c.PostTagKey.ByPostID(postID)
	data, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, TagTTL)
	return nil
}

// GetPostTags 获取文章标签缓存
func (c *CacheManager) GetPostTags(postID int32) ([]*entity.Tag, bool) {
	key := c.PostTagKey.ByPostID(postID)
	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var tags []*entity.Tag
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, false
	}
	return tags, true
}

// SetPostTagsMap 缓存多篇文章的标签映射
func (c *CacheManager) SetPostTagsMap(postIDs []int32, tagMap map[int32][]*entity.Tag) error {
	// 排序 postIDs 保证 key 一致性
	sortedIDs := make([]int, len(postIDs))
	for i, id := range postIDs {
		sortedIDs[i] = int(id)
	}
	sort.Ints(sortedIDs)
	idsStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(sortedIDs)), ","), "[]")
	key := c.PostTagKey.ByPostIDs(idsStr)

	data, err := json.Marshal(tagMap)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, TagTTL)
	return nil
}

// GetPostTagsMap 获取多篇文章的标签映射缓存
func (c *CacheManager) GetPostTagsMap(postIDs []int32) (map[int32][]*entity.Tag, bool) {
	sortedIDs := make([]int, len(postIDs))
	for i, id := range postIDs {
		sortedIDs[i] = int(id)
	}
	sort.Ints(sortedIDs)
	idsStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(sortedIDs)), ","), "[]")
	key := c.PostTagKey.ByPostIDs(idsStr)

	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var tagMap map[int32][]*entity.Tag
	if err := json.Unmarshal(data, &tagMap); err != nil {
		return nil, false
	}
	return tagMap, true
}

// DeletePostTags 删除文章标签缓存
func (c *CacheManager) DeletePostTags(postID int32) {
	key := c.PostTagKey.ByPostID(postID)
	c.cache.Delete(key)
}

// ==================== 文章分类关联缓存 ====================

// SetPostCategories 缓存文章分类
func (c *CacheManager) SetPostCategories(postID int32, categories []*entity.Category) error {
	key := c.PostCatKey.ByPostID(postID)
	data, err := json.Marshal(categories)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, CategoryTTL)
	return nil
}

// GetPostCategories 获取文章分类缓存
func (c *CacheManager) GetPostCategories(postID int32) ([]*entity.Category, bool) {
	key := c.PostCatKey.ByPostID(postID)
	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var categories []*entity.Category
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, false
	}
	return categories, true
}

// SetPostCategoriesMap 缓存多篇文章的分类映射
func (c *CacheManager) SetPostCategoriesMap(postIDs []int32, catMap map[int32][]*entity.Category) error {
	sortedIDs := make([]int, len(postIDs))
	for i, id := range postIDs {
		sortedIDs[i] = int(id)
	}
	sort.Ints(sortedIDs)
	idsStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(sortedIDs)), ","), "[]")
	key := c.PostCatKey.ByPostIDs(idsStr)

	data, err := json.Marshal(catMap)
	if err != nil {
		return err
	}
	c.cache.SetWithTTL(key, data, CategoryTTL)
	return nil
}

// GetPostCategoriesMap 获取多篇文章的分类映射缓存
func (c *CacheManager) GetPostCategoriesMap(postIDs []int32) (map[int32][]*entity.Category, bool) {
	sortedIDs := make([]int, len(postIDs))
	for i, id := range postIDs {
		sortedIDs[i] = int(id)
	}
	sort.Ints(sortedIDs)
	idsStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(sortedIDs)), ","), "[]")
	key := c.PostCatKey.ByPostIDs(idsStr)

	data, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	var catMap map[int32][]*entity.Category
	if err := json.Unmarshal(data, &catMap); err != nil {
		return nil, false
	}
	return catMap, true
}

// DeletePostCategories 删除文章分类缓存
func (c *CacheManager) DeletePostCategories(postID int32) {
	key := c.PostCatKey.ByPostID(postID)
	c.cache.Delete(key)
}

// ==================== 批量删除 ====================

// InvalidatePostCache 清除文章相关所有缓存
func (c *CacheManager) InvalidatePostCache(postID int32) {
	c.DeletePostDetail(postID)
	c.DeletePostTags(postID)
	c.DeletePostCategories(postID)
	c.DeletePostMeta(postID)
	c.cache.Delete(c.PostKey.PrevNext(postID))
}

// InvalidateAllPostList 清除所有文章列表缓存
func (c *CacheManager) InvalidateAllPostList() {
	// 由于 fastcache 的限制，这里重置整个缓存
	// 生产环境可以考虑使用更细粒度的缓存策略
	c.cache.Reset()
}

// WarmUpCache 缓存预热
func (c *CacheManager) WarmUpCache(
	getAllCategories func() ([]*entity.Category, error),
	getAllTags func() ([]*entity.Tag, error),
) error {
	// 预热分类缓存
	if categories, err := getAllCategories(); err == nil {
		c.SetCategoryAll(categories)
		for _, cat := range categories {
			c.SetCategoryByID(cat)
		}
	}

	// 预热标签缓存
	if tags, err := getAllTags(); err == nil {
		c.SetTagAll(tags)
		for _, tag := range tags {
			c.SetTagByID(tag)
		}
	}

	return nil
}

// Stats 返回缓存统计信息
func (c *CacheManager) Stats() *fastcache.Stats {
	stats := new(fastcache.Stats)
	c.cache.cache.UpdateStats(stats)
	return stats
}
