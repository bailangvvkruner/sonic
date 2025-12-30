package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

// OptimizedPostService 优化的内存文章服务 - 无锁读取，批量写入
type OptimizedPostService struct {
	posts      []map[string]interface{}
	slugIndex  map[string]int            // slug -> index 映射
	idIndex    map[int64]int             // id -> index 映射
	mu         sync.RWMutex              // 仅用于写入时保护
	loaded     bool                      // 是否已从数据库加载
	writeQueue chan *PostOperation       // 异步写入队列
}

type PostOperation struct {
	Type     string                 // "create", "update", "delete"
	Data     map[string]interface{}
	Done     chan error             // 完成通知
}

func NewOptimizedPostService() PostService {
	return &OptimizedPostService{
		posts:      make([]map[string]interface{}, 0),
		slugIndex:  make(map[string]int),
		idIndex:    make(map[int64]int),
		writeQueue: make(chan *PostOperation, 100),
	}
}

// LoadFromDB 从SQLite全量加载（启动时执行一次）
func (s *OptimizedPostService) LoadFromDB() error {
	// 模拟从SQLite加载数据
	// 实际实现中，这里会连接数据库并读取所有文章
	
	// 临时测试数据
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 模拟加载1000篇文章
	for i := 1; i <= 1000; i++ {
		post := map[string]interface{}{
			"id":         int64(i),
			"title":      "测试文章 " + string(rune(i)),
			"content":    "这是测试文章的内容...",
			"slug":       "test-post-" + string(rune(i)),
			"status":     "published",
			"category":   int64(i % 10),
			"tags":       []int64{int64(i % 5)},
			"created_at": time.Now().Add(-time.Hour * time.Duration(i)),
			"updated_at": time.Now(),
		}
		
		s.posts = append(s.posts, post)
		s.slugIndex[post["slug"].(string)] = i - 1
		s.idIndex[post["id"].(int64)] = i - 1
	}
	
	s.loaded = true
	return nil
}

// StartWriteQueue 启动异步写入队列（可选）
func (s *OptimizedPostService) StartWriteQueue() {
	go func() {
		for op := range s.writeQueue {
			var err error
			switch op.Type {
			case "create":
				err = s.createSync(op.Data)
			case "update":
				err = s.updateSync(op.Data)
			case "delete":
				err = s.deleteSync(op.Data)
			}
			if op.Done != nil {
				op.Done <- err
			}
		}
	}()
}

// GetBySlug O(1) 查询 - 无锁
func (s *OptimizedPostService) GetBySlug(ctx context.Context, slug string) (map[string]interface{}, error) {
	if !s.loaded {
		return nil, errors.New("data not loaded")
	}
	
	idx, exists := s.slugIndex[slug]
	if !exists {
		return nil, errors.New("post not found")
	}
	
	// 直接返回，无锁开销
	return s.posts[idx], nil
}

// GetByID O(1) 查询 - 无锁
func (s *OptimizedPostService) GetByID(ctx context.Context, id int64) (map[string]interface{}, error) {
	if !s.loaded {
		return nil, errors.New("data not loaded")
	}
	
	idx, exists := s.idIndex[id]
	if !exists {
		return nil, errors.New("post not found")
	}
	
	return s.posts[idx], nil
}

// GetRecentPosts O(1) 分页查询 - 无锁
func (s *OptimizedPostService) GetRecentPosts(ctx context.Context, page, size int) ([]map[string]interface{}, error) {
	if !s.loaded {
		return []map[string]interface{}{}, nil
	}
	
	start := (page - 1) * size
	if start >= len(s.posts) {
		return []map[string]interface{}{}, nil
	}
	
	end := start + size
	if end > len(s.posts) {
		end = len(s.posts)
	}
	
	// 直接切片，无锁
	result := make([]map[string]interface{}, end-start)
	copy(result, s.posts[start:end])
	return result, nil
}

// Search O(n) 但只在内存中搜索 - 无锁
func (s *OptimizedPostService) Search(ctx context.Context, keyword string, page, size int) ([]map[string]interface{}, error) {
	if !s.loaded || keyword == "" {
		return []map[string]interface{}{}, nil
	}
	
	var results []map[string]interface{}
	for _, post := range s.posts {
		title, _ := post["title"].(string)
		if len(title) > 0 && len(keyword) > 0 && 
		   (len(title) >= len(keyword) && title[:len(keyword)] == keyword) {
			results = append(results, post)
		}
	}
	
	start := (page - 1) * size
	if start >= len(results) {
		return []map[string]interface{}{}, nil
	}
	
	end := start + size
	if end > len(results) {
		end = len(results)
	}
	
	return results[start:end], nil
}

