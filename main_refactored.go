package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/go-sonic/sonic/consts"
)

// ==================== 配置模块 ====================
type Config struct {
	Sonic   SonicConfig   `mapstructure:"sonic"`
	SQLite3 *SQLiteConfig `mapstructure:"sqlite3"`
}

type SonicConfig struct {
	Mode              string `mapstructure:"mode"`
	AdminURLPath      string `mapstructure:"admin_url_path"`
	WorkDir           string `mapstructure:"work_dir"`
	LogDir            string `mapstructure:"log_dir"`
	TemplateDir       string `mapstructure:"template_dir"`
	AdminResourcesDir string `mapstructure:"admin_resources_dir"`
	UploadDir         string `mapstructure:"upload_dir"`
	ThemeDir          string `mapstructure:"theme_dir"`
}

type SQLiteConfig struct {
	Enable bool   `mapstructure:"enable"`
	File   string `mapstructure:"file"`
}

func loadConfig() *Config {
	var configFile string
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigType("yaml")

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath("./conf/")
		viper.SetConfigName("config")
	}

	viper.SetDefault("sonic.admin_url_path", "admin")
	viper.SetDefault("sonic.mode", "development")

	conf := &Config{}
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("配置文件读取失败: %v", err))
	}
	if err := viper.Unmarshal(conf); err != nil {
		panic(fmt.Sprintf("配置解析失败: %v", err))
	}

	// 设置默认路径
	if conf.Sonic.WorkDir == "" {
		pwd, _ := os.Getwd()
		conf.Sonic.WorkDir, _ = filepath.Abs(pwd)
	}

	normalizeDir := func(path *string, subDir string) {
		if *path == "" {
			*path = filepath.Join(conf.Sonic.WorkDir, subDir)
		} else {
			temp, _ := filepath.Abs(*path)
			*path = temp
		}
	}

	normalizeDir(&conf.Sonic.LogDir, "log")
	normalizeDir(&conf.Sonic.TemplateDir, "resources/template")
	normalizeDir(&conf.Sonic.AdminResourcesDir, "resources/admin")
	normalizeDir(&conf.Sonic.UploadDir, consts.SonicUploadDir)
	normalizeDir(&conf.Sonic.ThemeDir, "resources/template/theme")

	if conf.SQLite3 != nil && conf.SQLite3.Enable {
		normalizeDir(&conf.SQLite3.File, "sonic.db")
	}

	// 创建必要目录
	os.MkdirAll(conf.Sonic.LogDir, os.ModePerm)
	os.MkdirAll(conf.Sonic.UploadDir, os.ModePerm)

	return conf
}

func isDev() bool {
	return viper.GetString("sonic.mode") == "development"
}

// ==================== 数据库模块 ====================
type DB struct {
	*gorm.DB
}

func NewDB(conf *Config) *DB {
	var db *gorm.DB
	var err error

	if conf.SQLite3 != nil && conf.SQLite3.Enable {
		db, err = gorm.Open(sqlite.Open(conf.SQLite3.File), &gorm.Config{})
	} else {
		panic("暂只支持SQLite数据库")
	}

	if err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(200)
	sqlDB.SetMaxOpenConns(300)
	sqlDB.SetConnMaxIdleTime(time.Hour)

	// 自动迁移表结构
	db.AutoMigrate(
		&Post{}, &Category{}, &Comment{}, &Tag{},
		&User{}, &Option{}, &Menu{}, &Link{},
		&Photo{}, &Journal{}, &Log{}, &Attachment{},
		&ThemeSetting{}, &PostCategory{}, &PostTag{},
	)

	return &DB{DB: db}
}

// ==================== 数据模型 ====================
type Base struct {
	ID         int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreateTime time.Time  `gorm:"not null" json:"create_time"`
	UpdateTime *time.Time `json:"update_time"`
}

