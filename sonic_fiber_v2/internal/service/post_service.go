package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

type PostServiceImpl struct {
	mu    sync.RWMutex
	posts []map[string]interface{}
	idSeq int64
}

func NewPostService() PostService {
	return &PostServiceImpl{
		posts: make([]map[string]interface{}, 0),
		idSeq: 1,
	}
}

func (s *PostServiceImpl) GetRecentPosts(ctx context.Context, page, size int) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := (page - 1) * size
	if start >= len(s.posts) {
		return []map[string]interface{}{}, nil
	}

	end := start + size
	if end > len(s.posts) {
		end = len(s.posts)
	}

	result := make([]map[string]interface{}, end-start)
	copy(result, s.posts[start:end])
	return result, nil
}

func (s *PostServiceImpl) GetBySlug(ctx context.Context, slug string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, post := range s.posts {
		if post["slug"] == slug {
			return post, nil
		}
	}
	return nil, errors.New("post not found")
}

func (s *PostServiceImpl) GetByID(ctx context.Context, id int64) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, post := range s.posts {
		if post["id"] == id {
			return post, nil
		}
	}
	return nil, errors.New("post not found")
}

func (s *PostServiceImpl) Search(ctx context.Context, keyword string, page, size int) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []map[string]interface{}
	for _, post := range s.posts {
		title, _ := post["title"].(string)
		content, _ := post["content"].(string)
		if len(title) > 0 && len(keyword) > 0 && 
		   (len(title) >= len(keyword) && title[:len(keyword)] == keyword ||
		    len(content) >= len(keyword) && content[:len(keyword)] == keyword) {
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

func (s *PostServiceImpl) GetArchives(ctx context.Context) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

func (s *PostServiceImpl) ListAdmin(ctx context.Context, page, size int, status, keyword string) ([]map[string]interface{}, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

func (s *PostServiceImpl) Create(ctx context.Context, title, content, slug, status string, category int64, tags []int64) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.idSeq++
	post := map[string]interface{}{
		"id":         s.idSeq,
		"title":      title,
		"content":    content,
		"slug":       slug,
		"status":     status,
		"category":   category,
		"tags":       tags,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	s.posts = append([]map[string]interface{}{post}, s.posts...)
	return post, nil
}

func (s *PostServiceImpl) Update(ctx context.Context, id int64, title, content, slug, status string, category int64, tags []int64) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, post := range s.posts {
		if post["id"] == id {
			s.posts[i]["title"] = title
			s.posts[i]["content"] = content
			s.posts[i]["slug"] = slug
			s.posts[i]["status"] = status
			s.posts[i]["category"] = category
			s.posts[i]["tags"] = tags
			s.posts[i]["updated_at"] = time.Now()
			return s.posts[i], nil
		}
	}
	return nil, errors.New("post not found")
}

func (s *PostServiceImpl) Delete(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, post := range s.posts {
		if post["id"] == id {
			s.posts = append(s.posts[:i], s.posts[i+1:]...)
			return nil
		}
	}
	return errors.New("post not found")
}

func (s *PostServiceImpl) GetByArchive(ctx context.Context, slug string) ([]map[string]interface{}, error) {
	// 简化实现，返回所有文章
	return s.posts, nil
}
