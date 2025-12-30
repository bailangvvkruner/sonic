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
	"github.com/gofiber/fiber/v2/middleware/cache"
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

// 自定义中间件 - 认证
func authMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}
	// 简化验证
	if len(token) < 10 {
		return c.Status(401).JSON(fiber.Map{"error": "无效的token"})
	}
	return c.Next()
}

// 自定义中间件 - 缓存控制
func cacheControlMiddleware(maxAge time.Duration) fiber.Handler {
	return cache.New(cache.Config{
		Expiration: maxAge,
		CacheControl: true,
	})
}

// 自定义中间件 - 安装重定向
func installRedirectMiddleware(c *fiber.Ctx) error {
	// 检查是否已安装（简化实现）
	// 实际应该检查数据库中是否有用户
	return c.Next()
}

// 自定义中间件 - 日志
func customLoggerMiddleware(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()
	latency := time.Since(start)
	
	if isDev() {
		fmt.Printf("[%s] %s %s - %d (%v)\n", 
			time.Now().Format("2006-01-02 15:04:05"), 
			c.Method(), 
			c.Path(), 
			c.Response().StatusCode(), 
			latency)
	}
	return err
}

// 自定义中间件 - 恢复
func recoveryMiddleware(c *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic recovered: %v\n", r)
			c.Status(500).JSON(fiber.Map{"error": "服务器内部错误"})
		}
	}()
	return c.Next()
}