type Post struct {
	Base
	Type            consts.PostType   `gorm:"index" json:"type"`
	Title           string            `gorm:"size:255;not null" json:"title"`
	Slug            string            `gorm:"size:255;uniqueIndex" json:"slug"`
	Content         string            `gorm:"type:text" json:"content"`
	Summary         string            `gorm:"type:text" json:"summary"`
	Status          consts.PostStatus `gorm:"default:1" json:"status"`
	Password        string            `gorm:"size:255" json:"password"`
	Thumbnail       string            `gorm:"size:1023" json:"thumbnail"`
	Visits          int64             `gorm:"default:0" json:"visits"`
	Likes           int64             `gorm:"default:0" json:"likes"`
	WordCount       int64             `gorm:"default:0" json:"word_count"`
	TopPriority     int32             `gorm:"default:0" json:"top_priority"`
	DisallowComment bool              `gorm:"default:false" json:"disallow_comment"`
	MetaKeywords    string            `gorm:"size:511" json:"meta_keywords"`
	MetaDescription string            `gorm:"size:1023" json:"meta_description"`
	Template        string            `gorm:"size:255" json:"template"`
	EditorType      consts.EditorType `gorm:"default:0" json:"editor_type"`
}

type Category struct {
	Base
	Name        string `gorm:"size:255;not null" json:"name"`
	Slug        string `gorm:"size:255;uniqueIndex" json:"slug"`
	Description string `gorm:"type:text" json:"description"`
	ParentID    int32  `gorm:"default:0" json:"parent_id"`
	Password    string `gorm:"size:255" json:"password"`
	Thumbnail   string `gorm:"size:1023" json:"thumbnail"`
}

type Comment struct {
	Base
	PostID    int32  `gorm:"index" json:"post_id"`
	ParentID  int32  `gorm:"default:0" json:"parent_id"`
	Author    string `gorm:"size:255;not null" json:"author"`
	Email     string `gorm:"size:255" json:"email"`
	Website   string `gorm:"size:255" json:"website"`
	Content   string `gorm:"type:text" json:"content"`
	Status    int    `gorm:"default:1" json:"status"`
	UserAgent string `gorm:"size:511" json:"user_agent"`
	IPAddress string `gorm:"size:64" json:"ip_address"`
	IsAdmin   bool   `gorm:"default:false" json:"is_admin"`
}

type Tag struct {
	Base
	Name        string `gorm:"size:255;not null" json:"name"`
	Slug        string `gorm:"size:255;uniqueIndex" json:"slug"`
	Description string `gorm:"type:text" json:"description"`
}

