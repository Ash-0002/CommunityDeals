package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Init initializes the global logger. Call once at startup.
func Init(env string) {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var err error
	log, err = cfg.Build()
	if err != nil {
		// Fallback to a minimal logger using os.Stderr
		encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapcore.DebugLevel)
		log = zap.New(core)
	}
}

// Get returns the global logger. Panics if Init has not been called.
func Get() *zap.Logger {
	if log == nil {
		panic("logger.Init() must be called before using logger.Get()")
	}
	return log
}

// Sugar returns the sugared (printf-style) global logger.
func Sugar() *zap.SugaredLogger {
	return Get().Sugar()
}

// Sync flushes any buffered log entries. Call on graceful shutdown.
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}