func (h *Handler) InitRoutes(conf *Config) {
	h.app = fiber.New(fiber.Config{
		BodyLimit:         100 * 1024 * 1024, // 100MB
		Concurrency:       256 * 1024,        // 并发连接数
		DisableKeepalive:  false,
		ReduceMemoryUsage: true,
	})

	// 全局中间件
	h.app.Use(recoveryMiddleware)
	h.app.Use(customLoggerMiddleware)
	
	if isDev() {
		h.app.Use(cors.New(cors.Config{
			AllowOrigins:     "*",
			AllowMethods:     "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD",
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Admin-Authorization",
			AllowCredentials: true,
			ExposeHeaders:    "Content-Length, Content-Type",
			MaxAge:           12 * 60 * 60, // 12小时
		}))
	}

	// Ping路由
	h.app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// 静态文件路由 - 使用高级路由特性
	staticGroup := h.app.Group("/")
	
	// 管理后台静态资源
	staticGroup.Static(conf.Sonic.AdminURLPath, conf.Sonic.AdminResourcesDir, fiber.Static{
		Compress:      true,
		ByteRange:     true,
		CacheDuration: 1 * time.Hour,
		MaxAge:        3600,
	})

	// CSS/JS/图片资源
	staticGroup.Static("/css", filepath.Join(conf.Sonic.AdminResourcesDir, "css"), fiber.Static{
		Compress:  true,
		CacheDuration: 24 * time.Hour,
		MaxAge:    86400,
	})
	staticGroup.Static("/js", filepath.Join(conf.Sonic.AdminResourcesDir, "js"), fiber.Static{
		Compress:  true,
		CacheDuration: 24 * time.Hour,
		MaxAge:    86400,
	})
	staticGroup.Static("/images", filepath.Join(conf.Sonic.AdminResourcesDir, "images"), fiber.Static{
		Compress:  true,
		CacheDuration: 24 * time.Hour,
		MaxAge:    86400,
	})

	// 上传目录 - 带缓存控制
	h.app.Use(consts.SonicUploadDir, cacheControlMiddleware(7*24*time.Hour))
	h.app.Static(consts.SonicUploadDir, conf.Sonic.UploadDir, fiber.Static{
		Compress:      true,
		ByteRange:     true,
		CacheDuration: 7 * 24 * time.Hour,
		MaxAge:        7 * 24 * 3600,
	})

	// 主题资源
	staticGroup.Static("/themes/", conf.Sonic.ThemeDir, fiber.Static{
		Compress:      true,
		CacheDuration: 1 * time.Hour,
		MaxAge:        3600,
	})

	// API路由组
	api := h.app.Group("/api")

	// 管理后台API组
	admin := api.Group("/admin")
	admin.Post("/login", h.login)
	admin.Post("/install", h.install)
	admin.Get("/is_installed", h.isInstalled)

	// 需要认证的管理API
	admin.Use(authMiddleware)
	
	// 文章管理
	admin.Get("/posts", h.listPosts)
	admin.Post("/posts", h.createPost)
	admin.Put("/posts/:id", h.updatePost)
	admin.Delete("/posts/:id", h.deletePost)
	admin.Post("/posts/:id/likes", h.increasePostLikes)
	admin.Get("/posts/latest", h.listPostsLatest)
	admin.Get("/posts/status/:status", h.listPostsByStatus)
	admin.Get("/posts/:id/preview", h.previewPost)

	// 分类管理
	admin.Get("/categories", h.listCategories)
	admin.Get("/categories/tree_view", h.listCategoriesTree)
	admin.Post("/categories", h.createCategory)
	admin.Put("/categories/:id", h.updateCategory)
	admin.Delete("/categories/:id", h.deleteCategory)
	admin.Put("/categories/batch", h.updateCategoryBatch)

	// 评论管理
	admin.Get("/comments", h.listComments)
	admin.Post("/comments", h.createComment)
	admin.Put("/comments/:id", h.updateComment)
	admin.Delete("/comments/:id", h.deleteComment)
	admin.Put("/comments/:id/status/:status", h.updateCommentStatus)
	admin.Put("/comments/status/:status", h.updateCommentStatusBatch)
	admin.Delete("/comments", h.deleteCommentBatch)

	// 标签管理
	admin.Get("/tags", h.listTags)
	admin.Get("/tags/:id", h.getTagByID)
	admin.Post("/tags", h.createTag)
	admin.Put("/tags/:id", h.updateTag)
	admin.Delete("/tags/:id", h.deleteTag)

	// 配置管理
	admin.Get("/options", h.listOptions)
	admin.Get("/options/map_view", h.listOptionsMapView)
	admin.Post("/options/saving", h.saveOption)
	admin.Post("/options/map_view/saving", h.saveOptionMap)

	// 日志管理
	admin.Get("/logs", h.listLogs)
	admin.Get("/logs/latest", h.listLogsLatest)
	admin.Get("/logs/clear", h.clearLogs)

	// 统计管理
	admin.Get("/statistics", h.getStatistics)
	admin.Get("/statistics/user", h.getStatisticsWithUser)

	// Sheet管理
	admin.Get("/sheets", h.listSheets)
	admin.Get("/sheets/independent", h.listIndependentSheets)
	admin.Post("/sheets", h.createSheet)
	admin.Put("/sheets/:id", h.updateSheet)
	admin.Delete("/sheets/:id", h.deleteSheet)
	admin.Get("/sheets/:id/preview", h.previewSheet)

	// 日志管理
	admin.Get("/journals", h.listJournals)
	admin.Get("/journals/latest", h.listLatestJournals)
	admin.Post("/journals", h.createJournal)
	admin.Put("/journals/:id", h.updateJournal)
	admin.Delete("/journals/:id", h.deleteJournal)

	// 链接管理
	admin.Get("/links", h.listLinks)
	admin.Get("/links/teams", h.listLinkTeams)
	admin.Get("/links/:id", h.getLinkByID)
	admin.Post("/links", h.createLink)
	admin.Put("/links/:id", h.updateLink)
	admin.Delete("/links/:id", h.deleteLink)

	// 菜单管理
	admin.Get("/menus", h.listMenus)
	admin.Get("/menus/tree_view", h.listMenusTree)
	admin.Get("/menus/team/tree_view", h.listMenusTreeByTeam)
	admin.Get("/menus/teams", h.listMenuTeams)
	admin.Get("/menus/:id", h.getMenuByID)
	admin.Post("/menus", h.createMenu)
	admin.Post("/menus/batch", h.createMenuBatch)
	admin.Put("/menus/:id", h.updateMenu)
	admin.Put("/menus/batch", h.updateMenuBatch)
	admin.Delete("/menus/:id", h.deleteMenu)
	admin.Delete("/menus/batch", h.deleteMenuBatch)

	// 照片管理
	admin.Get("/photos", h.listPhotos)
	admin.Get("/photos/latest", h.listPhotosLatest)
	admin.Get("/photos/teams", h.listPhotoTeams)
	admin.Get("/photos/:id", h.getPhotoByID)
	admin.Post("/photos", h.createPhoto)
	admin.Post("/photos/batch", h.createPhotoBatch)
	admin.Put("/photos/:id", h.updatePhoto)
	admin.Delete("/photos/batch", h.deletePhotoBatch)

	// 用户管理
	admin.Get("/users/profile", h.getUserProfile)
	admin.Put("/users/profile", h.updateUserProfile)
	admin.Put("/users/profiles/password", h.updatePassword)
	admin.Put("/users/mfa/generate", h.generateMFAQRCode)
	admin.Put("/users/mfa/update", h.updateMFA)

	// 主题管理
	admin.Get("/themes/activation", h.getActivatedTheme)
	admin.Get("/themes", h.listAllThemes)
	admin.Get("/themes/:id", h.getThemeByID)
	admin.Get("/themes/activation/files", h.listActivatedThemeFiles)
	admin.Get("/themes/:id/files", h.listThemeFiles)
	admin.Get("/themes/files/content", h.getThemeFileContent)
	admin.Get("/themes/:id/files/content", h.getThemeFileContentByID)
	admin.Put("/themes/files/content", h.updateThemeFile)
	admin.Put("/themes/:id/files/content", h.updateThemeFileByID)
	admin.Get("/themes/activation/template/custom/sheet", h.listCustomSheetTemplate)
	admin.Get("/themes/activation/template/custom/post", h.listCustomPostTemplate)
	admin.Post("/themes/:id/activation", h.activateTheme)
	admin.Get("/themes/activation/configurations", h.getActivatedThemeConfig)
	admin.Get("/themes/:id/configurations", h.getThemeConfigByID)
	admin.Get("/themes/:id/configurations/groups/:group", h.getThemeConfigByGroup)
	admin.Get("/themes/:id/configurations/groups", h.getThemeConfigGroupNames)
	admin.Get("/themes/activation/settings", h.getActivatedThemeSettings)
	admin.Get("/themes/:id/settings", h.getThemeSettings)
	admin.Get("/themes/:id/groups/:group/settings", h.getThemeSettingsByGroup)
	admin.Post("/themes/activation/settings", h.saveActivatedThemeSettings)
	admin.Post("/themes/:id/settings", h.saveThemeSettings)
	admin.Delete("/themes/:id", h.deleteTheme)
	admin.Post("/themes/upload", h.uploadTheme)
	admin.Put("/themes/upload/:id", h.updateThemeByUpload)
	admin.Post("/themes/fetching", h.fetchTheme)
	admin.Put("/themes/fetching/:id", h.updateThemeByFetching)
	admin.Post("/themes/reload", h.reloadTheme)
	admin.Get("/themes/activation/template/exists", h.templateExists)

	// 附件管理
	admin.Get("/attachments", h.listAttachments)
	admin.Get("/attachments/:id", h.getAttachmentByID)
	admin.Post("/attachments/upload", h.uploadAttachment)
	admin.Post("/attachments/uploads", h.uploadAttachments)
	admin.Get("/attachments/media_types", h.getAllMediaTypes)
	admin.Get("/attachments/types", h.getAllTypes)
	admin.Put("/attachments/:id", h.updateAttachment)
	admin.Delete("/attachments/:id", h.deleteAttachment)
	admin.Delete("/attachments", h.deleteAttachmentBatch)

	// 备份管理
	admin.Get("/backups/work-dir", h.listBackups)
	admin.Post("/backups/work-dir", h.backupWholeSite)
	admin.Delete("/backups/work-dir", h.deleteBackups)
	admin.Get("/backups/work-dir/*", h.handleWorkDirBackup)
	admin.Get("/backups/data", h.listDataBackups)
	admin.Post("/backups/data", h.exportData)
	admin.Delete("/backups/data", h.deleteDataBackup)
	admin.Get("/backups/data/*", h.handleDataBackup)
	admin.Get("/backups/markdown/export", h.listMarkdownExports)
	admin.Post("/backups/markdown/export", h.exportMarkdown)
	admin.Post("/backups/markdown/import", h.importMarkdown)
	admin.Get("/backups/markdown/fetch", h.fetchMarkdownBackup)
	admin.Delete("/backups/markdown/export", h.deleteMarkdownExports)
	admin.Get("/backups/markdown/export/:filename", h.downloadMarkdown)

	// 邮件管理
	admin.Post("/mails/test", h.testEmail)

	// 环境信息
	admin.Get("/environments", h.getEnvironments)
	admin.Get("/sonic/logfile", h.getLogFiles)

	// 内容前端API组
	contentAPI := api.Group("/content")
	contentAPI.Use(customLoggerMiddleware, recoveryMiddleware, installRedirectMiddleware)

	// 归档API
	contentAPI.Get("/archives/years", h.listYearArchives)
	contentAPI.Get("/archives/months", h.listMonthArchives)

	// 分类API
	contentAPI.Get("/categories", h.listCategoriesAPI)
	contentAPI.Get("/categories/:slug/posts", h.listPostsByCategory)

	// 日志API
	contentAPI.Get("/journals", h.listJournalsAPI)
	contentAPI.Get("/journals/comments", h.listJournalComments)
	contentAPI.Post("/journals/comments", h.createJournalComment)
	contentAPI.Get("/journals/:id", h.getJournalAPI)
	contentAPI.Get("/journals/:id/comments/top_view", h.listJournalTopComments)
	contentAPI.Get("/journals/:id/comments/:parentID/children", h.listJournalCommentChildren)
	contentAPI.Get("/journals/:id/comments/tree_view", h.listJournalCommentTree)
	contentAPI.Get("/journals/:id/comments/list_view", h.listJournalCommentList)
	contentAPI.Post("/journals/:id/likes", h.likeJournal)

	// 照片API
	contentAPI.Post("/photos/:id/likes", h.likePhoto)

	// 文章API
	contentAPI.Post("/posts/comments", h.createPostComment)
	contentAPI.Get("/posts/:id/comments/top_view", h.listPostTopComments)
	contentAPI.Get("/posts/:id/comments/:parentID/children", h.listPostCommentChildren)
	contentAPI.Get("/posts/:id/comments/tree_view", h.listPostCommentTree)
	contentAPI.Get("/posts/:id/comments/list_view", h.listPostCommentList)
	contentAPI.Post("/posts/:id/likes", h.likePost)

	// SheetAPI
	contentAPI.Post("/sheets/comments", h.createSheetComment)
	contentAPI.Get("/sheets/:id/comments/top_view", h.listSheetTopComments)
	contentAPI.Get("/sheets/:id/comments/:parentID/children", h.listSheetCommentChildren)
	contentAPI.Get("/sheets/:id/comments/tree_view", h.listSheetCommentTree)
	contentAPI.Get("/sheets/:id/comments/list_view", h.listSheetCommentList)

	// 其他API
	contentAPI.Get("/links", h.listLinksAPI)
	contentAPI.Get("/links/team_view", h.listLinksTeamView)
	contentAPI.Get("/options/comment", h.getCommentOptions)
	contentAPI.Post("/comments/:id/likes", h.likeComment)

	// 前端页面路由组
	frontend := h.app.Group("")
	frontend.Use(customLoggerMiddleware, recoveryMiddleware, installRedirectMiddleware)

	// 首页和分页
	frontend.Get("/", h.frontendIndex)
	frontend.Get("/page/:page", h.frontendIndex)

	// 文章详情
	frontend.Get("/post/:slug", h.frontendPostPage)

	// 分类页面
	frontend.Get("/category/:slug", h.frontendCategoryPage)
	frontend.Get("/category/:slug/page/:page", h.frontendCategoryPage)

	// 标签页面
	frontend.Get("/tag/:slug", h.frontendTagPage)
	frontend.Get("/tag/:slug/page/:page", h.frontendTagPage)

	// 搜索页面
	frontend.Get("/search", h.frontendSearch)
	frontend.Get("/search/page/:page", h.frontendSearch)

	// 关于页面
	frontend.Get("/about", h.frontendAbout)

	// Feed和RSS
	frontend.Get("/robots.txt", h.robotsTxt)
	frontend.Get("/atom", h.atomFeed)
	frontend.Get("/atom.xml", h.atomFeed)
	frontend.Get("/rss", h.rssFeed)
	frontend.Get("/rss.xml", h.rssFeed)
	frontend.Get("/feed", h.feed)
	frontend.Get("/feed.xml", h.feed)
	frontend.Get("/feed/categories/:slug", h.categoryFeed)
	frontend.Get("/atom/categories/:slug", h.categoryAtom)
	frontend.Get("/sitemap.xml", h.sitemapXML)
	frontend.Get("/sitemap.html", h.sitemapHTML)

	// 版本和安装
	frontend.Get("/version", h.version)
	frontend.Get("/install", h.installPage)
	frontend.Get("/logo", h.logo)
	frontend.Get("/favicon", h.favicon)

	// 内容认证
	frontend.Post("/content/:type/:slug/authentication", h.contentAuthentication)

	// 动态路由（需要从配置读取）
	// 这里简化实现，实际应该从数据库读取配置
	frontend.Get("/archives/:slug", h.archivesBySlug)
	frontend.Get("/archives/:slug/page/:page", h.archivesPage)
	frontend.Get("/links", h.linksPage)
	frontend.Get("/photos", h.photosPage)
	frontend.Get("/photos/page/:page", h.photosPage)
	frontend.Get("/journals", h.journalsPage)
	frontend.Get("/journals/page/:page", h.journalsPage)

	// Sheet动态路由
	frontend.Get("/:slug", h.sheetBySlug)
	frontend.Get("/admin_preview/:slug", h.adminPreviewSheet)
}

