package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"sonic_fiber_v2/config"
)

// NewLogger 创建日志实例
func NewLogger(cfg *config.Config) *zap.Logger {
	// 配置日志级别
	var level zapcore.Level
	switch cfg.Log.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建核心
	var core zapcore.Core
	if cfg.Log.OutputPath == "" || cfg.Log.OutputPath == "console" {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
	} else {
		// 输出到文件
		file, err := os.OpenFile(cfg.Log.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(file),
			level,
		)
	}

	// 创建Logger
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// 替换全局Logger
	zap.ReplaceGlobals(logger)

	return logger
}
