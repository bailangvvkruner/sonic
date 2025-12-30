package service

import (
	"context"
	"io"
)

// PostService 文章服务接口
type PostService interface {
	GetRecentPosts(ctx context.Context, page, size int) ([]map[string]interface{}, error)
	GetBySlug(ctx context.Context, slug string) (map[string]interface{}, error)
	GetByID(ctx context.Context, id int64) (map[string]interface{}, error)
	Search(ctx context.Context, keyword string, page, size int) ([]map[string]interface{}, error)
	GetArchives(ctx context.Context) ([]map[string]interface{}, error)
	ListAdmin(ctx context.Context, page, size int, status, keyword string) ([]map[string]interface{}, int64, error)
	Create(ctx context.Context, title, content, slug, status string, category int64, tags []int64) (map[string]interface{}, error)
	Update(ctx context.Context, id int64, title, content, slug, status string, category int64, tags []int64) (map[string]interface{}, error)
	Delete(ctx context.Context, id int64) error
	GetByArchive(ctx context.Context, slug string) ([]map[string]interface{}, error)
}

// CategoryService 分类服务接口
type CategoryService interface {
	List(ctx context.Context) ([]map[string]interface{}, error)
	GetBySlug(ctx context.Context, slug string) (map[string]interface{}, error)
	GetPosts(ctx context.Context, slug string, page, size int) ([]map[string]interface{}, error)
	ListAdmin(ctx context.Context) ([]map[string]interface{}, error)
	Create(ctx context.Context, name, slug, description string, parentID int64) (map[string]interface{}, error)
	Update(ctx context.Context, id int64, name, slug, description string, parentID int64) (map[string]interface{}, error)
	Delete(ctx context.Context, id int64) error
}

// TagService 标签服务接口
type TagService interface {
	List(ctx context.Context) ([]map[string]interface{}, error)
	GetBySlug(ctx context.Context, slug string) (map[string]interface{}, error)
	GetPosts(ctx context.Context, slug string, page, size int) ([]map[string]interface{}, error)
	ListAdmin(ctx context.Context) ([]map[string]interface{}, error)
	Create(ctx context.Context, name, slug string) (map[string]interface{}, error)
	Update(ctx context.Context, id int64, name, slug string) (map[string]interface{}, error)
	Delete(ctx context.Context, id int64) error
}

// CommentService 评论服务接口
type CommentService interface {
	Create(ctx context.Context, postID int64, content, author, email string, parentID int64) (map[string]interface{}, error)
	ListAdmin(ctx context.Context, page, size int, postID int64) ([]map[string]interface{}, int64, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	Delete(ctx context.Context, id int64) error
}

// UserService 用户服务接口
type UserService interface {
	GetByID(ctx context.Context, id int64) (map[string]interface{}, error)
	UpdateProfile(ctx context.Context, id int64, nickname, email, avatar string) (map[string]interface{}, error)
	UpdatePassword(ctx context.Context, id int64, oldPassword, newPassword string) error
	Login(ctx context.Context, username, password string) (string, error)
	Install(ctx context.Context, username, password, email, siteName string) error
	VerifyToken(ctx context.Context, token string) (int64, error)
}

// ThemeService 主题服务接口
type ThemeService interface {
	GetActivatedTheme(ctx context.Context) (map[string]interface{}, error)
}

// AttachmentService 附件服务接口
type AttachmentService interface {
	Upload(ctx context.Context, filename string, src io.Reader, size int64) (map[string]interface{}, error)
	List(ctx context.Context, page, size int, keyword string) ([]map[string]interface{}, int64, error)
	Delete(ctx context.Context, id int64) error
}

// OptionService 配置服务接口
type OptionService interface {
	GetAll(ctx context.Context) (map[string]interface{}, error)
	Save(ctx context.Context, key string, value interface{}) error
}
