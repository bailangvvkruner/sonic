package service

import (
	"context"
	"errors"
	"sync"
)

type TagServiceImpl struct {
	mu    sync.RWMutex
	tags  []map[string]interface{}
	idSeq int64
}

func NewTagService() TagService {
	return &TagServiceImpl{
		tags:  make([]map[string]interface{}, 0),
		idSeq: 1,
	}
}

func (s *TagServiceImpl) List(ctx context.Context) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tags, nil
}

func (s *TagServiceImpl) GetBySlug(ctx context.Context, slug string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, tag := range s.tags {
		if tag["slug"] == slug {
			return tag, nil
		}
	}
	return nil, errors.New("tag not found")
}

func (s *TagServiceImpl) GetPosts(ctx context.Context, slug string, page, size int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (s *TagServiceImpl) ListAdmin(ctx context.Context) ([]map[string]interface{}, error) {
	return s.List(ctx)
}

func (s *TagServiceImpl) Create(ctx context.Context, name, slug string) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.idSeq++
	tag := map[string]interface{}{
		"id":   s.idSeq,
		"name": name,
		"slug": slug,
	}

	s.tags = append(s.tags, tag)
	return tag, nil
}

func (s *TagServiceImpl) Update(ctx context.Context, id int64, name, slug string) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, tag := range s.tags {
		if tag["id"] == id {
			s.tags[i]["name"] = name
			s.tags[i]["slug"] = slug
			return s.tags[i], nil
		}
	}
	return nil, errors.New("tag not found")
}

func (s *TagServiceImpl) Delete(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, tag := range s.tags {
		if tag["id"] == id {
			s.tags = append(s.tags[:i], s.tags[i+1:]...)
			return nil
		}
	}
	return errors.New("tag not found")
}
