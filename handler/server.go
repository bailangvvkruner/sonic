package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/dig"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/go-sonic/sonic/config"
	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/dal"
	"github.com/go-sonic/sonic/event"
	"github.com/go-sonic/sonic/handler/admin"
	"github.com/go-sonic/sonic/handler/content"
	"github.com/go-sonic/sonic/handler/content/api"
	"github.com/go-sonic/sonic/handler/middleware"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/template"
	"github.com/go-sonic/sonic/util/xerr"
)

type Server struct {
	logger                    *zap.Logger
	Config                    *config.Config
	HTTPServer                *http.Server
	Router                    *fiber.App
	Template                  *template.Template
	AuthMiddleware            *middleware.AuthMiddleware
	LogMiddleware             *middleware.FiberLoggerMiddleware
	RecoveryMiddleware        *middleware.RecoveryMiddleware
	InstallRedirectMiddleware *middleware.InstallRedirectMiddleware
	OptionService             service.OptionService
	ThemeService              service.ThemeService
	SheetService              service.SheetService
	AdminHandler              *admin.AdminHandler
	AttachmentHandler         *admin.AttachmentHandler
	BackupHandler             *admin.BackupHandler
	CategoryHandler           *admin.CategoryHandler
	InstallHandler            *admin.InstallHandler
	JournalHandler            *admin.JournalHandler
	JournalCommentHandler     *admin.JournalCommentHandler
	LinkHandler               *admin.LinkHandler
	LogHandler                *admin.LogHandler
	MenuHandler               *admin.MenuHandler
	OptionHandler             *admin.OptionHandler
	PhotoHandler              *admin.PhotoHandler
	PostHandler               *admin.PostHandler
	PostCommentHandler        *admin.PostCommentHandler
	SheetHandler              *admin.SheetHandler
	SheetCommentHandler       *admin.SheetCommentHandler
	StatisticHandler          *admin.StatisticHandler
	TagHandler                *admin.TagHandler
	ThemeHandler              *admin.ThemeHandler
	UserHandler               *admin.UserHandler
	EmailHandler              *admin.EmailHandler
	IndexHandler              *content.IndexHandler
	FeedHandler               *content.FeedHandler
	ArchiveHandler            *content.ArchiveHandler
	ViewHandler               *content.ViewHandler
	ContentCategoryHandler    *content.CategoryHandler
	ContentSheetHandler       *content.SheetHandler
	ContentTagHandler         *content.TagHandler
	ContentLinkHandler        *content.LinkHandler
	ContentPhotoHandler       *content.PhotoHandler
	ContentJournalHandler     *content.JournalHandler
	ContentSearchHandler      *content.SearchHandler
	ContentAPIArchiveHandler  *api.ArchiveHandler
	ContentAPICategoryHandler *api.CategoryHandler
	ContentAPIJournalHandler  *api.JournalHandler
	ContentAPILinkHandler     *api.LinkHandler
	ContentAPIPostHandler     *api.PostHandler
	ContentAPISheetHandler    *api.SheetHandler
	ContentAPIOptionHandler   *api.OptionHandler
	ContentAPIPhotoHandler    *api.PhotoHandler
	ContentAPICommentHandler  *api.CommentHandler
}

type ServerParams struct {
	dig.In
	Config                    *config.Config
	Logger                    *zap.Logger
	Event                     event.Bus
	Template                  *template.Template
	AuthMiddleware            *middleware.AuthMiddleware
	LogMiddleware             *middleware.FiberLoggerMiddleware
	RecoveryMiddleware        *middleware.RecoveryMiddleware
	InstallRedirectMiddleware *middleware.InstallRedirectMiddleware
	OptionService             service.OptionService
	ThemeService              service.ThemeService
	SheetService              service.SheetService
	AdminHandler              *admin.AdminHandler
	AttachmentHandler         *admin.AttachmentHandler
	BackupHandler             *admin.BackupHandler
	CategoryHandler           *admin.CategoryHandler
	InstallHandler            *admin.InstallHandler
	JournalHandler            *admin.JournalHandler
	JournalCommentHandler     *admin.JournalCommentHandler
	LinkHandler               *admin.LinkHandler
	LogHandler                *admin.LogHandler
	MenuHandler               *admin.MenuHandler
	OptionHandler             *admin.OptionHandler
	PhotoHandler              *admin.PhotoHandler
	PostHandler               *admin.PostHandler
	PostCommentHandler        *admin.PostCommentHandler
	SheetHandler              *admin.SheetHandler
	SheetCommentHandler       *admin.SheetCommentHandler
	StatisticHandler          *admin.StatisticHandler
	TagHandler                *admin.TagHandler
	ThemeHandler              *admin.ThemeHandler
	UserHandler               *admin.UserHandler
	EmailHandler              *admin.EmailHandler
	IndexHandler              *content.IndexHandler
	FeedHandler               *content.FeedHandler
	ArchiveHandler            *content.ArchiveHandler
	ViewHandler               *content.ViewHandler
	ContentCategoryHandler    *content.CategoryHandler
	ContentSheetHandler       *content.SheetHandler
	ContentTagHandler         *content.TagHandler
	ContentLinkHandler        *content.LinkHandler
	ContentPhotoHandler       *content.PhotoHandler
	ContentJournalHandler     *content.JournalHandler
	ContentSearchHandler      *content.SearchHandler
	ContentAPIArchiveHandler  *api.ArchiveHandler
	ContentAPICategoryHandler *api.CategoryHandler
	ContentAPIJournalHandler  *api.JournalHandler
	ContentAPILinkHandler     *api.LinkHandler
	ContentAPIPostHandler     *api.PostHandler
	ContentAPISheetHandler    *api.SheetHandler
	ContentAPIOptionHandler   *api.OptionHandler
	ContentAPIPhotoHandler    *api.PhotoHandler
	ContentAPICommentHandler  *api.CommentHandler
}

