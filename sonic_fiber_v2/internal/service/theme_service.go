package service

import (
	"context"
)

type ThemeServiceImpl struct{}

func NewThemeService() ThemeService {
	return &ThemeServiceImpl{}
}

func (s *ThemeServiceImpl) GetActivatedTheme(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"id":   1,
		"name": "default",
		"path": "/themes/default",
	}, nil
}