type User struct {
	Base
	Username    string     `gorm:"size:255;uniqueIndex;not null" json:"username"`
	Password    string     `gorm:"size:255;not null" json:"-"`
	Nickname    string     `gorm:"size:255" json:"nickname"`
	Email       string     `gorm:"size:255" json:"email"`
	Avatar      string     `gorm:"size:1023" json:"avatar"`
	Description string     `gorm:"type:text" json:"description"`
	MFAKey      string     `gorm:"size:255" json:"mfa_key"`
	MFAEnabled  bool       `gorm:"default:false" json:"mfa_enabled"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

type Option struct {
	Base
	Key   string `gorm:"size:255;uniqueIndex;not null" json:"key"`
	Value string `gorm:"type:text" json:"value"`
	Type  string `gorm:"size:50" json:"type"`
}

type Menu struct {
	Base
	Name     string `gorm:"size:255;not null" json:"name"`
	Url      string `gorm:"size:1023" json:"url"`
	Icon     string `gorm:"size:255" json:"icon"`
	ParentID int32  `gorm:"default:0" json:"parent_id"`
	Team     string `gorm:"size:50" json:"team"`
	Target   string `gorm:"size:20" json:"target"`
}

type Link struct {
	Base
	Name        string `gorm:"size:255;not null" json:"name"`
	Url         string `gorm:"size:1023;not null" json:"url"`
	Description string `gorm:"type:text" json:"description"`
	Logo        string `gorm:"size:1023" json:"logo"`
	Team        string `gorm:"size:50" json:"team"`
}

type Photo struct {
	Base
	Name        string     `gorm:"size:255;not null" json:"name"`
	Url         string     `gorm:"size:1023;not null" json:"url"`
	Description string     `gorm:"type:text" json:"description"`
	Location    string     `gorm:"size:255" json:"location"`
	TakeTime    *time.Time `json:"take_time"`
	Team        string     `gorm:"size:50" json:"team"`
}

type Journal struct {
	Base
	Content   string `gorm:"type:text;not null" json:"content"`
	Source    string `gorm:"size:50" json:"source"`
	Likes     int64  `gorm:"default:0" json:"likes"`
	IPAddress string `gorm:"size:64" json:"ip_address"`
	UserAgent string `gorm:"size:511" json:"user_agent"`
}

type Log struct {
	Base
	Level      string `gorm:"size:50;not null" json:"level"`
	Message    string `gorm:"type:text;not null" json:"message"`
	StackTrace string `gorm:"type:text" json:"stack_trace"`
}

type Attachment struct {
	Base
	Name      string `gorm:"size:255;not null" json:"name"`
	Path      string `gorm:"size:1023;not null" json:"path"`
	FileType  string `gorm:"size:50" json:"file_type"`
	FileSize  int64  `json:"file_size"`
	MediaType string `gorm:"size:100" json:"media_type"`
}

type ThemeSetting struct {
	Base
	ThemeID string `gorm:"size:255;not null" json:"theme_id"`
	Group   string `gorm:"size:255;not null" json:"group"`
	Items   string `gorm:"type:text" json:"items"`
}

type PostCategory struct {
	PostID     int32 `gorm:"primaryKey" json:"post_id"`
	CategoryID int32 `gorm:"primaryKey" json:"category_id"`
}

type PostTag struct {
	PostID int32 `gorm:"primaryKey" json:"post_id"`
	TagID  int32 `gorm:"primaryKey" json:"tag_id"`
}

// ==================== 服务层 ====================
type Service struct {
	db *DB
}

func NewService(db *DB) *Service {
	return &Service{db: db}
}

// 文章服务
func (s *Service) ListPosts(page, pageSize int, status consts.PostStatus) ([]*Post, int64, error) {
	var posts []*Post
	var count int64

	query := s.db.Model(&Post{})
	if status != 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&count)
	offset := (page - 1) * pageSize
	result := query.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&posts)
	return posts, count, result.Error
}

func (s *Service) GetPostBySlug(slug string) (*Post, error) {
	var post Post
	result := s.db.Where("slug = ?", slug).First(&post)
	return &post, result.Error
}

func (s *Service) CreatePost(post *Post) error {
	return s.db.Create(post).Error
}

func (s *Service) UpdatePost(id int32, post *Post) error {
	return s.db.Model(&Post{}).Where("id = ?", id).Updates(post).Error
}

func (s *Service) DeletePost(id int32) error {
	return s.db.Delete(&Post{}, id).Error
}

func (s *Service) IncreasePostLikes(id int32) error {
	return s.db.Model(&Post{}).Where("id = ?", id).Update("likes", gorm.Expr("likes + 1")).Error
}

// 分类服务
func (s *Service) ListCategories() ([]*Category, error) {
	var categories []*Category
	result := s.db.Order("create_time ASC").Find(&categories)
	return categories, result.Error
}

func (s *Service) CreateCategory(cat *Category) error {
	return s.db.Create(cat).Error
}

// 评论服务
func (s *Service) ListComments(postID int32) ([]*Comment, error) {
	var comments []*Comment
	query := s.db.Where("status = ?", 1).Order("create_time ASC")
	if postID > 0 {
		query = query.Where("post_id = ?", postID)
	}
	result := query.Find(&comments)
	return comments, result.Error
}

func (s *Service) CreateComment(comment *Comment) error {
	return s.db.Create(comment).Error
}

// 标签服务
func (s *Service) ListTags() ([]*Tag, error) {
	var tags []*Tag
	result := s.db.Order("create_time ASC").Find(&tags)
	return tags, result.Error
}

// 配置服务
func (s *Service) GetOption(key string) (string, error) {
	var option Option
	result := s.db.Where("key = ?", key).First(&option)
	if result.Error != nil {
		return "", result.Error
	}
	return option.Value, nil
}

func (s *Service) SaveOption(key, value string) error {
	option := Option{Key: key, Value: value}
	return s.db.Save(&option).Error
}

// 用户服务
func (s *Service) GetUserByUsername(username string) (*User, error) {
	var user User
	result := s.db.Where("username = ?", username).First(&user)
	return &user, result.Error
}

func (s *Service) UpdateUser(id int32, user *User) error {
	return s.db.Model(&User{}).Where("id = ?", id).Updates(user).Error
}

// ==================== Handler层 ====================
type Handler struct {
	service *Service
	app     *fiber.App
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitRoutes(conf *Config) {
	h.app = fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024, // 100MB
	})

	// 中间件
	h.app.Use(recover.New())
	h.app.Use(logger.New(logger.Config{
		Format: "${time} ${method} ${path} - ${status} ${latency}\n",
	}))

	if isDev() {
		h.app.Use(cors.New(cors.Config{
			AllowOrigins: "*",
			AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
			AllowHeaders: "Content-Type, Authorization",
		}))
	}

	// 静态文件
	h.app.Static("/admin", conf.Sonic.AdminResourcesDir)
	h.app.Static("/css", filepath.Join(conf.Sonic.AdminResourcesDir, "css"))
	h.app.Static("/js", filepath.Join(conf.Sonic.AdminResourcesDir, "js"))
	h.app.Static("/images", filepath.Join(conf.Sonic.AdminResourcesDir, "images"))
	h.app.Static(consts.SonicUploadDir, conf.Sonic.UploadDir)
	h.app.Static("/themes/", conf.Sonic.ThemeDir)

	// API路由
	api := h.app.Group("/api")

	// 管理后台API
	admin := api.Group("/admin")
	admin.Post("/login", h.login)
	admin.Post("/install", h.install)

	// 需要认证的路由
	admin.Use(h.authMiddleware)
	admin.Get("/posts", h.listPosts)
	admin.Post("/posts", h.createPost)
	admin.Put("/posts/:id", h.updatePost)
	admin.Delete("/posts/:id", h.deletePost)
	admin.Post("/posts/:id/likes", h.increasePostLikes)

	admin.Get("/categories", h.listCategories)
	admin.Post("/categories", h.createCategory)

	admin.Get("/comments", h.listComments)
	admin.Post("/comments", h.createComment)

	admin.Get("/tags", h.listTags)
	admin.Post("/tags", h.createTag)

	admin.Get("/options", h.listOptions)
	admin.Post("/options", h.saveOption)

	admin.Get("/users/profile", h.getUserProfile)
	admin.Put("/users/profile", h.updateUserProfile)

	// 内容前端API
	content := api.Group("/content")
	content.Get("/posts", h.frontendListPosts)
	content.Get("/posts/:slug", h.frontendGetPost)
	content.Get("/categories", h.frontendListCategories)
	content.Get("/tags", h.frontendListTags)
	content.Get("/comments/:postID", h.frontendListComments)
	content.Post("/comments", h.frontendCreateComment)
	content.Post("/posts/:slug/likes", h.frontendLikePost)

	// 前端页面路由
	h.app.Get("/", h.frontendIndex)
	h.app.Get("/page/:page", h.frontendIndex)
	h.app.Get("/post/:slug", h.frontendPostPage)
	h.app.Get("/category/:slug", h.frontendCategoryPage)
	h.app.Get("/tag/:slug", h.frontendTagPage)
	h.app.Get("/search", h.frontendSearch)
	h.app.Get("/about", h.frontendAbout)
}

// ==================== 认证中间件 ====================
func (h *Handler) authMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}

	// 简单的JWT验证（实际项目中需要完整实现）
	if !h.validateToken(token) {
		return c.Status(401).JSON(fiber.Map{"error": "无效的token"})
	}

	return c.Next()
}

func (h *Handler) validateToken(token string) bool {
	// 简化实现，实际需要完整的JWT验证
	return len(token) > 10
}

// ==================== Handler方法 ====================
func (h *Handler) login(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	user, err := h.service.GetUserByUsername(req.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "用户不存在或密码错误"})
	}

	// 简化密码验证（实际项目中需要加密验证）
	if user.Password != req.Password {
		return c.Status(401).JSON(fiber.Map{"error": "密码错误"})
	}

	// 生成token（简化实现）
	token := fmt.Sprintf("token_%d_%d", user.ID, time.Now().Unix())
	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
		},
	})
}

func (h *Handler) install(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	// 检查是否已安装
	var count int64
	h.service.db.Model(&User{}).Count(&count)
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "系统已安装"})
	}

	// 创建管理员用户
	user := &User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Nickname: "管理员",
	}
	if err := h.service.db.Create(user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "安装失败"})
	}

	return c.JSON(fiber.Map{"message": "安装成功"})
}

func (h *Handler) listPosts(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)
	status := consts.PostStatus(c.QueryInt("status", 0))

	posts, total, err := h.service.ListPosts(page, pageSize, status)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"list":  posts,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *Handler) createPost(c *fiber.Ctx) error {
	var post Post
	if err := c.BodyParser(&post); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	post.CreateTime = time.Now()
	if err := h.service.CreatePost(&post); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(post)
}

func (h *Handler) updatePost(c *fiber.Ctx) error {
	idStr := c.Params("id")
	var post Post
	if err := c.BodyParser(&post); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	now := time.Now()
	post.UpdateTime = &now

	var id int
	fmt.Sscanf(idStr, "%d", &id)
	if err := h.service.UpdatePost(int32(id), &post); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "更新成功"})
}

func (h *Handler) deletePost(c *fiber.Ctx) error {
	idStr := c.Params("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	if err := h.service.DeletePost(int32(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "删除成功"})
}

func (h *Handler) increasePostLikes(c *fiber.Ctx) error {
	idStr := c.Params("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	if err := h.service.IncreasePostLikes(int32(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "点赞成功"})
}

func (h *Handler) listCategories(c *fiber.Ctx) error {
	categories, err := h.service.ListCategories()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(categories)
}

func (h *Handler) createCategory(c *fiber.Ctx) error {
	var category Category
	if err := c.BodyParser(&category); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	category.CreateTime = time.Now()
	if err := h.service.CreateCategory(&category); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(category)
}

func (h *Handler) listComments(c *fiber.Ctx) error {
	postID := c.QueryInt("post_id", 0)
	comments, err := h.service.ListComments(int32(postID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(comments)
}

func (h *Handler) createComment(c *fiber.Ctx) error {
	var comment Comment
	if err := c.BodyParser(&comment); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	comment.CreateTime = time.Now()
	if err := h.service.CreateComment(&comment); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(comment)
}

func (h *Handler) listTags(c *fiber.Ctx) error {
	tags, err := h.service.ListTags()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tags)
}

func (h *Handler) createTag(c *fiber.Ctx) error {
	var tag Tag
	if err := c.BodyParser(&tag); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	tag.CreateTime = time.Now()
	if err := h.service.db.Create(&tag).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tag)
}

func (h *Handler) listOptions(c *fiber.Ctx) error {
	var options []Option
	if err := h.service.db.Find(&options).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(options)
}

func (h *Handler) saveOption(c *fiber.Ctx) error {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	if err := h.service.SaveOption(req.Key, req.Value); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "保存成功"})
}

func (h *Handler) getUserProfile(c *fiber.Ctx) error {
	// 从token中获取用户ID（简化实现）
	userID := int32(1)
	var user User
	result := h.service.db.Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "用户不存在"})
	}

	return c.JSON(fiber.Map{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"email":    user.Email,
		"avatar":   user.Avatar,
	})
}

func (h *Handler) updateUserProfile(c *fiber.Ctx) error {
	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Avatar   string `json:"avatar"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	userID := int32(1)
	updates := map[string]interface{}{
		"nickname": req.Nickname,
		"email":    req.Email,
		"avatar":   req.Avatar,
	}

	if err := h.service.db.Model(&User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "更新成功"})
}