func NewServer(param ServerParams, lifecycle fx.Lifecycle) *Server {
	conf := param.Config
	router := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(&dto.BaseDTO{
				Status:  code,
				Message: err.Error(),
			})
		},
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", conf.Server.Host, conf.Server.Port),
		Handler: router,
	}

	s := &Server{
		logger:                    param.Logger,
		Config:                    param.Config,
		HTTPServer:                httpServer,
		Router:                    router,
		Template:                  param.Template,
		AuthMiddleware:            param.AuthMiddleware,
		LogMiddleware:             param.LogMiddleware,
		RecoveryMiddleware:        param.RecoveryMiddleware,
		InstallRedirectMiddleware: param.InstallRedirectMiddleware,
		AdminHandler:              param.AdminHandler,
		AttachmentHandler:         param.AttachmentHandler,
		BackupHandler:             param.BackupHandler,
		CategoryHandler:           param.CategoryHandler,
		InstallHandler:            param.InstallHandler,
		JournalHandler:            param.JournalHandler,
		JournalCommentHandler:     param.JournalCommentHandler,
		LinkHandler:               param.LinkHandler,
		LogHandler:                param.LogHandler,
		MenuHandler:               param.MenuHandler,
		OptionHandler:             param.OptionHandler,
		PhotoHandler:              param.PhotoHandler,
		PostHandler:               param.PostHandler,
		PostCommentHandler:        param.PostCommentHandler,
		SheetHandler:              param.SheetHandler,
		SheetCommentHandler:       param.SheetCommentHandler,
		StatisticHandler:          param.StatisticHandler,
		TagHandler:                param.TagHandler,
		ThemeHandler:              param.ThemeHandler,
		UserHandler:               param.UserHandler,
		EmailHandler:              param.EmailHandler,
		OptionService:             param.OptionService,
		ThemeService:              param.ThemeService,
		SheetService:              param.SheetService,
		IndexHandler:              param.IndexHandler,
		FeedHandler:               param.FeedHandler,
		ArchiveHandler:            param.ArchiveHandler,
		ViewHandler:               param.ViewHandler,
		ContentCategoryHandler:    param.ContentCategoryHandler,
		ContentSheetHandler:       param.ContentSheetHandler,
		ContentTagHandler:         param.ContentTagHandler,
		ContentLinkHandler:        param.ContentLinkHandler,
		ContentPhotoHandler:       param.ContentPhotoHandler,
		ContentJournalHandler:     param.ContentJournalHandler,
		ContentAPIArchiveHandler:  param.ContentAPIArchiveHandler,
		ContentAPICategoryHandler: param.ContentAPICategoryHandler,
		ContentAPIJournalHandler:  param.ContentAPIJournalHandler,
		ContentAPILinkHandler:     param.ContentAPILinkHandler,
		ContentAPIPostHandler:     param.ContentAPIPostHandler,
		ContentAPISheetHandler:    param.ContentAPISheetHandler,
		ContentAPIOptionHandler:   param.ContentAPIOptionHandler,
		ContentSearchHandler:      param.ContentSearchHandler,
		ContentAPIPhotoHandler:    param.ContentAPIPhotoHandler,
		ContentAPICommentHandler:  param.ContentAPICommentHandler,
	})
	lifecycle.Append(fx.Hook{
		OnStop:  httpServer.Shutdown,
		OnStart: s.Run,
	})
	return s
}

func (s *Server) Run(ctx context.Context) error {
	go func() {
		if err := s.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// print err info when httpServer start failed
			s.logger.Error("unexpected error from ListenAndServe", zap.Error(err))
			fmt.Printf("http server start error:%s\n", err.Error())
			os.Exit(1)
		}
	}()
	return nil
}

type wrapperHandler func(ctx *fiber.Ctx) (interface{}, error)

func (s *Server) wrapHandler(handler wrapperHandler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		data, err := handler(ctx)
		if err != nil {
			s.logger.Error("handler error", zap.Error(err))
			status := xerr.GetHTTPStatus(err)
			return ctx.Status(status).JSON(&dto.BaseDTO{Status: status, Message: xerr.GetMessage(err)})
		}

		return ctx.Status(http.StatusOK).JSON(&dto.BaseDTO{
			Status:  http.StatusOK,
			Data:    data,
			Message: "OK",
		})
	}
}

type wrapperHTMLHandler func(ctx *fiber.Ctx, model template.Model) (templateName string, err error)

var (
	htmlContentType = "text/html; charset=utf-8"
	xmlContentType  = "application/xml; charset=utf-8"
)

func (s *Server) wrapHTMLHandler(handler wrapperHTMLHandler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		model := template.Model{}
		templateName, err := handler(ctx, model)
		if err != nil {
			return s.handleError(ctx, err)
		}
		if templateName == "" {
			return nil
		}
		if ctx.Get("Content-Type") == "" {
			ctx.Set("Content-Type", htmlContentType)
		}
		return s.Template.ExecuteTemplate(ctx.Response().BodyWriter(), templateName, model)
	}
}

func (s *Server) wrapTextHandler(handler wrapperHTMLHandler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		model := template.Model{}
		templateName, err := handler(ctx, model)
		if err != nil {
			return s.handleError(ctx, err)
		}
		if ctx.Get("Content-Type") == "" {
			ctx.Set("Content-Type", xmlContentType)
		}
		return s.Template.ExecuteTextTemplate(ctx.Response().BodyWriter(), templateName, model)
	}
}

func (s *Server) handleError(ctx *fiber.Ctx, err error) error {
	status := xerr.GetHTTPStatus(err)
	message := xerr.GetMessage(err)
	model := template.Model{}

	templateName, _ := s.ThemeService.Render(ctx, strconv.Itoa(status))
	t := s.Template.HTMLTemplate.Lookup(templateName)
	if t == nil {
		templateName = "common/error/error"
	}

	if ctx.Get("Content-Type") == "" {
		ctx.Set("Content-Type", htmlContentType)
	}

	model["status"] = status
	model["message"] = message
	model["err"] = err

	return s.Template.ExecuteTemplate(ctx.Response().BodyWriter(), templateName, model)
}

