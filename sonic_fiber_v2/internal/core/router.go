package core

import (
	"github.com/gofiber/fiber/v2"

	"sonic_fiber_v2/internal/handler"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(
	app *fiber.App,
	
	// 处理器
	indexHandler *handler.IndexHandler,
	postHandler *handler.PostHandler,
	categoryHandler *handler.CategoryHandler,
	tagHandler *handler.TagHandler,
	commentHandler *handler.CommentHandler,
	userHandler *handler.UserHandler,
	adminHandler *handler.AdminHandler,
	attachmentHandler *handler.AttachmentHandler,
) {
	// 静态资源路由
	app.Static("/css", "./resources/admin/css")
	app.Static("/js", "./resources/admin/js")
	app.Static("/images", "./resources/admin/images")
	app.Static("/themes", "./resources/template/theme")
	app.Static("/uploads", "./uploads")

	// API路由组
	apiGroup := app.Group("/api")
	
	// 内容API
	contentAPI := apiGroup.Group("/content")
	contentAPI.Get("", indexHandler.Home)
	contentAPI.Get("/page/:page", indexHandler.Page)
	contentAPI.Get("/posts/:slug", postHandler.GetBySlug)
	contentAPI.Get("/categories", categoryHandler.List)
	contentAPI.Get("/categories/:slug", categoryHandler.GetBySlug)
	contentAPI.Get("/tags", tagHandler.List)
	contentAPI.Get("/tags/:slug", tagHandler.GetBySlug)
	contentAPI.Post("/comments", commentHandler.Create)
	contentAPI.Get("/search", indexHandler.Search)

	// 管理后台API
	adminAPI := apiGroup.Group("/admin")
	
	// 公开路由
	adminAPI.Get("/is_installed", adminHandler.IsInstalled)
	adminAPI.Post("/login", adminHandler.Login)
	adminAPI.Post("/install", adminHandler.Install)

	// 需要认证的路由组
	authGroup := adminAPI.Group("")
	// 这里应该添加认证中间件
	// authGroup.Use(authMiddleware.Handle)

	// 附件管理
	authGroup.Post("/attachments/upload", attachmentHandler.Upload)
	authGroup.Get("/attachments", attachmentHandler.List)
	authGroup.Delete("/attachments/:id", attachmentHandler.Delete)

	// 文章管理
	authGroup.Get("/posts", postHandler.ListAdmin)
	authGroup.Post("/posts", postHandler.Create)
	authGroup.Get("/posts/:id", postHandler.GetByID)
	authGroup.Put("/posts/:id", postHandler.Update)
	authGroup.Delete("/posts/:id", postHandler.Delete)

	// 分类管理
	authGroup.Get("/categories", categoryHandler.ListAdmin)
	authGroup.Post("/categories", categoryHandler.Create)
	authGroup.Put("/categories/:id", categoryHandler.Update)
	authGroup.Delete("/categories/:id", categoryHandler.Delete)

	// 标签管理
	authGroup.Get("/tags", tagHandler.ListAdmin)
	authGroup.Post("/tags", tagHandler.Create)
	authGroup.Put("/tags/:id", tagHandler.Update)
	authGroup.Delete("/tags/:id", tagHandler.Delete)

	// 评论管理
	authGroup.Get("/comments", commentHandler.ListAdmin)
	authGroup.Put("/comments/:id/status", commentHandler.UpdateStatus)
	authGroup.Delete("/comments/:id", commentHandler.Delete)

	// 用户管理
	authGroup.Get("/user/profile", userHandler.GetProfile)
	authGroup.Put("/user/profile", userHandler.UpdateProfile)
	authGroup.Put("/user/password", userHandler.UpdatePassword)

	// 动态路由（根据配置生成）
	registerDynamicRoutes(app, indexHandler, postHandler, categoryHandler, tagHandler)
}

// registerDynamicRoutes 注册动态路由（如归档、标签等）
func registerDynamicRoutes(
	app *fiber.App,
	indexHandler *handler.IndexHandler,
	postHandler *handler.PostHandler,
	categoryHandler *handler.CategoryHandler,
	tagHandler *handler.TagHandler,
) {
	// 归档路由
	app.Get("/archives", indexHandler.Archives)
	app.Get("/archives/:slug", postHandler.GetByArchive)

	// 标签路由
	app.Get("/tags/:slug/posts", tagHandler.GetPosts)

	// 分类路由
	app.Get("/categories/:slug/posts", categoryHandler.GetPosts)

	// 文章详情
	app.Get("/posts/:slug", postHandler.GetBySlug)
}