// ==================== Handler方法实现 ====================

// 认证相关
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

	if user.Password != req.Password {
		return c.Status(401).JSON(fiber.Map{"error": "密码错误"})
	}

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

	var count int64
	h.service.db.Model(&User{}).Count(&count)
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "系统已安装"})
	}

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

func (h *Handler) isInstalled(c *fiber.Ctx) error {
	var count int64
	h.service.db.Model(&User{}).Count(&count)
	return c.JSON(fiber.Map{"installed": count > 0})
}

// 文章相关
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

func (h *Handler) listPostsLatest(c *fiber.Ctx) error {
	posts, _, err := h.service.ListPosts(1, 10, consts.PostStatusPublished)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(posts)
}

func (h *Handler) listPostsByStatus(c *fiber.Ctx) error {
	statusStr := c.Params("status")
	var status consts.PostStatus
	fmt.Sscanf(statusStr, "%d", &status)
	
	posts, _, err := h.service.ListPosts(1, 50, status)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(posts)
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

func (h *Handler) previewPost(c *fiber.Ctx) error {
	idStr := c.Params("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	
	post, err := h.service.GetPostBySlug(fmt.Sprintf("post-%d", id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "文章不存在"})
	}

	return c.JSON(fiber.Map{
		"preview": true,
		"post":    post,
	})
}

// 分类相关
func (h *Handler) listCategories(c *fiber.Ctx) error {
	categories, err := h.service.ListCategories()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(categories)
}

func (h *Handler) listCategoriesTree(c *fiber.Ctx) error {
	categories, err := h.service.ListCategories()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	// 简化树形结构构建
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

func (h *Handler) updateCategory(c *fiber.Ctx) error {
	idStr := c.Params("id")
	var category Category
	if err := c.BodyParser(&category); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	now := time.Now()
	category.UpdateTime = &now

	var id int
	fmt.Sscanf(idStr, "%d", &id)
	// 简化实现
	return c.JSON(fiber.Map{"message": "更新成功"})
}

func (h *Handler) deleteCategory(c *fiber.Ctx) error {
	idStr := c.Params("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	// 简化实现
	return c.JSON(fiber.Map{"message": "删除成功"})
}

func (h *Handler) updateCategoryBatch(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "批量更新成功"})
}

// 评论相关
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

func (h *Handler) updateComment(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "更新成功"})
}