// RegisterRouters 注册所有路由
func (s *Server) RegisterRouters() {
	router := s.Router

	// 配置开发环境CORS
	if config.IsDev() {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"PUT", "PATCH", "GET", "DELETE", "POST", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Admin-Authorization", "Content-Type"},
			AllowCredentials: true,
			ExposeHeaders:    []string{"Content-Length"},
		}))
	}

	// 健康检查
	router.GET("/ping", func(ctx *fiber.Ctx) error {
		return ctx.SendString("pong")
	})

	// 静态文件路由
	s.registerStaticRoutes(router)

	// 管理员API路由
	s.registerAdminAPIRoutes(router)

	// 内容路由
	s.registerContentRoutes(router)

	// 内容API路由
	s.registerContentAPIRoutes(router)
}

// registerStaticRoutes 注册静态文件路由
func (s *Server) registerStaticRoutes(router fiber.Router) {
	staticRouter := router.Group("/")
	staticRouter.Static(s.Config.Sonic.AdminURLPath, s.Config.Sonic.AdminResourcesDir)
	staticRouter.Static("/css", filepath.Join(s.Config.Sonic.AdminResourcesDir, "css"))
	staticRouter.Static("/js", filepath.Join(s.Config.Sonic.AdminResourcesDir, "js"))
	staticRouter.Static("/images", filepath.Join(s.Config.Sonic.AdminResourcesDir, "images"))
	staticRouter.Use(middleware.NewCacheControlMiddleware(middleware.WithMaxAge(time.Hour*24*7)).CacheControl()).
		Static(consts.SonicUploadDir, s.Config.Sonic.UploadDir)
	staticRouter.Static("/themes/", s.Config.Sonic.ThemeDir)
}

// registerAdminAPIRoutes 注册管理员API路由
func (s *Server) registerAdminAPIRoutes(router fiber.Router) {
	adminAPIRouter := router.Group("/api/admin")
	adminAPIRouter.Use(
		s.LogMiddleware.LoggerWithConfig(middleware.FiberLoggerConfig{}),
		s.RecoveryMiddleware.RecoveryWithLogger(),
		s.InstallRedirectMiddleware.InstallRedirect(),
	)

	// 未认证路由
	adminAPIRouter.GET("/is_installed", s.wrapHandler(s.AdminHandler.IsInstalled))
	adminAPIRouter.POST("/login/precheck", s.wrapHandler(s.AdminHandler.AuthPreCheck))
	adminAPIRouter.POST("/login", s.wrapHandler(s.AdminHandler.Auth))
	adminAPIRouter.POST("/refresh/:refreshToken", s.wrapHandler(s.AdminHandler.RefreshToken))
	adminAPIRouter.POST("/installations", s.wrapHandler(s.InstallHandler.InstallBlog))

	// 认证路由
	authRouter := adminAPIRouter.Group("")
	authRouter.Use(s.AuthMiddleware.GetWrapHandler())

	// 基础认证路由
	authRouter.POST("/logout", s.wrapHandler(s.AdminHandler.LogOut))
	authRouter.POST("/password/code", s.wrapHandler(s.AdminHandler.SendResetCode))
	authRouter.GET("/environments", s.wrapHandler(s.AdminHandler.GetEnvironments))
	authRouter.GET("/sonic/logfile", s.wrapHandler(s.AdminHandler.GetLogFiles))

	// 附件路由
	s.registerAttachmentRoutes(authRouter)

	// 备份路由
	s.registerBackupRoutes(authRouter)

	// 分类路由
	s.registerCategoryRoutes(authRouter)

	// 文章路由
	s.registerPostRoutes(authRouter)

	// 选项路由
	s.registerOptionRoutes(authRouter)

	// 日志路由
	s.registerLogRoutes(authRouter)

	// 统计路由
	s.registerStatisticRoutes(authRouter)

	// 页面路由
	s.registerSheetRoutes(authRouter)

	// 日记路由
	s.registerJournalRoutes(authRouter)

	// 链接路由
	s.registerLinkRoutes(authRouter)

	// 菜单路由
	s.registerMenuRoutes(authRouter)

	// 标签路由
	s.registerTagRoutes(authRouter)

	// 照片路由
	s.registerPhotoRoutes(authRouter)

	// 用户路由
	s.registerUserRoutes(authRouter)

	// 主题路由
	s.registerThemeRoutes(authRouter)

	// 邮件路由
	s.registerEmailRoutes(authRouter)
}

// registerAttachmentRoutes 注册附件路由
func (s *Server) registerAttachmentRoutes(router fiber.Router) {
	attachmentRouter := router.Group("/attachments")
	attachmentRouter.POST("/upload", s.wrapHandler(s.AttachmentHandler.UploadAttachment))
	attachmentRouter.POST("/uploads", s.wrapHandler(s.AttachmentHandler.UploadAttachments))
	attachmentRouter.DELETE("/:id", s.wrapHandler(s.AttachmentHandler.DeleteAttachment))
	attachmentRouter.DELETE("", s.wrapHandler(s.AttachmentHandler.DeleteAttachmentInBatch))
	attachmentRouter.GET("", s.wrapHandler(s.AttachmentHandler.QueryAttachment))
	attachmentRouter.GET("/:id", s.wrapHandler(s.AttachmentHandler.GetAttachmentByID))
	attachmentRouter.PUT("/:id", s.wrapHandler(s.AttachmentHandler.UpdateAttachment))
	attachmentRouter.GET("/media_types", s.wrapHandler(s.AttachmentHandler.GetAllMediaType))
	attachmentRouter.GET("types", s.wrapHandler(s.AttachmentHandler.GetAllTypes))
}

