package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

type CommentServiceImpl struct {
	mu       sync.RWMutex
	comments []map[string]interface{}
	idSeq    int64
}

func NewCommentService() CommentService {
	return &CommentServiceImpl{
		comments: make([]map[string]interface{}, 0),
		idSeq:    1,
	}
}

func (s *CommentServiceImpl) Create(ctx context.Context, postID int64, content, author, email string, parentID int64) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.idSeq++
	comment := map[string]interface{}{
		"id":         s.idSeq,
		"post_id":    postID,
		"content":    content,
		"author":     author,
		"email":      email,
		"parent_id":  parentID,
		"status":     "approved",
		"created_at": time.Now(),
	}

	s.comments = append(s.comments, comment)
	return comment, nil
}

func (s *CommentServiceImpl) ListAdmin(ctx context.Context, page, size int, postID int64) ([]map[string]interface{}, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []map[string]interface{}
	for _, comment := range s.comments {
		if postID != 0 && comment["post_id"] != postID {
			continue
		}
		filtered = append(filtered, comment)
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

func (s *CommentServiceImpl) UpdateStatus(ctx context.Context, id int64, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, comment := range s.comments {
		if comment["id"] == id {
			s.comments[i]["status"] = status
			return nil
		}
	}
	return errors.New("comment not found")
}

func (s *CommentServiceImpl) Delete(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, comment := range s.comments {
		if comment["id"] == id {
			s.comments = append(s.comments[:i], s.comments[i+1:]...)
			return nil
		}
	}
	return errors.New("comment not found")
}