// ==================== 前端Handler方法 ====================
func (h *Handler) frontendIndex(c *fiber.Ctx) error {
	pageStr := c.Params("page")
	var page int
	if pageStr == "" {
		page = 1
	} else {
		fmt.Sscanf(pageStr, "%d", &page)
	}
	posts, total, err := h.service.ListPosts(page, 10, consts.PostStatusPublished)
	if err != nil {
		return c.Status(500).SendString("服务器错误")
	}

	// 返回HTML（简化实现）
	html := `<html><head><title>博客首页</title></head><body><h1>博客文章</h1><ul>`
	for _, post := range posts {
		html += fmt.Sprintf(`<li><a href="/post/%s">%s</a> - %s</li>`, post.Slug, post.Title, post.CreateTime.Format("2006-01-02"))
	}
	html += `</ul>`
	if page > 1 {
		html += fmt.Sprintf(`<a href="/page/%d">上一页</a> `, page-1)
	}
	if int64(page*10) < total {
		html += fmt.Sprintf(`<a href="/page/%d">下一页</a>`, page+1)
	}
	html += `</body></html>`
	return c.SendString(html)
}

func (h *Handler) frontendPostPage(c *fiber.Ctx) error {
	slug := c.Params("slug")
	post, err := h.service.GetPostBySlug(slug)
	if err != nil {
		return c.Status(404).SendString("文章不存在")
	}

	html := fmt.Sprintf(`<html><head><title>%s</title></head><body>
		<h1>%s</h1>
		<p>发布时间: %s | 浏览: %d | 点赞: %d</p>
		<div>%s</div>
		<br>
		<a href="/">返回首页</a>
		</body></html>`, post.Title, post.Title, post.CreateTime.Format("2006-01-02"), post.Visits, post.Likes, post.Content)
	return c.SendString(html)
}

