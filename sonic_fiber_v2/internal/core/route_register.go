package core

import (
	"github.com/gofiber/fiber/v2"

	"sonic_fiber_v2/internal/handler"
)

// RouteRegister 路由注册器
type RouteRegister struct {
	app            *fiber.App
	indexHandler   *handler.IndexHandler
	postHandler    *handler.PostHandler
	categoryHandler *handler.CategoryHandler
	tagHandler     *handler.TagHandler
	commentHandler *handler.CommentHandler
	userHandler    *handler.UserHandler
	adminHandler   *handler.AdminHandler
	attachmentHandler *handler.AttachmentHandler
}

func NewRouteRegister(
	app *fiber.App,
	indexHandler *handler.IndexHandler,
	postHandler *handler.PostHandler,
	categoryHandler *handler.CategoryHandler,
	tagHandler *handler.TagHandler,
	commentHandler *handler.CommentHandler,
	userHandler *handler.UserHandler,
	adminHandler *handler.AdminHandler,
	attachmentHandler *handler.AttachmentHandler,
) *RouteRegister {
	return &RouteRegister{
		app:            app,
		indexHandler:   indexHandler,
		postHandler:    postHandler,
		categoryHandler: categoryHandler,
		tagHandler:     tagHandler,
		commentHandler: commentHandler,
		userHandler:    userHandler,
		adminHandler:   adminHandler,
		attachmentHandler: attachmentHandler,
	}
}

func (r *RouteRegister) Register() {
	RegisterRoutes(
		r.app,
		r.indexHandler,
		r.postHandler,
		r.categoryHandler,
		r.tagHandler,
		r.commentHandler,
		r.userHandler,
		r.adminHandler,
		r.attachmentHandler,
	)
}

// InvokeRouteRegister 用于fx.Invoke
func InvokeRouteRegister(register *RouteRegister) {
	register.Register()
}
