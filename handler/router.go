package handler

import (
	"context"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/go-sonic/sonic/config"
	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/dal"
	"github.com/go-sonic/sonic/handler/middleware"
)

func (s *Server) RegisterRouters() {
	router := s.Router
	if config.IsDev() {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     "*",
			AllowMethods:     "PUT, PATCH, GET, DELETE, POST, OPTIONS",
			AllowHeaders:     "Origin, Admin-Authorization, Content-Type",
			AllowCredentials: true,
			ExposeHeaders:    "Content-Length",
		}))
	}

	{
		router.Get("/ping", func(ctx *fiber.Ctx) error {
			return ctx.SendString("pong")
		})
		{
			staticRouter := router.Group("/")
			staticRouter.Static(s.Config.Sonic.AdminURLPath, s.Config.Sonic.AdminResourcesDir)
			staticRouter.Static("/css", filepath.Join(s.Config.Sonic.AdminResourcesDir, "css"))
			staticRouter.Static("/js", filepath.Join(s.Config.Sonic.AdminResourcesDir, "js"))
			staticRouter.Static("/images", filepath.Join(s.Config.Sonic.AdminResourcesDir, "images"))
			
			// Upload dir with CacheControl
			// In Fiber, we can apply middleware to the path before Static
			router.Use(consts.SonicUploadDir, middleware.NewCacheControlMiddleware(middleware.WithMaxAge(time.Hour*24*7)).CacheControl())
			router.Static(consts.SonicUploadDir, s.Config.Sonic.UploadDir)
			
			staticRouter.Static("/themes/", s.Config.Sonic.ThemeDir)
		}
		{
			adminAPIRouter := router.Group("/api/admin")
			adminAPIRouter.Use(s.LogMiddleware.LoggerWithConfig(middleware.GinLoggerConfig{}), s.RecoveryMiddleware.RecoveryWithLogger(), s.InstallRedirectMiddleware.InstallRedirect())
			adminAPIRouter.Get("/is_installed", s.wrapHandler(s.AdminHandler.IsInstalled))
			adminAPIRouter.Post("/login/precheck", s.wrapHandler(s.AdminHandler.AuthPreCheck))
			adminAPIRouter.Post("/login", s.wrapHandler(s.AdminHandler.Auth))
			adminAPIRouter.Post("/refresh/:refreshToken", s.wrapHandler(s.AdminHandler.RefreshToken))
			adminAPIRouter.Post("/installations", s.wrapHandler(s.InstallHandler.InstallBlog))
			{
				authRouter := adminAPIRouter.Group("")
				authRouter.Use(s.AuthMiddleware.GetWrapHandler())
				authRouter.Post("/logout", s.wrapHandler(s.AdminHandler.LogOut))
				authRouter.Post("/password/code", s.wrapHandler(s.AdminHandler.SendResetCode))
				authRouter.Get("/environments", s.wrapHandler(s.AdminHandler.GetEnvironments))
				authRouter.Get("/sonic/logfile", s.wrapHandler(s.AdminHandler.GetLogFiles))
				{
					attachmentRouter := authRouter.Group("/attachments")
					attachmentRouter.Post("/upload", s.wrapHandler(s.AttachmentHandler.UploadAttachment))
					attachmentRouter.Post("/uploads", s.wrapHandler(s.AttachmentHandler.UploadAttachments))
					attachmentRouter.Get("/media_types", s.wrapHandler(s.AttachmentHandler.GetAllMediaType))
					attachmentRouter.Get("types", s.wrapHandler(s.AttachmentHandler.GetAllTypes))
					attachmentRouter.Delete("/:id", s.wrapHandler(s.AttachmentHandler.DeleteAttachment))
					attachmentRouter.Delete("", s.wrapHandler(s.AttachmentHandler.DeleteAttachmentInBatch))
					attachmentRouter.Get("", s.wrapHandler(s.AttachmentHandler.QueryAttachment))
					attachmentRouter.Get("/:id", s.wrapHandler(s.AttachmentHandler.GetAttachmentByID))
					attachmentRouter.Put("/:id", s.wrapHandler(s.AttachmentHandler.UpdateAttachment))
				}
				{
					backupRouter := authRouter.Group("/backups")
					backupRouter.Post("/work-dir", s.wrapHandler(s.BackupHandler.BackupWholeSite))
					backupRouter.Get("/work-dir", s.wrapHandler(s.BackupHandler.ListBackups))
					backupRouter.Get("/work-dir/*", s.BackupHandler.HandleWorkDir)
					backupRouter.Delete("/work-dir", s.wrapHandler(s.BackupHandler.DeleteBackups))
					backupRouter.Post("/data", s.wrapHandler(s.BackupHandler.ExportData))
					backupRouter.Delete("/data", s.wrapHandler(s.BackupHandler.DeleteDataFile))
					backupRouter.Get("/data/*", s.BackupHandler.HandleData)
					backupRouter.Post("/markdown/export", s.wrapHandler(s.BackupHandler.ExportMarkdown))
					backupRouter.Post("/markdown/import", s.wrapHandler(s.BackupHandler.ImportMarkdown))
					backupRouter.Get("/markdown/fetch", s.wrapHandler(s.BackupHandler.GetMarkDownBackup))
					backupRouter.Get("/markdown/export", s.wrapHandler(s.BackupHandler.ListMarkdowns))
					backupRouter.Delete("/markdown/export", s.wrapHandler(s.BackupHandler.DeleteMarkdowns))
					backupRouter.Get("/markdown/export/:filename", s.BackupHandler.DownloadMarkdown)
				}
				{
					categoryRouter := authRouter.Group("/categories")
					categoryRouter.Put("/batch", s.wrapHandler(s.CategoryHandler.UpdateCategoryBatch))
					categoryRouter.Get("/:categoryID", s.wrapHandler(s.CategoryHandler.GetCategoryByID))
					categoryRouter.Get("", s.wrapHandler(s.CategoryHandler.ListAllCategory))
					categoryRouter.Get("/tree_view", s.wrapHandler(s.CategoryHandler.ListAsTree))
					categoryRouter.Post("", s.wrapHandler(s.CategoryHandler.CreateCategory))
					categoryRouter.Put("/:categoryID", s.wrapHandler(s.CategoryHandler.UpdateCategory))
					categoryRouter.Delete("/:categoryID", s.wrapHandler(s.CategoryHandler.DeleteCategory))
				}
				{
					postRouter := authRouter.Group("/posts")
					// Move specific routes before parameterized routes
					{
						postCommentRouter := postRouter.Group("/comments")
						postCommentRouter.Get("", s.wrapHandler(s.PostCommentHandler.ListPostComment))
						postCommentRouter.Get("/latest", s.wrapHandler(s.PostCommentHandler.ListPostCommentLatest))
						postCommentRouter.Get("/:postID/tree_view", s.wrapHandler(s.PostCommentHandler.ListPostCommentAsTree))
						postCommentRouter.Get("/:postID/list_view", s.wrapHandler(s.PostCommentHandler.ListPostCommentWithParent))
						postCommentRouter.Post("", s.wrapHandler(s.PostCommentHandler.CreatePostComment))
						postCommentRouter.Put("/:commentID", s.wrapHandler(s.PostCommentHandler.UpdatePostComment))
						postCommentRouter.Put("/:commentID/status/:status", s.wrapHandler(s.PostCommentHandler.UpdatePostCommentStatus))
						postCommentRouter.Put("/status/:status", s.wrapHandler(s.PostCommentHandler.UpdatePostCommentStatusBatch))
						postCommentRouter.Delete("/:commentID", s.wrapHandler(s.PostCommentHandler.DeletePostComment))
						postCommentRouter.Delete("", s.wrapHandler(s.PostCommentHandler.DeletePostCommentBatch))
					}
					postRouter.Get("", s.wrapHandler(s.PostHandler.ListPosts))
					postRouter.Get("/latest", s.wrapHandler(s.PostHandler.ListLatestPosts))
					postRouter.Get("/status/:status", s.wrapHandler(s.PostHandler.ListPostsByStatus))
					postRouter.Get("/:postID", s.wrapHandler(s.PostHandler.GetByPostID))
					postRouter.Post("", s.wrapHandler(s.PostHandler.CreatePost))
					postRouter.Put("/:postID", s.wrapHandler(s.PostHandler.UpdatePost))
					postRouter.Put("/:postID/status/:status", s.wrapHandler(s.PostHandler.UpdatePostStatus))
					postRouter.Put("/status/:status", s.wrapHandler(s.PostHandler.UpdatePostStatusBatch))
					postRouter.Put("/:postID/status/draft/content", s.wrapHandler(s.PostHandler.UpdatePostDraft))
					postRouter.Delete("/:postID", s.wrapHandler(s.PostHandler.DeletePost))
					postRouter.Delete("", s.wrapHandler(s.PostHandler.DeletePostBatch))
					postRouter.Get("/:postID/preview", s.PostHandler.PreviewPost)
				}
				{
					optionRouter := authRouter.Group("/options")
					optionRouter.Get("", s.wrapHandler(s.OptionHandler.ListAllOptions))
					optionRouter.Get("/map_view", s.wrapHandler(s.OptionHandler.ListAllOptionsAsMap))
					optionRouter.Post("/map_view/keys", s.wrapHandler(s.OptionHandler.ListAllOptionsAsMapWithKey))
					optionRouter.Post("/saving", s.wrapHandler(s.OptionHandler.SaveOption))
					optionRouter.Post("/map_view/saving", s.wrapHandler(s.OptionHandler.SaveOptionWithMap))
				}
				{
					logRouter := authRouter.Group("/logs")
					logRouter.Get("/latest", s.wrapHandler(s.LogHandler.PageLatestLog))
					logRouter.Get("", s.wrapHandler(s.LogHandler.PageLog))
					logRouter.Get("/clear", s.wrapHandler(s.LogHandler.ClearLog))
				}
				{
					statisticRouter := authRouter.Group("/statistics")
					statisticRouter.Get("", s.wrapHandler(s.StatisticHandler.Statistics))
					statisticRouter.Get("user", s.wrapHandler(s.StatisticHandler.StatisticsWithUser))
				}
				{
					sheetRouter := authRouter.Group("/sheets")
					sheetRouter.Get("/independent", s.wrapHandler(s.SheetHandler.IndependentSheets))
					{
						sheetCommentRouter := sheetRouter.Group("/comments")
						sheetCommentRouter.Get("", s.wrapHandler(s.SheetCommentHandler.ListSheetComment))
						sheetCommentRouter.Get("/latest", s.wrapHandler(s.SheetCommentHandler.ListSheetCommentLatest))
						sheetCommentRouter.Get("/:sheetID/tree_view", s.wrapHandler(s.SheetCommentHandler.ListSheetCommentAsTree))
						sheetCommentRouter.Get("/:sheetID/list_view", s.wrapHandler(s.SheetCommentHandler.ListSheetCommentWithParent))
						sheetCommentRouter.Post("/", s.wrapHandler(s.SheetCommentHandler.CreateSheetComment))
						sheetCommentRouter.Put("/:commentID/status/:status", s.wrapHandler(s.SheetCommentHandler.UpdateSheetCommentStatus))
						sheetCommentRouter.Put("/status/:status", s.wrapHandler(s.SheetCommentHandler.UpdateSheetCommentStatusBatch))
						sheetCommentRouter.Delete("/:commentID", s.wrapHandler(s.SheetCommentHandler.DeleteSheetComment))
						sheetCommentRouter.Delete("", s.wrapHandler(s.SheetCommentHandler.DeleteSheetCommentBatch))
					}
					sheetRouter.Get("/:sheetID", s.wrapHandler(s.SheetHandler.GetSheetByID))
					sheetRouter.Get("", s.wrapHandler(s.SheetHandler.ListSheet))
					sheetRouter.Post("", s.wrapHandler(s.SheetHandler.CreateSheet))
					sheetRouter.Put("/:sheetID", s.wrapHandler(s.SheetHandler.UpdateSheet))
					sheetRouter.Put("/:sheetID/:status", s.wrapHandler(s.SheetHandler.UpdateSheetStatus))
					sheetRouter.Put("/:sheetID/status/draft/content", s.wrapHandler(s.SheetHandler.UpdateSheetDraft))
					sheetRouter.Delete("/:sheetID", s.wrapHandler(s.SheetHandler.DeleteSheet))
					sheetRouter.Get("/preview/:sheetID", s.SheetHandler.PreviewSheet)
				}
				{
					journalRouter := authRouter.Group("/journals")
					// Move specific routes before parameterized routes
					{
						journalCommentRouter := journalRouter.Group("/comments")
						journalCommentRouter.Get("", s.wrapHandler(s.JournalCommentHandler.ListJournalComment))
						journalCommentRouter.Get("/latest", s.wrapHandler(s.JournalCommentHandler.ListJournalCommentLatest))
						journalCommentRouter.Get("/:journalID/tree_view", s.wrapHandler(s.JournalCommentHandler.ListJournalCommentAsTree))
						journalCommentRouter.Get("/:journalID/list_view", s.wrapHandler(s.JournalCommentHandler.ListJournalCommentWithParent))
						journalCommentRouter.Post("/", s.wrapHandler(s.JournalCommentHandler.CreateJournalComment))
						journalCommentRouter.Put("/:commentID/status/:status", s.wrapHandler(s.JournalCommentHandler.UpdateJournalCommentStatus))
						journalCommentRouter.Put("/status/:status", s.wrapHandler(s.JournalCommentHandler.UpdateJournalStatusBatch))
						journalCommentRouter.Put("/:commentID", s.wrapHandler(s.JournalCommentHandler.UpdateJournalComment))
						journalCommentRouter.Delete("/:commentID", s.wrapHandler(s.JournalCommentHandler.DeleteJournalComment))
						journalCommentRouter.Delete("", s.wrapHandler(s.JournalCommentHandler.DeleteJournalCommentBatch))
					}
					journalRouter.Get("", s.wrapHandler(s.JournalHandler.ListJournal))
					journalRouter.Get("/latest", s.wrapHandler(s.JournalHandler.ListLatestJournal))
					journalRouter.Post("", s.wrapHandler(s.JournalHandler.CreateJournal))
					journalRouter.Put("/:journalID", s.wrapHandler(s.JournalHandler.UpdateJournal))
					journalRouter.Delete("/:journalID", s.wrapHandler(s.JournalHandler.DeleteJournal))
				}

				{
					linkRouter := authRouter.Group("/links")
					linkRouter.Get("", s.wrapHandler(s.LinkHandler.ListLinks))
					linkRouter.Get("/teams", s.wrapHandler(s.LinkHandler.ListLinkTeams))
					linkRouter.Get("/:id", s.wrapHandler(s.LinkHandler.GetLinkByID))
					linkRouter.Post("", s.wrapHandler(s.LinkHandler.CreateLink))
					linkRouter.Put("/:id", s.wrapHandler(s.LinkHandler.UpdateLink))
					linkRouter.Delete("/:id", s.wrapHandler(s.LinkHandler.DeleteLink))
				}
				{
					menuRouter := authRouter.Group("/menus")
					menuRouter.Get("", s.wrapHandler(s.MenuHandler.ListMenus))
					menuRouter.Get("/tree_view", s.wrapHandler(s.MenuHandler.ListMenusAsTree))
					menuRouter.Get("/team/tree_view", s.wrapHandler(s.MenuHandler.ListMenusAsTreeByTeam))
					menuRouter.Get("/teams", s.wrapHandler(s.MenuHandler.ListMenuTeams))
					menuRouter.Get("/:id", s.wrapHandler(s.MenuHandler.GetMenuByID))
					menuRouter.Post("", s.wrapHandler(s.MenuHandler.CreateMenu))
					menuRouter.Post("/batch", s.wrapHandler(s.MenuHandler.CreateMenuBatch))
					menuRouter.Put("/:id", s.wrapHandler(s.MenuHandler.UpdateMenu))
					menuRouter.Put("/batch", s.wrapHandler(s.MenuHandler.UpdateMenuBatch))
					menuRouter.Delete("/:id", s.wrapHandler(s.MenuHandler.DeleteMenu))
					menuRouter.Delete("/batch", s.wrapHandler(s.MenuHandler.DeleteMenuBatch))
				}
				{
					tagRouter := authRouter.Group("/tags")
					tagRouter.Get("", s.wrapHandler(s.TagHandler.ListTags))
					tagRouter.Get("/:id", s.wrapHandler(s.TagHandler.GetTagByID))
					tagRouter.Post("", s.wrapHandler(s.TagHandler.CreateTag))
					tagRouter.Put("/:id", s.wrapHandler(s.TagHandler.UpdateTag))
					tagRouter.Delete("/:id", s.wrapHandler(s.TagHandler.DeleteTag))
				}
				{
					photoRouter := authRouter.Group("/photos")
					photoRouter.Get("/latest", s.wrapHandler(s.PhotoHandler.ListPhoto))
					photoRouter.Get("", s.wrapHandler(s.PhotoHandler.PagePhotos))
					photoRouter.Get("/teams", s.wrapHandler(s.PhotoHandler.ListPhotoTeams))
					photoRouter.Get("/:id", s.wrapHandler(s.PhotoHandler.GetPhotoByID))
					photoRouter.Delete("/batch", s.wrapHandler(s.PhotoHandler.DeletePhotoBatch))
					photoRouter.Post("", s.wrapHandler(s.PhotoHandler.CreatePhoto))
					photoRouter.Post("/batch", s.wrapHandler(s.PhotoHandler.CreatePhotoBatch))
					photoRouter.Put("/:id", s.wrapHandler(s.PhotoHandler.UpdatePhoto))
				}
				{
					userRouter := authRouter.Group("/users")
					userRouter.Get("/profiles", s.wrapHandler(s.UserHandler.GetCurrentUserProfile))
					userRouter.Put("/profiles", s.wrapHandler(s.UserHandler.UpdateUserProfile))
					userRouter.Put("/profiles/password", s.wrapHandler(s.UserHandler.UpdatePassword))
					userRouter.Put("/mfa/generate", s.wrapHandler(s.UserHandler.GenerateMFAQRCode))
					userRouter.Put("/mfa/update", s.wrapHandler(s.UserHandler.UpdateMFA))
				}
				{
					themeRouter := authRouter.Group("themes")
					themeRouter.Get("/activation", s.wrapHandler(s.ThemeHandler.GetActivatedTheme))
					themeRouter.Get("/:themeID", s.wrapHandler(s.ThemeHandler.GetThemeByID))
					themeRouter.Get("", s.wrapHandler(s.ThemeHandler.ListAllThemes))
					themeRouter.Get("/activation/files", s.wrapHandler(s.ThemeHandler.ListActivatedThemeFile))
					themeRouter.Get("/:themeID/files", s.wrapHandler(s.ThemeHandler.ListThemeFileByID))
					themeRouter.Get("files/content", s.wrapHandler(s.ThemeHandler.GetThemeFileContent))
					themeRouter.Get("/:themeID/files/content", s.wrapHandler(s.ThemeHandler.GetThemeFileContentByID))
					themeRouter.Put("/files/content", s.wrapHandler(s.ThemeHandler.UpdateThemeFile))
					themeRouter.Put("/:themeID/files/content", s.wrapHandler(s.ThemeHandler.UpdateThemeFileByID))
					themeRouter.Get("activation/template/custom/sheet", s.wrapHandler(s.ThemeHandler.ListCustomSheetTemplate))
					themeRouter.Get("activation/template/custom/post", s.wrapHandler(s.ThemeHandler.ListCustomPostTemplate))
					themeRouter.Post("/:themeID/activation", s.wrapHandler(s.ThemeHandler.ActivateTheme))
					themeRouter.Get("activation/configurations", s.wrapHandler(s.ThemeHandler.GetActivatedThemeConfig))
					themeRouter.Get("/:themeID/configurations", s.wrapHandler(s.ThemeHandler.GetThemeConfigByID))
					themeRouter.Get("/:themeID/configurations/groups/:group", s.wrapHandler(s.ThemeHandler.GetThemeConfigByGroup))
					themeRouter.Get("/:themeID/configurations/groups", s.wrapHandler(s.ThemeHandler.GetThemeConfigGroupNames))
					themeRouter.Get("activation/settings", s.wrapHandler(s.ThemeHandler.GetActivatedThemeSettingMap))
					themeRouter.Get("/:themeID/settings", s.wrapHandler(s.ThemeHandler.GetThemeSettingMapByID))
					themeRouter.Get("/:themeID/groups/:group/settings", s.wrapHandler(s.ThemeHandler.GetThemeSettingMapByGroupAndThemeID))
					themeRouter.Post("activation/settings", s.wrapHandler(s.ThemeHandler.SaveActivatedThemeSetting))
					themeRouter.Post("/:themeID/settings", s.wrapHandler(s.ThemeHandler.SaveThemeSettingByID))
					themeRouter.Delete("/:themeID", s.wrapHandler(s.ThemeHandler.DeleteThemeByID))
					themeRouter.Post("upload", s.wrapHandler(s.ThemeHandler.UploadTheme))
					themeRouter.Put("upload/:themeID", s.wrapHandler(s.ThemeHandler.UpdateThemeByUpload))
					themeRouter.Post("fetching", s.wrapHandler(s.ThemeHandler.FetchTheme))
					themeRouter.Put("fetching/:themeID", s.wrapHandler(s.ThemeHandler.UpdateThemeByFetching))
					themeRouter.Post("reload", s.wrapHandler(s.ThemeHandler.ReloadTheme))
					themeRouter.Get("activation/template/exists", s.wrapHandler(s.ThemeHandler.TemplateExist))
				}
				{
					emailRouter := authRouter.Group("/mails")
					emailRouter.Post("/test", s.wrapHandler(s.EmailHandler.Test))
				}
			}
		}
		{
			contentRouter := router.Group("")
			contentRouter.Use(s.LogMiddleware.LoggerWithConfig(middleware.GinLoggerConfig{}), s.RecoveryMiddleware.RecoveryWithLogger(), s.InstallRedirectMiddleware.InstallRedirect())

			contentRouter.Post("/content/:type/:slug/authentication", s.wrapHTMLHandler(s.ViewHandler.Authenticate))

			contentRouter.Get("", s.wrapHTMLHandler(s.IndexHandler.Index))
			contentRouter.Get("/page/:page", s.wrapHTMLHandler(s.IndexHandler.IndexPage))
			contentRouter.Get("/robots.txt", s.wrapTextHandler(s.FeedHandler.Robots))
			contentRouter.Get("/atom", s.wrapTextHandler(s.FeedHandler.Atom))
			contentRouter.Get("/atom.xml", s.wrapTextHandler(s.FeedHandler.Atom))
			contentRouter.Get("/rss", s.wrapTextHandler(s.FeedHandler.Feed))
			contentRouter.Get("/rss.xml", s.wrapTextHandler(s.FeedHandler.Feed))
			contentRouter.Get("/feed", s.wrapTextHandler(s.FeedHandler.Feed))
			contentRouter.Get("/feed.xml", s.wrapTextHandler(s.FeedHandler.Feed))
			contentRouter.Get("/feed/categories/:slug", s.wrapTextHandler(s.FeedHandler.CategoryFeed))
			contentRouter.Get("/atom/categories/:slug", s.wrapTextHandler(s.FeedHandler.CategoryAtom))
			contentRouter.Get("/sitemap.xml", s.wrapTextHandler(s.FeedHandler.SitemapXML))
			contentRouter.Get("/sitemap.html", s.wrapHTMLHandler(s.FeedHandler.SitemapHTML))

			contentRouter.Get("/version", s.wrapHandler(s.ViewHandler.Version))
			contentRouter.Get("/install", s.ViewHandler.Install)
			contentRouter.Get("/logo", s.wrapHandler(s.ViewHandler.Logo))
			contentRouter.Get("/favicon", s.wrapHandler(s.ViewHandler.Favicon))
			contentRouter.Get("/search", s.wrapHTMLHandler(s.ContentSearchHandler.Search))
			contentRouter.Get("/search/page/:page", s.wrapHTMLHandler(s.ContentSearchHandler.PageSearch))
			err := s.registerDynamicRouters(contentRouter)
			if err != nil {
				s.logger.DPanic("regiterDynamicRouters err", zap.Error(err))
			}
		}
		{
			contentAPIRouter := router.Group("/api/content")
			contentAPIRouter.Use(s.LogMiddleware.LoggerWithConfig(middleware.GinLoggerConfig{}), s.RecoveryMiddleware.RecoveryWithLogger())

			contentAPIRouter.Get("/archives/years", s.wrapHandler(s.ContentAPIArchiveHandler.ListYearArchives))
			contentAPIRouter.Get("/archives/months", s.wrapHandler(s.ContentAPIArchiveHandler.ListMonthArchives))

			contentAPIRouter.Get("/categories", s.wrapHandler(s.ContentAPICategoryHandler.ListCategories))
			contentAPIRouter.Get("/categories/:slug/posts", s.wrapHandler(s.ContentAPICategoryHandler.ListPosts))

			contentAPIRouter.Get("/journals", s.wrapHandler(s.ContentAPIJournalHandler.ListJournal))
			contentAPIRouter.Get("/journals/comments", s.wrapHandler(s.ContentAPIJournalHandler.ListComment)) // Moved up if exists or needed
			contentAPIRouter.Post("/journals/comments", s.wrapHandler(s.ContentAPIJournalHandler.CreateComment)) // Create is specific
			contentAPIRouter.Get("/journals/:journalID", s.wrapHandler(s.ContentAPIJournalHandler.GetJournal))
			contentAPIRouter.Get("/journals/:journalID/comments/top_view", s.wrapHandler(s.ContentAPIJournalHandler.ListTopComment))
			contentAPIRouter.Get("/journals/:journalID/comments/:parentID/children", s.wrapHandler(s.ContentAPIJournalHandler.ListChildren))
			contentAPIRouter.Get("/journals/:journalID/comments/tree_view", s.wrapHandler(s.ContentAPIJournalHandler.ListCommentTree))
			contentAPIRouter.Get("/journals/:journalID/comments/list_view", s.wrapHandler(s.ContentAPIJournalHandler.ListComment))
			contentAPIRouter.Post("/journals/:journalID/likes", s.wrapHandler(s.ContentAPIJournalHandler.Like))

			contentAPIRouter.Post("/photos/:photoID/likes", s.wrapHandler(s.ContentAPIPhotoHandler.Like))

			contentAPIRouter.Post("/posts/comments", s.wrapHandler(s.ContentAPIPostHandler.CreateComment)) // Create comment is specific
			contentAPIRouter.Get("/posts/:postID/comments/top_view", s.wrapHandler(s.ContentAPIPostHandler.ListTopComment))
			contentAPIRouter.Get("/posts/:postID/comments/:parentID/children", s.wrapHandler(s.ContentAPIPostHandler.ListChildren))
			contentAPIRouter.Get("/posts/:postID/comments/tree_view", s.wrapHandler(s.ContentAPIPostHandler.ListCommentTree))
			contentAPIRouter.Get("/posts/:postID/comments/list_view", s.wrapHandler(s.ContentAPIPostHandler.ListComment))
			contentAPIRouter.Post("/posts/:postID/likes", s.wrapHandler(s.ContentAPIPostHandler.Like))

			contentAPIRouter.Post("/sheets/comments", s.wrapHandler(s.ContentAPISheetHandler.CreateComment)) // Create comment is specific
			contentAPIRouter.Get("/sheets/:sheetID/comments/top_view", s.wrapHandler(s.ContentAPISheetHandler.ListTopComment))
			contentAPIRouter.Get("/sheets/:sheetID/comments/:parentID/children", s.wrapHandler(s.ContentAPISheetHandler.ListChildren))
			contentAPIRouter.Get("/sheets/:sheetID/comments/tree_view", s.wrapHandler(s.ContentAPISheetHandler.ListCommentTree))
			contentAPIRouter.Get("/sheets/:sheetID/comments/list_view", s.wrapHandler(s.ContentAPISheetHandler.ListComment))

			contentAPIRouter.Get("/links", s.wrapHandler(s.ContentAPILinkHandler.ListLinks))
			contentAPIRouter.Get("/links/team_view", s.wrapHandler(s.ContentAPILinkHandler.LinkTeamVO))

			contentAPIRouter.Get("/options/comment", s.wrapHandler(s.ContentAPIOptionHandler.Comment))

			contentAPIRouter.Post("/comments/:commentID/likes", s.wrapHandler(s.ContentAPICommentHandler.Like))
		}
	}
}

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
	contentRouter.Get(archivePath, s.wrapHTMLHandler(s.ArchiveHandler.Archives))
	contentRouter.Get(archivePath+"/page/:page", s.wrapHTMLHandler(s.ArchiveHandler.ArchivesPage))
	contentRouter.Get(archivePath+"/:slug", s.wrapHTMLHandler(s.ArchiveHandler.ArchivesBySlug))

	contentRouter.Get(tagPath, s.wrapHTMLHandler(s.ContentTagHandler.Tags))
	contentRouter.Get(tagPath+"/:slug/page/:page", s.wrapHTMLHandler(s.ContentTagHandler.TagPostPage))
	contentRouter.Get(tagPath+"/:slug", s.wrapHTMLHandler(s.ContentTagHandler.TagPost))

	contentRouter.Get(categoryPath, s.wrapHTMLHandler(s.ContentCategoryHandler.Categories))
	contentRouter.Get(categoryPath+"/:slug", s.wrapHTMLHandler(s.ContentCategoryHandler.CategoryDetail))
	contentRouter.Get(categoryPath+"/:slug/page/:page", s.wrapHTMLHandler(s.ContentCategoryHandler.CategoryDetailPage))

	contentRouter.Get(linkPath, s.wrapHTMLHandler(s.ContentLinkHandler.Link))

	contentRouter.Get(photoPath, s.wrapHTMLHandler(s.ContentPhotoHandler.Phtotos))
	contentRouter.Get(photoPath+"/page/:page", s.wrapHTMLHandler(s.ContentPhotoHandler.PhotosPage))

	contentRouter.Get(journalPath, s.wrapHTMLHandler(s.ContentJournalHandler.Journals))
	contentRouter.Get(journalPath+"/page/:page", s.wrapHTMLHandler(s.ContentJournalHandler.JournalsPage))
	contentRouter.Get("admin_preview/"+archivePath+"/:slug", s.wrapHTMLHandler(s.ArchiveHandler.AdminArchivesBySlug))
	if sheetPermaLinkType == consts.SheetPermaLinkTypeRoot {
		contentRouter.Get("/:slug", s.wrapHTMLHandler(s.ContentSheetHandler.SheetBySlug))
	} else {
		contentRouter.Get(sheetPath+"/:slug", s.wrapHTMLHandler(s.ContentSheetHandler.SheetBySlug))
	}
	contentRouter.Get("admin_preview/"+sheetPath+"/:slug", s.wrapHTMLHandler(s.ContentSheetHandler.AdminSheetBySlug))
	return nil
}