func (h *Handler) frontendCategoryPage(c *fiber.Ctx) error {
	slug := c.Params("slug")
	// 简化实现，实际需要关联查询
	return c.SendString(fmt.Sprintf("分类: %s (功能开发中)", slug))
}

func (h *Handler) frontendTagPage(c *fiber.Ctx) error {
	slug := c.Params("slug")
	return c.SendString(fmt.Sprintf("标签: %s (功能开发中)", slug))
}

func (h *Handler) frontendSearch(c *fiber.Ctx) error {
	query := c.Query("q")
	return c.SendString(fmt.Sprintf("搜索: %s (功能开发中)", query))
}

func (h *Handler) frontendAbout(c *fiber.Ctx) error {
	return c.SendString(`<html><head><title>关于</title></head><body><h1>关于</h1><p>这是一个博客系统</p><a href="/">返回首页</a></body></html>`)
}

func (h *Handler) frontendListPosts(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	posts, total, err := h.service.ListPosts(page, 10, consts.PostStatusPublished)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"list":  posts,
		"total": total,
		"page":  page,
	})
}

func (h *Handler) frontendGetPost(c *fiber.Ctx) error {
	slug := c.Params("slug")
	post, err := h.service.GetPostBySlug(slug)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "文章不存在"})
	}

	// 增加浏览量
	h.service.db.Model(&Post{}).Where("id = ?", post.ID).Update("visits", gorm.Expr("visits + 1"))

	return c.JSON(post)
}