// registerBackupRoutes 注册备份路由
func (s *Server) registerBackupRoutes(router fiber.Router) {
	backupRouter := router.Group("/backups")
	backupRouter.POST("/work-dir", s.wrapHandler(s.BackupHandler.BackupWholeSite))
	backupRouter.GET("/work-dir", s.wrapHandler(s.BackupHandler.ListBackups))
	backupRouter.GET("/work-dir/*path", s.BackupHandler.HandleWorkDir)
	backupRouter.DELETE("/work-dir", s.wrapHandler(s.BackupHandler.DeleteBackups))
	backupRouter.POST("/data", s.wrapHandler(s.BackupHandler.ExportData))
	backupRouter.DELETE("/data", s.wrapHandler(s.BackupHandler.DeleteDataFile))
	backupRouter.GET("/data/*path", s.BackupHandler.HandleData)
	backupRouter.POST("/markdown/export", s.wrapHandler(s.BackupHandler.ExportMarkdown))
	backupRouter.POST("/markdown/import", s.wrapHandler(s.BackupHandler.ImportMarkdown))
	backupRouter.GET("/markdown/fetch", s.wrapHandler(s.BackupHandler.GetMarkDownBackup))
	backupRouter.GET("/markdown/export", s.wrapHandler(s.BackupHandler.ListMarkdowns))
	backupRouter.DELETE("/markdown/export", s.wrapHandler(s.BackupHandler.DeleteMarkdowns))
	backupRouter.GET("/markdown/export/:filename", s.BackupHandler.DownloadMarkdown)
}

// registerCategoryRoutes 注册分类路由
func (s *Server) registerCategoryRoutes(router fiber.Router) {
	categoryRouter := router.Group("/categories")
	categoryRouter.PUT("/batch", s.wrapHandler(s.CategoryHandler.UpdateCategoryBatch))
	categoryRouter.GET("/:categoryID", s.wrapHandler(s.CategoryHandler.GetCategoryByID))
	categoryRouter.GET("", s.wrapHandler(s.CategoryHandler.ListAllCategory))
	categoryRouter.GET("/tree_view", s.wrapHandler(s.CategoryHandler.ListAsTree))
	categoryRouter.POST("", s.wrapHandler(s.CategoryHandler.CreateCategory))
	categoryRouter.PUT("/:categoryID", s.wrapHandler(s.CategoryHandler.UpdateCategory))
	categoryRouter.DELETE("/:categoryID", s.wrapHandler(s.CategoryHandler.DeleteCategory))
}

// registerPostRoutes 注册文章路由
func (s *Server) registerPostRoutes(router fiber.Router) {
	postRouter := router.Group("/posts")
	postRouter.GET("", s.wrapHandler(s.PostHandler.ListPosts))
	postRouter.GET("/latest", s.wrapHandler(s.PostHandler.ListLatestPosts))
	postRouter.GET("/status/:status", s.wrapHandler(s.PostHandler.ListPostsByStatus))
	postRouter.GET("/:postID", s.wrapHandler(s.PostHandler.GetByPostID))
	postRouter.POST("", s.wrapHandler(s.PostHandler.CreatePost))
	postRouter.PUT("/:postID", s.wrapHandler(s.PostHandler.UpdatePost))
	postRouter.PUT("/:postID/status/:status", s.wrapHandler(s.PostHandler.UpdatePostStatus))
	postRouter.PUT("/status/:status", s.wrapHandler(s.PostHandler.UpdatePostStatusBatch))
	postRouter.PUT("/:postID/status/draft/content", s.wrapHandler(s.PostHandler.UpdatePostDraft))
	postRouter.DELETE("/:postID", s.wrapHandler(s.PostHandler.DeletePost))
	postRouter.DELETE("", s.wrapHandler(s.PostHandler.DeletePostBatch))
	postRouter.GET("/:postID/preview", s.PostHandler.PreviewPost)

	// 文章评论路由
	postCommentRouter := postRouter.Group("/comments")
	postCommentRouter.GET("", s.wrapHandler(s.PostCommentHandler.ListPostComment))
	postCommentRouter.GET("/latest", s.wrapHandler(s.PostCommentHandler.ListPostCommentLatest))
	postCommentRouter.GET("/:postID/tree_view", s.wrapHandler(s.PostCommentHandler.ListPostCommentAsTree))
	postCommentRouter.GET("/:postID/list_view", s.wrapHandler(s.PostCommentHandler.ListPostCommentWithParent))
	postCommentRouter.POST("", s.wrapHandler(s.PostCommentHandler.CreatePostComment))
	postCommentRouter.PUT("/:commentID", s.wrapHandler(s.PostCommentHandler.UpdatePostComment))
	postCommentRouter.PUT("/:commentID/status/:status", s.wrapHandler(s.PostCommentHandler.UpdatePostCommentStatus))
	postCommentRouter.PUT("/status/:status", s.wrapHandler(s.PostCommentHandler.UpdatePostCommentStatusBatch))
	postCommentRouter.DELETE("/:commentID", s.wrapHandler(s.PostCommentHandler.DeletePostComment))
	postCommentRouter.DELETE("", s.wrapHandler(s.PostCommentHandler.DeletePostCommentBatch))
}

// registerOptionRoutes 注册选项路由
func (s *Server) registerOptionRoutes(router fiber.Router) {
	optionRouter := router.Group("/options")
	optionRouter.GET("", s.wrapHandler(s.OptionHandler.ListAllOptions))
	optionRouter.GET("/map_view", s.wrapHandler(s.OptionHandler.ListAllOptionsAsMap))
	optionRouter.POST("/map_view/keys", s.wrapHandler(s.OptionHandler.ListAllOptionsAsMapWithKey))
	optionRouter.POST("/saving", s.wrapHandler(s.OptionHandler.SaveOption))
	optionRouter.POST("/map_view/saving", s.wrapHandler(s.OptionHandler.SaveOptionWithMap))
}

// registerLogRoutes 注册日志路由
func (s *Server) registerLogRoutes(router fiber.Router) {
	logRouter := router.Group("/logs")
	logRouter.GET("/latest", s.wrapHandler(s.LogHandler.PageLatestLog))
	logRouter.GET("", s.wrapHandler(s.LogHandler.PageLog))
	logRouter.GET("/clear", s.wrapHandler(s.LogHandler.ClearLog))
}