// GetArchives O(n) 但只在内存中操作 - 无锁
func (s *OptimizedPostService) GetArchives(ctx context.Context) ([]map[string]interface{}, error) {
	if !s.loaded {
		return []map[string]interface{}{}, nil
	}
	
	archives := make([]map[string]interface{}, 0)
	archiveMap := make(map[string][]map[string]interface{})
	
	for _, post := range s.posts {
		created, _ := post["created_at"].(time.Time)
		year := created.Format("2006")
		month := created.Format("01")
		key := year + "-" + month
		archiveMap[key] = append(archiveMap[key], post)
	}
	
	for key, posts := range archiveMap {
		archives = append(archives, map[string]interface{}{
			"key":   key,
			"count": len(posts),
			"posts": posts,
		})
	}
	
	return archives, nil
}

// ListAdmin O(n) 过滤 - 无锁
func (s *OptimizedPostService) ListAdmin(ctx context.Context, page, size int, status, keyword string) ([]map[string]interface{}, int64, error) {
	if !s.loaded {
		return []map[string]interface{}{}, 0, nil
	}
	
	var filtered []map[string]interface{}
	for _, post := range s.posts {
		if status != "" && post["status"] != status {
			continue
		}
		if keyword != "" {
			title, _ := post["title"].(string)
			if len(title) == 0 || len(title) < len(keyword) || title[:len(keyword)] != keyword {
				continue
			}
		}
		filtered = append(filtered, post)
	}
	
	total := int64(len(filtered))
	start := (page - 1) * size
	if start >= len(filtered) {
		return []map[string]interface{}{}, total, nil
	}
	
	end := start + size
	if end > len(filtered) {
		end = len(filtered)
	}
	
	return filtered[start:end], total, nil
}

// Create 异步写入 - 非阻塞
func (s *OptimizedPostService) Create(ctx context.Context, title, content, slug, status string, category int64, tags []int64) (map[string]interface{}, error) {
	post := map[string]interface{}{
		"id":         time.Now().UnixNano(),
		"title":      title,
		"content":    content,
		"slug":       slug,
		"status":     status,
		"category":   category,
		"tags":       tags,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	
	// 立即返回结果，异步更新索引
	s.writeQueue <- &PostOperation{
		Type: "create",
		Data: post,
	}
	
	return post, nil
}

// Update 异步写入 - 非阻塞
func (s *OptimizedPostService) Update(ctx context.Context, id int64, title, content, slug, status string, category int64, tags []int64) (map[string]interface{}, error) {
	idx, exists := s.idIndex[id]
	if !exists {
		return nil, errors.New("post not found")
	}
	
	// 立即返回当前数据，异步更新
	post := s.posts[idx]
	post["title"] = title
	post["content"] = content
	post["slug"] = slug
	post["status"] = status
	post["category"] = category
	post["tags"] = tags
	post["updated_at"] = time.Now()
	
	s.writeQueue <- &PostOperation{
		Type: "update",
		Data: post,
	}
	
	return post, nil
}

// Delete 异步写入 - 非阻塞
func (s *OptimizedPostService) Delete(ctx context.Context, id int64) error {
	_, exists := s.idIndex[id]
	if !exists {
		return errors.New("post not found")
	}
	
	s.writeQueue <- &PostOperation{
		Type: "delete",
		Data: map[string]interface{}{"id": id},
	}
	
	return nil
}

// 同步操作（在后台goroutine中执行）
func (s *OptimizedPostService) createSync(data map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.posts = append([]map[string]interface{}{data}, s.posts...)
	s.slugIndex[data["slug"].(string)] = 0
	s.idIndex[data["id"].(int64)] = 0
	
	// 更新其他索引
	for i := range s.posts {
		s.slugIndex[s.posts[i]["slug"].(string)] = i
		s.idIndex[s.posts[i]["id"].(int64)] = i
	}
	
	return nil
}

func (s *OptimizedPostService) updateSync(data map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	id := data["id"].(int64)
	idx, exists := s.idIndex[id]
	if !exists {
		return errors.New("post not found")
	}
	
	s.slugIndex[data["slug"].(string)] = idx
	return nil
}

func (s *OptimizedPostService) deleteSync(data map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	id := data["id"].(int64)
	idx, exists := s.idIndex[id]
	if !exists {
		return errors.New("post not found")
	}
	
	slug := s.posts[idx]["slug"].(string)
	
	// 删除
	s.posts = append(s.posts[:idx], s.posts[idx+1:]...)
	delete(s.slugIndex, slug)
	delete(s.idIndex, id)
	
	// 更新索引
	for i := range s.posts {
		s.slugIndex[s.posts[i]["slug"].(string)] = i
		s.idIndex[s.posts[i]["id"].(int64)] = i
	}
	
	return nil
}

func (s *OptimizedPostService) GetByArchive(ctx context.Context, slug string) ([]map[string]interface{}, error) {
	return s.posts, nil
}
