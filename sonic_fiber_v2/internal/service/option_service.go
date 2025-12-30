package service

import (
	"context"
	"sync"
)

type OptionServiceImpl struct {
	mu      sync.RWMutex
	options map[string]interface{}
}

func NewOptionService() OptionService {
	return &OptionServiceImpl{
		options: map[string]interface{}{
			"site_name":        "Sonic Blog",
			"site_description": "A fast blog system",
			"archive_prefix":   "/archives",
			"category_prefix":  "/categories",
			"tag_prefix":       "/tags",
			"sheet_prefix":     "/sheets",
			"journal_prefix":   "/journals",
			"photo_prefix":     "/photos",
			"link_prefix":      "/links",
		},
	}
}

func (s *OptionServiceImpl) GetAll(ctx context.Context) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.options, nil
}

func (s *OptionServiceImpl) Save(ctx context.Context, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.options[key] = value
	return nil
}