// registerStatisticRoutes 注册统计路由
func (s *Server) registerStatisticRoutes(router fiber.Router) {
	statisticRouter := router.Group("/statistics")
	statisticRouter.GET("", s.wrapHandler(s.StatisticHandler.Statistics))
	statisticRouter.GET("user", s.wrapHandler(s.StatisticHandler.StatisticsWithUser))
}

// registerSheetRoutes 注册页面路由
func (s *Server) registerSheetRoutes(router fiber.Router) {
	sheetRouter := router.Group("/sheets")
	sheetRouter.GET("/:sheetID", s.wrapHandler(s.SheetHandler.GetSheetByID))
	sheetRouter.GET("", s.wrapHandler(s.SheetHandler.ListSheet))
	sheetRouter.POST("", s.wrapHandler(s.SheetHandler.CreateSheet))
	sheetRouter.PUT("/:sheetID", s.wrapHandler(s.SheetHandler.UpdateSheet))
	sheetRouter.PUT("/:sheetID/:status", s.wrapHandler(s.SheetHandler.UpdateSheetStatus))
	sheetRouter.PUT("/:sheetID/status/draft/content", s.wrapHandler(s.SheetHandler.UpdateSheetDraft))
	sheetRouter.DELETE("/:sheetID", s.wrapHandler(s.SheetHandler.DeleteSheet))
	sheetRouter.GET("/preview/:sheetID", s.SheetHandler.PreviewSheet)
	sheetRouter.GET("/independent", s.wrapHandler(s.SheetHandler.IndependentSheets))

	// 页面评论路由
	sheetCommentRouter := sheetRouter.Group("/comments")
	sheetCommentRouter.GET("", s.wrapHandler(s.SheetCommentHandler.ListSheetComment))
	sheetCommentRouter.GET("/latest", s.wrapHandler(s.SheetCommentHandler.ListSheetCommentLatest))
	sheetCommentRouter.GET("/:sheetID/tree_view", s.wrapHandler(s.SheetCommentHandler.ListSheetCommentAsTree))
	sheetCommentRouter.GET("/:sheetID/list_view", s.wrapHandler(s.SheetCommentHandler.ListSheetCommentWithParent))
	sheetCommentRouter.POST("/", s.wrapHandler(s.SheetCommentHandler.CreateSheetComment))
	sheetCommentRouter.PUT("/:commentID/status/:status", s.wrapHandler(s.SheetCommentHandler.UpdateSheetCommentStatus))
	sheetCommentRouter.PUT("/status/:status", s.wrapHandler(s.SheetCommentHandler.UpdateSheetCommentStatusBatch))
	sheetCommentRouter.DELETE("/:commentID", s.wrapHandler(s.SheetCommentHandler.DeleteSheetComment))
	sheetCommentRouter.DELETE("", s.wrapHandler(s.SheetCommentHandler.DeleteSheetCommentBatch))
}

// registerJournalRoutes 注册日记路由
func (s *Server) registerJournalRoutes(router fiber.Router) {
	journalRouter := router.Group("/journals")
	journalRouter.GET("", s.wrapHandler(s.JournalHandler.ListJournal))
	journalRouter.GET("/latest", s.wrapHandler(s.JournalHandler.ListLatestJournal))
	journalRouter.POST("", s.wrapHandler(s.JournalHandler.CreateJournal))
	journalRouter.PUT("/:journalID", s.wrapHandler(s.JournalHandler.UpdateJournal))
	journalRouter.DELETE("/:journalID", s.wrapHandler(s.JournalHandler.DeleteJournal))

	// 日记评论路由
	journalCommentRouter := journalRouter.Group("/comments")
	journalCommentRouter.GET("", s.wrapHandler(s.JournalCommentHandler.ListJournalComment))
	journalCommentRouter.GET("/latest", s.wrapHandler(s.JournalCommentHandler.ListJournalCommentLatest))
	journalCommentRouter.GET("/:journalID/tree_view", s.wrapHandler(s.JournalCommentHandler.ListJournalCommentAsTree))
	journalCommentRouter.GET("/:journalID/list_view", s.wrapHandler(s.JournalCommentHandler.ListJournalCommentWithParent))
	journalCommentRouter.POST("/", s.wrapHandler(s.JournalCommentHandler.CreateJournalComment))
	journalCommentRouter.PUT("/:commentID/status/:status", s.wrapHandler(s.JournalCommentHandler.UpdateJournalCommentStatus))
	journalCommentRouter.PUT("/status/:status", s.wrapHandler(s.JournalCommentHandler.UpdateJournalStatusBatch))
	journalCommentRouter.PUT("/:commentID", s.wrapHandler(s.JournalCommentHandler.UpdateJournalComment))
	journalCommentRouter.DELETE("/:commentID", s.wrapHandler(s.JournalCommentHandler.DeleteJournalComment))
	journalCommentRouter.DELETE("", s.wrapHandler(s.JournalCommentHandler.DeleteJournalCommentBatch))
}

// registerLinkRoutes 注册链接路由
func (s *Server) registerLinkRoutes(router fiber.Router) {
	linkRouter := router.Group("/links")
	linkRouter.GET("", s.wrapHandler(s.LinkHandler.ListLinks))
	linkRouter.GET("/:id", s.wrapHandler(s.LinkHandler.GetLinkByID))
	linkRouter.POST("", s.wrapHandler(s.LinkHandler.CreateLink))
	linkRouter.PUT("/:id", s.wrapHandler(s.LinkHandler.UpdateLink))
	linkRouter.DELETE("/:id", s.wrapHandler(s.LinkHandler.DeleteLink))
	linkRouter.GET("/teams", s.wrapHandler(s.LinkHandler.ListLinkTeams))
}