func (h *Handler) frontendListCategories(c *fiber.Ctx) error {
	categories, err := h.service.ListCategories()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(categories)
}

func (h *Handler) frontendListTags(c *fiber.Ctx) error {
	tags, err := h.service.ListTags()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tags)
}

func (h *Handler) frontendListComments(c *fiber.Ctx) error {
	postIDStr := c.Params("postID")
	var postID int
	fmt.Sscanf(postIDStr, "%d", &postID)
	comments, err := h.service.ListComments(int32(postID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(comments)
}

func (h *Handler) frontendCreateComment(c *fiber.Ctx) error {
	var comment Comment
	if err := c.BodyParser(&comment); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	comment.CreateTime = time.Now()
	comment.Status = 1 // 待审核
	if err := h.service.CreateComment(&comment); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "评论提交成功，等待审核"})
}

func (h *Handler) frontendLikePost(c *fiber.Ctx) error {
	slug := c.Params("slug")
	post, err := h.service.GetPostBySlug(slug)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "文章不存在"})
	}

	if err := h.service.IncreasePostLikes(post.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "点赞成功"})
}

// ==================== 主程序 ====================
func main() {
	fmt.Println("=== Sonic博客系统重构版 ===")

	// 加载配置
	conf := loadConfig()
	fmt.Println("✓ 配置加载完成")

	// 初始化数据库
	db := NewDB(conf)
	fmt.Println("✓ 数据库初始化完成")

	// 初始化服务
	service := NewService(db)
	fmt.Println("✓ 服务初始化完成")

	// 初始化Handler
	handler := NewHandler(service)
	handler.InitRoutes(conf)
	fmt.Println("✓ 路由初始化完成")

	// 启动服务
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("✓ 服务启动: http://localhost:%s\n", port)
	fmt.Println("✓ 管理后台: http://localhost:%s/admin")
	fmt.Println("✓ API文档: http://localhost:%s/api")

	if err := handler.app.Listen(":" + port); err != nil {
		panic(fmt.Sprintf("服务启动失败: %v", err))
	}
}