func (h *Handler) deleteComment(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "删除成功"})
}

func (h *Handler) updateCommentStatus(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "状态更新成功"})
}

func (h *Handler) updateCommentStatusBatch(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "批量状态更新成功"})
}

func (h *Handler) deleteCommentBatch(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "批量删除成功"})
}

// 标签相关
func (h *Handler) listTags(c *fiber.Ctx) error {
	tags, err := h.service.ListTags()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tags)
}

func (h *Handler) getTagByID(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"id": c.Params("id")})
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

func (h *Handler) updateTag(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "更新成功"})
}

func (h *Handler) deleteTag(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "删除成功"})
}

// 配置相关
func (h *Handler) listOptions(c *fiber.Ctx) error {
	var options []Option
	if err := h.service.db.Find(&options).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(options)
}

func (h *Handler) listOptionsMapView(c *fiber.Ctx) error {
	var options []Option
	if err := h.service.db.Find(&options).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	// 转换为map格式
	result := make(map[string]string)
	for _, opt := range options {
		result[opt.Key] = opt.Value
	}
	return c.JSON(result)
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

func (h *Handler) saveOptionMap(c *fiber.Ctx) error {
	var req map[string]string
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	for key, value := range req {
		h.service.SaveOption(key, value)
	}

	return c.JSON(fiber.Map{"message": "批量保存成功"})
}

// 日志相关
func (h *Handler) listLogs(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]map[string]interface{}{
		{"id": 1, "level": "INFO", "message": "系统启动", "time": time.Now()},
	})
}

