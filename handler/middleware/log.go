package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/go-sonic/sonic/util"
)

type GinLoggerMiddleware struct {
	logger *zap.Logger
}

func NewGinLoggerMiddleware(logger *zap.Logger) *GinLoggerMiddleware {
	return &GinLoggerMiddleware{
		logger: logger,
	}
}

// GinLoggerConfig LoggerConfig defines the config for Logger middleware
type GinLoggerConfig struct {
	// SkipPaths is an url path array which logs are not written.
	// Optional.
	SkipPaths []string
}

// LoggerWithConfig instance a Logger middleware with config.
func (g *GinLoggerMiddleware) LoggerWithConfig(conf GinLoggerConfig) fiber.Handler {
	logger := g.logger.WithOptions(zap.WithCaller(false))
	notLogged := conf.SkipPaths

	var skip map[string]struct{}

	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notLogged {
			skip[path] = struct{}{}
		}
	}

	return func(ctx *fiber.Ctx) error {
		// Populate context with IP and UserAgent
		userCtx := ctx.UserContext()
		if userCtx == nil {
			userCtx = context.Background()
		}
		userCtx = util.SetClientIP(userCtx, ctx.IP())
		userCtx = util.SetUserAgent(userCtx, ctx.Get("User-Agent"))
		ctx.SetUserContext(userCtx)

		// Start timer
		start := time.Now()
		path := ctx.Path()
		raw := string(ctx.Request().URI().QueryString())

		// Process request
		err := ctx.Next()

		if err != nil {
			logger.Error(err.Error())
		}
		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			if raw != "" {
				path = path + "?" + raw
			}
			path = strings.ReplaceAll(path, "\n", "")
			path = strings.ReplaceAll(path, "\r", "")
			clientIP := strings.ReplaceAll(ctx.IP(), "\n", "")
			clientIP = strings.ReplaceAll(clientIP, "\r", "")

			logger.Info("[FIBER]",
				zap.Time("beginTime", start),
				zap.Int("status", ctx.Response().StatusCode()),
				zap.Duration("latency", time.Since(start)),
				zap.String("clientIP", clientIP),
				zap.String("method", ctx.Method()),
				zap.String("path", path))
		}
		return err
	}
}
