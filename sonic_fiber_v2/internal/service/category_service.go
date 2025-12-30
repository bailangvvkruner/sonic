package service

import (
	"context"
	"errors"
	"sync"
)

type CategoryServiceImpl struct {
	mu         sync.RWMutex
	categories []map[string]interface{}
	idSeq      int64
}

func NewCategoryService() CategoryService {
	return &CategoryServiceImpl{
		categories: make([]map[string]interface{}, 0),
		idSeq:      1,
	}
}

func (s *CategoryServiceImpl) List(ctx context.Context) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.categories, nil
}

func (s *CategoryServiceImpl) GetBySlug(ctx context.Context, slug string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, category := range s.categories {
		if category["slug"] == slug {
			return category, nil
		}
	}
	return nil, errors.New("category not found")
}

func (s *CategoryServiceImpl) GetPosts(ctx context.Context, slug string, page, size int) ([]map[string]interface{}, error) {
	// 简化实现，返回空数组
	return []map[string]interface{}{}, nil
}

func (s *CategoryServiceImpl) ListAdmin(ctx context.Context) ([]map[string]interface{}, error) {
	return s.List(ctx)
}

func (s *CategoryServiceImpl) Create(ctx context.Context, name, slug, description string, parentID int64) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.idSeq++
	category := map[string]interface{}{
		"id":          s.idSeq,
		"name":        name,
		"slug":        slug,
		"description": description,
		"parent_id":   parentID,
	}

	s.categories = append(s.categories, category)
	return category, nil
}

func (s *CategoryServiceImpl) Update(ctx context.Context, id int64, name, slug, description string, parentID int64) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, category := range s.categories {
		if category["id"] == id {
			s.categories[i]["name"] = name
			s.categories[i]["slug"] = slug
			s.categories[i]["description"] = description
			s.categories[i]["parent_id"] = parentID
			return s.categories[i], nil
		}
	}
	return nil, errors.New("category not found")
}

func (s *CategoryServiceImpl) Delete(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, category := range s.categories {
		if category["id"] == id {
			s.categories = append(s.categories[:i], s.categories[i+1:]...)
			return nil
		}
	}
	return errors.New("category not found")
}