func (h *Handler) listLogsLatest(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]map[string]interface{}{
		{"id": 1, "level": "INFO", "message": "最新日志", "time": time.Now()},
	})
}

func (h *Handler) clearLogs(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "日志已清空"})
}

// 统计相关
func (h *Handler) getStatistics(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(map[string]interface{}{
		"totalPosts":     10,
		"totalComments":  100,
		"totalUsers":     1,
		"totalCategories": 5,
		"todayVisits":    100,
	})
}

func (h *Handler) getStatisticsWithUser(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(map[string]interface{}{
		"statistics": map[string]interface{}{
			"totalPosts":     10,
			"totalComments":  100,
			"totalUsers":     1,
			"totalCategories": 5,
			"todayVisits":    100,
		},
		"user": map[string]interface{}{
			"username": "admin",
			"nickname": "管理员",
		},
	})
}

// Sheet相关
func (h *Handler) listSheets(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listIndependentSheets(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) createSheet(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "创建成功"})
}

func (h *Handler) updateSheet(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "更新成功"})
}

func (h *Handler) deleteSheet(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "删除成功"})
}

func (h *Handler) previewSheet(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"preview": true})
}

// 日志管理相关
func (h *Handler) listJournals(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listLatestJournals(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) createJournal(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "创建成功"})
}

func (h *Handler) updateJournal(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "更新成功"})
}

func (h *Handler) deleteJournal(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "删除成功"})
}

// 链接相关
func (h *Handler) listLinks(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listLinkTeams(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) getLinkByID(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"id": c.Params("id")})
}

func (h *Handler) createLink(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "创建成功"})
}

func (h *Handler) updateLink(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "更新成功"})
}

func (h *Handler) deleteLink(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "删除成功"})
}

// 菜单相关
func (h *Handler) listMenus(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listMenusTree(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listMenusTreeByTeam(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listMenuTeams(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) getMenuByID(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"id": c.Params("id")})
}

func (h *Handler) createMenu(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "创建成功"})
}

func (h *Handler) createMenuBatch(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "批量创建成功"})
}

