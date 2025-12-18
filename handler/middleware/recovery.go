package middleware

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/go-sonic/sonic/model/dto"
)

type RecoveryMiddleware struct {
	logger *zap.Logger
}

func NewRecoveryMiddleware(logger *zap.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

func (r *RecoveryMiddleware) RecoveryWithLogger() fiber.Handler {
	logger := r.logger.WithOptions(zap.AddCallerSkip(2))

	return func(ctx *fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				//nolint:errorlint
				if ne, ok := err.(*net.OpError); ok {
					//nolint:errorlint
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if brokenPipe {
					logger.Error(ctx.Path(),
						zap.Any("error", err),
					)
				} else {
					logger.DPanic("[Recovery]  panic recovered", zap.Any("error", err))
				}

				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					// In Fiber/Fasthttp, we just return.
					return
				} else {
					code := http.StatusInternalServerError
					ctx.Status(code).JSON(&dto.BaseDTO{Status: code, Message: http.StatusText(code)})
				}
			}
		}()
		return ctx.Next()
	}
}