// registerMenuRoutes 注册菜单路由
func (s *Server) registerMenuRoutes(router fiber.Router) {
	menuRouter := router.Group("/menus")
	menuRouter.GET("", s.wrapHandler(s.MenuHandler.ListMenus))
	menuRouter.GET("/tree_view", s.wrapHandler(s.MenuHandler.ListMenusAsTree))
	menuRouter.GET("/team/tree_view", s.wrapHandler(s.MenuHandler.ListMenusAsTreeByTeam))
	menuRouter.GET("/:id", s.wrapHandler(s.MenuHandler.GetMenuByID))
	menuRouter.POST("", s.wrapHandler(s.MenuHandler.CreateMenu))
	menuRouter.POST("/batch", s.wrapHandler(s.MenuHandler.CreateMenuBatch))
	menuRouter.PUT("/:id", s.wrapHandler(s.MenuHandler.UpdateMenu))
	menuRouter.PUT("/batch", s.wrapHandler(s.MenuHandler.UpdateMenuBatch))
	menuRouter.DELETE("/:id", s.wrapHandler(s.MenuHandler.DeleteMenu))
	menuRouter.DELETE("/batch", s.wrapHandler(s.MenuHandler.DeleteMenuBatch))
	menuRouter.GET("/teams", s.wrapHandler(s.MenuHandler.ListMenuTeams))
}

// registerTagRoutes 注册标签路由
func (s *Server) registerTagRoutes(router fiber.Router) {
	tagRouter := router.Group("/tags")
	tagRouter.GET("", s.wrapHandler(s.TagHandler.ListTags))
	tagRouter.GET("/:id", s.wrapHandler(s.TagHandler.GetTagByID))
	tagRouter.POST("", s.wrapHandler(s.TagHandler.CreateTag))
	tagRouter.PUT("/:id", s.wrapHandler(s.TagHandler.UpdateTag))
	tagRouter.DELETE("/:id", s.wrapHandler(s.TagHandler.DeleteTag))
}

// registerPhotoRoutes 注册照片路由
func (s *Server) registerPhotoRoutes(router fiber.Router) {
	photoRouter := router.Group("/photos")
	photoRouter.GET("/latest", s.wrapHandler(s.PhotoHandler.ListPhoto))
	photoRouter.GET("", s.wrapHandler(s.PhotoHandler.PagePhotos))
	photoRouter.GET("/:id", s.wrapHandler(s.PhotoHandler.GetPhotoByID))
	photoRouter.DELETE("/batch", s.wrapHandler(s.PhotoHandler.DeletePhotoBatch))
	photoRouter.POST("", s.wrapHandler(s.PhotoHandler.CreatePhoto))
	photoRouter.POST("/batch", s.wrapHandler(s.PhotoHandler.CreatePhotoBatch))
	photoRouter.PUT("/:id", s.wrapHandler(s.PhotoHandler.UpdatePhoto))
	photoRouter.GET("/teams", s.wrapHandler(s.PhotoHandler.ListPhotoTeams))
}

// registerUserRoutes 注册用户路由
func (s *Server) registerUserRoutes(router fiber.Router) {
	userRouter := router.Group("/users")
	userRouter.GET("/profiles", s.wrapHandler(s.UserHandler.GetCurrentUserProfile))
	userRouter.PUT("/profiles", s.wrapHandler(s.UserHandler.UpdateUserProfile))
	userRouter.PUT("/profiles/password", s.wrapHandler(s.UserHandler.UpdatePassword))
	userRouter.PUT("/mfa/generate", s.wrapHandler(s.UserHandler.GenerateMFAQRCode))
	userRouter.PUT("/mfa/update", s.wrapHandler(s.UserHandler.UpdateMFA))
}

// registerThemeRoutes 注册主题路由
func (s *Server) registerThemeRoutes(router fiber.Router) {
	themeRouter := router.Group("themes")
	themeRouter.GET("/activation", s.wrapHandler(s.ThemeHandler.GetActivatedTheme))
	themeRouter.GET("/:themeID", s.wrapHandler(s.ThemeHandler.GetThemeByID))
	themeRouter.GET("", s.wrapHandler(s.ThemeHandler.ListAllThemes))
	themeRouter.GET("/activation/files", s.wrapHandler(s.ThemeHandler.ListActivatedThemeFile))
	themeRouter.GET("/:themeID/files", s.wrapHandler(s.ThemeHandler.ListThemeFileByID))
	themeRouter.GET("files/content", s.wrapHandler(s.ThemeHandler.GetThemeFileContent))
	themeRouter.GET("/:themeID/files/content", s.wrapHandler(s.ThemeHandler.GetThemeFileContentByID))
	themeRouter.PUT("/files/content", s.wrapHandler(s.ThemeHandler.UpdateThemeFile))
	themeRouter.PUT("/:themeID/files/content", s.wrapHandler(s.ThemeHandler.UpdateThemeFileByID))
	themeRouter.GET("activation/template/custom/sheet", s.wrapHandler(s.ThemeHandler.ListCustomSheetTemplate))
	themeRouter.GET("activation/template/custom/post", s.wrapHandler(s.ThemeHandler.ListCustomPostTemplate))
	themeRouter.POST("/:themeID/activation", s.wrapHandler(s.ThemeHandler.ActivateTheme))
	themeRouter.GET("activation/configurations", s.wrapHandler(s.ThemeHandler.GetActivatedThemeConfig))
	themeRouter.GET("/:themeID/configurations", s.wrapHandler(s.ThemeHandler.GetThemeConfigByID))
	themeRouter.GET("/:themeID/configurations/groups/:group", s.wrapHandler(s.ThemeHandler.GetThemeConfigByGroup))
	themeRouter.GET("/:themeID/configurations/groups", s.wrapHandler(s.ThemeHandler.GetThemeConfigGroupNames))
	themeRouter.GET("activation/settings", s.wrapHandler(s.ThemeHandler.GetActivatedThemeSettingMap))
	themeRouter.GET("/:themeID/settings", s.wrapHandler(s.ThemeHandler.GetThemeSettingMapByID))
	themeRouter.GET("/:themeID/groups/:group/settings", s.wrapHandler(s.ThemeHandler.GetThemeSettingMapByGroupAndThemeID))
	themeRouter.POST("activation/settings", s.wrapHandler(s.ThemeHandler.SaveActivatedThemeSetting))
	themeRouter.POST("/:themeID/settings", s.wrapHandler(s.ThemeHandler.SaveThemeSettingByID))
	themeRouter.DELETE("/:themeID", s.wrapHandler(s.ThemeHandler.DeleteThemeByID))
	themeRouter.POST("upload", s.wrapHandler(s.ThemeHandler.UploadTheme))
	themeRouter.PUT("upload/:themeID", s.wrapHandler(s.ThemeHandler.UpdateThemeByUpload))
	themeRouter.POST("fetching", s.wrapHandler(s.ThemeHandler.FetchTheme))
	themeRouter.PUT("fetching/:themeID", s.wrapHandler(s.ThemeHandler.UpdateThemeByFetching))
	themeRouter.POST("reload", s.wrapHandler(s.ThemeHandler.ReloadTheme))
	themeRouter.GET("activation/template/exists", s.wrapHandler(s.ThemeHandler.TemplateExist))
}