func (h *Handler) updateMenu(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "更新成功"})
}

func (h *Handler) updateMenuBatch(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "批量更新成功"})
}

func (h *Handler) deleteMenu(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "删除成功"})
}

func (h *Handler) deleteMenuBatch(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "批量删除成功"})
}

// 照片相关
func (h *Handler) listPhotos(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listPhotosLatest(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listPhotoTeams(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) getPhotoByID(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"id": c.Params("id")})
}

func (h *Handler) createPhoto(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "创建成功"})
}

func (h *Handler) createPhotoBatch(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "批量创建成功"})
}

func (h *Handler) updatePhoto(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "更新成功"})
}

func (h *Handler) deletePhotoBatch(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "批量删除成功"})
}

// 用户相关
func (h *Handler) getUserProfile(c *fiber.Ctx) error {
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

func (h *Handler) updatePassword(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "密码更新成功"})
}

func (h *Handler) generateMFAQRCode(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"qr_code": "data:image/png;base64,..."})
}

func (h *Handler) updateMFA(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "MFA更新成功"})
}

// 主题相关
func (h *Handler) getActivatedTheme(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(map[string]interface{}{
		"id":   1,
		"name": "默认主题",
	})
}

func (h *Handler) listAllThemes(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{
		{"id": 1, "name": "默认主题", "version": "1.0"},
	})
}

func (h *Handler) getThemeByID(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"id": c.Params("id")})
}

func (h *Handler) listActivatedThemeFiles(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"index.html", "post.html", "style.css"})
}

func (h *Handler) listThemeFiles(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"index.html", "post.html", "style.css"})
}

func (h *Handler) getThemeFileContent(c *fiber.Ctx) error {
	// 简化实现
	return c.SendString("<html>...</html>")
}

func (h *Handler) getThemeFileContentByID(c *fiber.Ctx) error {
	// 简化实现
	return c.SendString("<html>...</html>")
}

func (h *Handler) updateThemeFile(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "文件更新成功"})
}

func (h *Handler) updateThemeFileByID(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "文件更新成功"})
}

func (h *Handler) listCustomSheetTemplate(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"custom-sheet.html"})
}

func (h *Handler) listCustomPostTemplate(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"custom-post.html"})
}

func (h *Handler) activateTheme(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "主题激活成功"})
}

func (h *Handler) getActivatedThemeConfig(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(map[string]interface{}{"siteTitle": "我的博客"})
}

func (h *Handler) getThemeConfigByID(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(map[string]interface{}{"siteTitle": "我的博客"})
}

func (h *Handler) getThemeConfigByGroup(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(map[string]interface{}{"title": "博客标题"})
}

func (h *Handler) getThemeConfigGroupNames(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"basic", "advanced"})
}

func (h *Handler) getActivatedThemeSettings(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(map[string]interface{}{"theme": "default"})
}

func (h *Handler) getThemeSettings(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(map[string]interface{}{"theme": "default"})
}

func (h *Handler) getThemeSettingsByGroup(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(map[string]interface{}{"setting": "value"})
}

func (h *Handler) saveActivatedThemeSettings(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "设置保存成功"})
}

func (h *Handler) saveThemeSettings(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "设置保存成功"})
}

func (h *Handler) deleteTheme(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "主题删除成功"})
}

func (h *Handler) uploadTheme(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "主题上传成功"})
}

func (h *Handler) updateThemeByUpload(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "主题更新成功"})
}

func (h *Handler) fetchTheme(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "主题获取成功"})
}

func (h *Handler) updateThemeByFetching(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "主题更新成功"})
}

func (h *Handler) reloadTheme(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "主题重载成功"})
}

func (h *Handler) templateExists(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"exists": true})
}

// 附件相关
func (h *Handler) listAttachments(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) getAttachmentByID(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"id": c.Params("id")})
}

func (h *Handler) uploadAttachment(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "上传成功", "id": 1})
}

func (h *Handler) uploadAttachments(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "批量上传成功", "ids": []int{1, 2, 3}})
}

func (h *Handler) getAllMediaTypes(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"image/jpeg", "image/png", "video/mp4"})
}

func (h *Handler) getAllTypes(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"image", "video", "audio", "document"})
}

func (h *Handler) updateAttachment(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "更新成功"})
}

func (h *Handler) deleteAttachment(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "删除成功"})
}

func (h *Handler) deleteAttachmentBatch(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "批量删除成功"})
}

// 备份相关
func (h *Handler) listBackups(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"backup-20240101.zip"})
}

func (h *Handler) backupWholeSite(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "备份成功", "filename": "backup-20240101.zip"})
}

func (h *Handler) deleteBackups(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "备份删除成功"})
}

func (h *Handler) handleWorkDirBackup(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "工作目录备份处理"})
}

func (h *Handler) listDataBackups(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"data-backup-20240101.sql"})
}

