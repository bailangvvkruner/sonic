package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type FiberLoggerMiddleware struct {
	logger *zap.Logger
}

func NewFiberLoggerMiddleware(logger *zap.Logger) *FiberLoggerMiddleware {
	return &FiberLoggerMiddleware{
		logger: logger,
	}
}

// FiberLoggerConfig LoggerConfig defines the config for Logger middleware
type FiberLoggerConfig struct {
	// SkipPaths is an url path array which logs are not written.
	// Optional.
	SkipPaths []string
}

// LoggerWithConfig instance a Logger middleware with config.
func (g *FiberLoggerMiddleware) LoggerWithConfig(conf FiberLoggerConfig) fiber.Handler {
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
		// Start timer
		start := time.Now()
		path := ctx.Path()
		raw := ctx.OriginalURL()

		// Process request
		err := ctx.Next()

		if err != nil {
			logger.Error(err.Error())
		}
		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			if raw != "" && raw != path {
				path = raw
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