// registerEmailRoutes 注册邮件路由
func (s *Server) registerEmailRoutes(router fiber.Router) {
	emailRouter := router.Group("/mails")
	emailRouter.POST("/test", s.wrapHandler(s.EmailHandler.Test))
}

// registerContentRoutes 注册内容路由
func (s *Server) registerContentRoutes(router fiber.Router) {
	contentRouter := router.Group("")
	contentRouter.Use(
		s.LogMiddleware.LoggerWithConfig(middleware.FiberLoggerConfig{}),
		s.RecoveryMiddleware.RecoveryWithLogger(),
		s.InstallRedirectMiddleware.InstallRedirect(),
	)

	contentRouter.POST("/content/:type/:slug/authentication", s.wrapHTMLHandler(s.ViewHandler.Authenticate))

	contentRouter.GET("", s.wrapHTMLHandler(s.IndexHandler.Index))
	contentRouter.GET("/page/:page", s.wrapHTMLHandler(s.IndexHandler.IndexPage))
	contentRouter.GET("/robots.txt", s.wrapTextHandler(s.FeedHandler.Robots))
	contentRouter.GET("/atom", s.wrapTextHandler(s.FeedHandler.Atom))
	contentRouter.GET("/atom.xml", s.wrapTextHandler(s.FeedHandler.Atom))
	contentRouter.GET("/rss", s.wrapTextHandler(s.FeedHandler.Feed))
	contentRouter.GET("/rss.xml", s.wrapTextHandler(s.FeedHandler.Feed))
	contentRouter.GET("/feed", s.wrapTextHandler(s.FeedHandler.Feed))
	contentRouter.GET("/feed.xml", s.wrapTextHandler(s.FeedHandler.Feed))
	contentRouter.GET("/feed/categories/:slug", s.wrapTextHandler(s.FeedHandler.CategoryFeed))
	contentRouter.GET("/atom/categories/:slug", s.wrapTextHandler(s.FeedHandler.CategoryAtom))
	contentRouter.GET("/sitemap.xml", s.wrapTextHandler(s.FeedHandler.SitemapXML))
	contentRouter.GET("/sitemap.html", s.wrapHTMLHandler(s.FeedHandler.SitemapHTML))

	contentRouter.GET("/version", s.wrapHandler(s.ViewHandler.Version))
	contentRouter.GET("/install", s.ViewHandler.Install)
	contentRouter.GET("/logo", s.wrapHandler(s.ViewHandler.Logo))
	contentRouter.GET("/favicon", s.wrapHandler(s.ViewHandler.Favicon))
	contentRouter.GET("/search", s.wrapHTMLHandler(s.ContentSearchHandler.Search))
	contentRouter.GET("/search/page/:page", s.wrapHTMLHandler(s.ContentSearchHandler.PageSearch))

	// 注册动态路由
	err := s.registerDynamicRouters(contentRouter)
	if err != nil {
		s.logger.DPanic("regiterDynamicRouters err", zap.Error(err))
	}
}