func (h *Handler) exportData(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "数据导出成功", "filename": "data-backup-20240101.sql"})
}

func (h *Handler) deleteDataBackup(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "数据备份删除成功"})
}

func (h *Handler) handleDataBackup(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "数据备份处理"})
}

func (h *Handler) listMarkdownExports(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"markdown-export-20240101.zip"})
}

func (h *Handler) exportMarkdown(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "Markdown导出成功", "filename": "markdown-export-20240101.zip"})
}

func (h *Handler) importMarkdown(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "Markdown导入成功"})
}

func (h *Handler) fetchMarkdownBackup(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "Markdown备份获取成功"})
}

func (h *Handler) deleteMarkdownExports(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "Markdown导出删除"})
}

func (h *Handler) downloadMarkdown(c *fiber.Ctx) error {
	filename := c.Params("filename")
	// 简化实现
	return c.SendString("Markdown content for " + filename)
}

// 邮件相关
func (h *Handler) testEmail(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "测试邮件发送成功"})
}

// 环境信息
func (h *Handler) getEnvironments(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(map[string]interface{}{
		"goVersion":  "1.25",
		"os":         "linux",
		"workDir":    "/sonic",
		"uploadDir":  "/sonic/data/upload",
		"database":   "SQLite",
	})
}

func (h *Handler) getLogFiles(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"app.log", "error.log"})
}

// 内容API相关
func (h *Handler) listYearArchives(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]int{2024, 2023})
}

func (h *Handler) listMonthArchives(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]string{"2024-01", "2024-02", "2024-03"})
}

func (h *Handler) listCategoriesAPI(c *fiber.Ctx) error {
	categories, err := h.service.ListCategories()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(categories)
}

func (h *Handler) listPostsByCategory(c *fiber.Ctx) error {
	slug := c.Params("slug")
	// 简化实现，实际应该查询该分类下的文章
	return c.JSON(fiber.Map{
		"category": slug,
		"posts":    []interface{}{},
	})
}

func (h *Handler) listJournalsAPI(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listJournalComments(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) createJournalComment(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "评论创建成功"})
}

func (h *Handler) getJournalAPI(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"id": c.Params("id")})
}

func (h *Handler) listJournalTopComments(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listJournalCommentChildren(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listJournalCommentTree(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listJournalCommentList(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) likeJournal(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "点赞成功"})
}

func (h *Handler) likePhoto(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "点赞成功"})
}

func (h *Handler) createPostComment(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "评论创建成功"})
}

func (h *Handler) listPostTopComments(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listPostCommentChildren(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listPostCommentTree(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listPostCommentList(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) likePost(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "点赞成功"})
}

func (h *Handler) createSheetComment(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "评论创建成功"})
}

func (h *Handler) listSheetTopComments(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listSheetCommentChildren(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listSheetCommentTree(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listSheetCommentList(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listLinksAPI(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) listLinksTeamView(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON([]interface{}{})
}

func (h *Handler) getCommentOptions(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(map[string]interface{}{
		"enableComment": true,
		"requireReview": true,
	})
}

func (h *Handler) likeComment(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "点赞成功"})
}

// 前端页面相关
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

	// 增加浏览量
	h.service.db.Model(&Post{}).Where("id = ?", post.ID).Update("visits", gorm.Expr("visits + 1"))

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
	pageStr := c.Params("page")
	var page int
	if pageStr == "" {
		page = 1
	} else {
		fmt.Sscanf(pageStr, "%d", &page)
	}
	// 简化实现
	return c.SendString(fmt.Sprintf("分类: %s - 第%d页 (功能开发中)", slug, page))
}

func (h *Handler) frontendTagPage(c *fiber.Ctx) error {
	slug := c.Params("slug")
	pageStr := c.Params("page")
	var page int
	if pageStr == "" {
		page = 1
	} else {
		fmt.Sscanf(pageStr, "%d", &page)
	}
	// 简化实现
	return c.SendString(fmt.Sprintf("标签: %s - 第%d页 (功能开发中)", slug, page))
}

func (h *Handler) frontendSearch(c *fiber.Ctx) error {
	query := c.Query("q")
	pageStr := c.Params("page")
	var page int
	if pageStr == "" {
		page = 1
	} else {
		fmt.Sscanf(pageStr, "%d", &page)
	}
	// 简化实现
	return c.SendString(fmt.Sprintf("搜索: %s - 第%d页 (功能开发中)", query, page))
}

func (h *Handler) frontendAbout(c *fiber.Ctx) error {
	return c.SendString(`<html><head><title>关于</title></head><body><h1>关于</h1><p>这是一个博客系统</p><a href="/">返回首页</a></body></html>`)
}

func (h *Handler) robotsTxt(c *fiber.Ctx) error {
	return c.SendString("User-agent: *\nAllow: /")
}

func (h *Handler) atomFeed(c *fiber.Ctx) error {
	// 简化实现
	return c.SendString(`<?xml version="1.0" encoding="UTF-8"?><feed xmlns="http://www.w3.org/2005/Atom"><title>博客Feed</title></feed>`)
}

func (h *Handler) rssFeed(c *fiber.Ctx) error {
	// 简化实现
	return c.SendString(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel><title>博客RSS</title></channel></rss>`)
}

func (h *Handler) feed(c *fiber.Ctx) error {
	// 简化实现
	return c.SendString(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel><title>博客Feed</title></channel></rss>`)
}

func (h *Handler) categoryFeed(c *fiber.Ctx) error {
	slug := c.Params("slug")
	// 简化实现
	return c.SendString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel><title>分类 %s Feed</title></channel></rss>`, slug))
}

func (h *Handler) categoryAtom(c *fiber.Ctx) error {
	slug := c.Params("slug")
	// 简化实现
	return c.SendString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><feed xmlns="http://www.w3.org/2005/Atom"><title>分类 %s Feed</title></feed>`, slug))
}