// registerContentAPIRoutes 注册内容API路由
func (s *Server) registerContentAPIRoutes(router fiber.Router) {
	contentAPIRouter := router.Group("/api/content")
	contentAPIRouter.Use(
		s.LogMiddleware.LoggerWithConfig(middleware.FiberLoggerConfig{}),
		s.RecoveryMiddleware.RecoveryWithLogger(),
	)

	// 归档API
	contentAPIRouter.GET("/archives/years", s.wrapHandler(s.ContentAPIArchiveHandler.ListYearArchives))
	contentAPIRouter.GET("/archives/months", s.wrapHandler(s.ContentAPIArchiveHandler.ListMonthArchives))

	// 分类API
	contentAPIRouter.GET("/categories", s.wrapHandler(s.ContentAPICategoryHandler.ListCategories))
	contentAPIRouter.GET("/categories/:slug/posts", s.wrapHandler(s.ContentAPICategoryHandler.ListPosts))

	// 日记API
	contentAPIRouter.GET("/journals", s.wrapHandler(s.ContentAPIJournalHandler.ListJournal))
	contentAPIRouter.GET("/journals/:journalID", s.wrapHandler(s.ContentAPIJournalHandler.GetJournal))
	contentAPIRouter.GET("/journals/:journalID/comments/top_view", s.wrapHandler(s.ContentAPIJournalHandler.ListTopComment))
	contentAPIRouter.GET("/journals/:journalID/comments/:parentID/children", s.wrapHandler(s.ContentAPIJournalHandler.ListChildren))
	contentAPIRouter.GET("/journals/:journalID/comments/tree_view", s.wrapHandler(s.ContentAPIJournalHandler.ListCommentTree))
	contentAPIRouter.GET("/journals/:journalID/comments/list_view", s.wrapHandler(s.ContentAPIJournalHandler.ListComment))
	contentAPIRouter.POST("/journals/comments", s.wrapHandler(s.ContentAPIJournalHandler.CreateComment))
	contentAPIRouter.POST("/journals/:journalID/likes", s.wrapHandler(s.ContentAPIJournalHandler.Like))

	// 照片API
	contentAPIRouter.POST("/photos/:photoID/likes", s.wrapHandler(s.ContentAPIPhotoHandler.Like))

	// 文章API
	contentAPIRouter.GET("/posts/:postID/comments/top_view", s.wrapHandler(s.ContentAPIPostHandler.ListTopComment))
	contentAPIRouter.GET("/posts/:postID/comments/:parentID/children", s.wrapHandler(s.ContentAPIPostHandler.ListChildren))
	contentAPIRouter.GET("/posts/:postID/comments/tree_view", s.wrapHandler(s.ContentAPIPostHandler.ListCommentTree))
	contentAPIRouter.GET("/posts/:postID/comments/list_view", s.wrapHandler(s.ContentAPIPostHandler.ListComment))
	contentAPIRouter.POST("/posts/comments", s.wrapHandler(s.ContentAPIPostHandler.CreateComment))
	contentAPIRouter.POST("/posts/:postID/likes", s.wrapHandler(s.ContentAPIPostHandler.Like))

	// 页面API
	contentAPIRouter.GET("/sheets/:sheetID/comments/top_view", s.wrapHandler(s.ContentAPISheetHandler.ListTopComment))
	contentAPIRouter.GET("/sheets/:sheetID/comments/:parentID/children", s.wrapHandler(s.ContentAPISheetHandler.ListChildren))
	contentAPIRouter.GET("/sheets/:sheetID/comments/tree_view", s.wrapHandler(s.ContentAPISheetHandler.ListCommentTree))
	contentAPIRouter.GET("/sheets/:sheetID/comments/list_view", s.wrapHandler(s.ContentAPISheetHandler.ListComment))
	contentAPIRouter.POST("/sheets/comments", s.wrapHandler(s.ContentAPISheetHandler.CreateComment))

	// 链接API
	contentAPIRouter.GET("/links", s.wrapHandler(s.ContentAPILinkHandler.ListLinks))
	contentAPIRouter.GET("/links/team_view", s.wrapHandler(s.ContentAPILinkHandler.LinkTeamVO))

	// 选项API
	contentAPIRouter.GET("/options/comment", s.wrapHandler(s.ContentAPIOptionHandler.Comment))

	// 评论API
	contentAPIRouter.POST("/comments/:commentID/likes", s.wrapHandler(s.ContentAPICommentHandler.Like))
}

// registerDynamicRouters 注册动态路由
func (s *Server) registerDynamicRouters(contentRouter fiber.Router) error {
	ctx := context.Background()
	ctx = dal.SetCtxQuery(ctx, dal.GetQueryByCtx(ctx).ReplaceDB(dal.GetDB().Session(
		&gorm.Session{Logger: dal.DB.Logger.LogMode(logger.Warn)},
	)))

	archivePath, err := s.OptionService.GetArchivePrefix(ctx)
	if err != nil {
		return err
	}
	categoryPath, err := s.OptionService.GetCategoryPrefix(ctx)
	if err != nil {
		return err
	}
	sheetPermaLinkType, err := s.OptionService.GetSheetPermalinkType(ctx)
	if err != nil {
		return err
	}
	sheetPath, err := s.OptionService.GetSheetPrefix(ctx)
	if err != nil {
		return err
	}
	tagPath, err := s.OptionService.GetTagPrefix(ctx)
	if err != nil {
		return err
	}
	journalPath, err := s.OptionService.GetJournalPrefix(ctx)
	if err != nil {
		return err
	}

	photoPath, err := s.OptionService.GetPhotoPrefix(ctx)
	if err != nil {
		return err
	}
	linkPath, err := s.OptionService.GetLinkPrefix(ctx)
	if err != nil {
		return err
	}
	contentRouter.GET(archivePath, s.wrapHTMLHandler(s.ArchiveHandler.Archives))
	contentRouter.GET(archivePath+"/page/:page", s.wrapHTMLHandler(s.ArchiveHandler.ArchivesPage))
	contentRouter.GET(archivePath+"/:slug", s.wrapHTMLHandler(s.ArchiveHandler.ArchivesBySlug))

	contentRouter.GET(tagPath, s.wrapHTMLHandler(s.ContentTagHandler.Tags))
	contentRouter.GET(tagPath+"/:slug/page/:page", s.wrapHTMLHandler(s.ContentTagHandler.TagPostPage))
	contentRouter.GET(tagPath+"/:slug", s.wrapHTMLHandler(s.ContentTagHandler.TagPost))

	contentRouter.GET(categoryPath, s.wrapHTMLHandler(s.ContentCategoryHandler.Categories))
	contentRouter.GET(categoryPath+"/:slug", s.wrapHTMLHandler(s.ContentCategoryHandler.CategoryDetail))
	contentRouter.GET(categoryPath+"/:slug/page/:page", s.wrapHTMLHandler(s.ContentCategoryHandler.CategoryDetailPage))

	contentRouter.GET(linkPath, s.wrapHTMLHandler(s.ContentLinkHandler.Link))

	contentRouter.GET(photoPath, s.wrapHTMLHandler(s.ContentPhotoHandler.Phtotos))
	contentRouter.GET(photoPath+"/page/:page", s.wrapHTMLHandler(s.ContentPhotoHandler.PhotosPage))

	contentRouter.GET(journalPath, s.wrapHTMLHandler(s.ContentJournalHandler.Journals))
	contentRouter.GET(journalPath+"/page/:page", s.wrapHTMLHandler(s.ContentJournalHandler.JournalsPage))
	contentRouter.GET("admin_preview/"+archivePath+"/:slug", s.wrapHTMLHandler(s.ArchiveHandler.AdminArchivesBySlug))
	if sheetPermaLinkType == consts.SheetPermaLinkTypeRoot {
		contentRouter.GET("/:slug")
	} else {
		contentRouter.GET(sheetPath+"/:slug", s.wrapHTMLHandler(s.ContentSheetHandler.SheetBySlug))
	}
	contentRouter.GET("admin_preview/"+sheetPath+"/:slug", s.wrapHTMLHandler(s.ContentSheetHandler.AdminSheetBySlug))
	return nil
}