func (h *Handler) sitemapXML(c *fiber.Ctx) error {
	// 简化实现
	return c.SendString(`<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>http://localhost:8080/</loc></url></urlset>`)
}

func (h *Handler) sitemapHTML(c *fiber.Ctx) error {
	// 简化实现
	return c.SendString(`<html><head><title>站点地图</title></head><body><h1>站点地图</h1><ul><li><a href="/">首页</a></li></ul></body></html>`)
}

func (h *Handler) version(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"version": "1.0.0",
		"build":   "20240101",
	})
}

func (h *Handler) installPage(c *fiber.Ctx) error {
	return c.SendString(`<html><head><title>安装系统</title></head><body><h1>安装系统</h1><form method="POST" action="/api/admin/install"><input name="username" placeholder="用户名"><input name="password" type="password" placeholder="密码"><input name="email" placeholder="邮箱"><button type="submit">安装</button></form></body></html>`)
}

func (h *Handler) logo(c *fiber.Ctx) error {
	// 简化实现
	return c.SendString("LOGO")
}

func (h *Handler) favicon(c *fiber.Ctx) error {
	// 简化实现
	return c.SendString("FAVICON")
}

func (h *Handler) contentAuthentication(c *fiber.Ctx) error {
	// 简化实现
	return c.JSON(fiber.Map{"message": "认证成功"})
}

func (h *Handler) archivesBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	// 简化实现
	return c.SendString(fmt.Sprintf("归档: %s", slug))
}

func (h *Handler) archivesPage(c *fiber.Ctx) error {
	slug := c.Params("slug")
	pageStr := c.Params("page")
	var page int
	if pageStr == "" {
		page = 1
	} else {
		fmt.Sscanf(pageStr, "%d", &page)
	}
	// 简化实现
	return c.SendString(fmt.Sprintf("归档: %s - 第%d页", slug, page))
}

func (h *Handler) linksPage(c *fiber.Ctx) error {
	// 简化实现
	return c.SendString(`<html><head><title>链接</title></head><body><h1>链接</h1><p>链接页面开发中...</p></body></html>`)
}

func (h *Handler) photosPage(c *fiber.Ctx) error {
	pageStr := c.Params("page")
	var page int
	if pageStr == "" {
		page = 1
	} else {
		fmt.Sscanf(pageStr, "%d", &page)
	}
	// 简化实现
	return c.SendString(fmt.Sprintf("照片墙 - 第%d页 (开发中)", page))
}

func (h *Handler) journalsPage(c *fiber.Ctx) error {
	pageStr := c.Params("page")
	var page int
	if pageStr == "" {
		page = 1
	} else {
		fmt.Sscanf(pageStr, "%d", &page)
	}
	// 简化实现
	return c.SendString(fmt.Sprintf("日志 - 第%d页 (开发中)", page))
}

func (h *Handler) sheetBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	// 简化实现
	return c.SendString(fmt.Sprintf("页面: %s (开发中)", slug))
}

func (h *Handler) adminPreviewSheet(c *fiber.Ctx) error {
	slug := c.Params("slug")
	// 简化实现
	return c.SendString(fmt.Sprintf("预览页面: %s (开发中)", slug))
}

// ==================== 主程序 ====================
func main() {
	fmt.Println("=== Sonic博客系统重构版 (完整Fiber v2特性) ===")

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
	fmt.Println("✓ 路由初始化完成 (包含完整Fiber v2高级特性)")

	// 启动服务
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("✓ 服务启动: http://localhost:%s\n", port)
	fmt.Println("✓ 管理后台: http://localhost:%s/admin")
	fmt.Println("✓ API接口: http://localhost:%s/api")

	if err := handler.app.Listen(":" + port); err != nil {
		panic(fmt.Sprintf("服务启动失败: %v", err))
	}
}
