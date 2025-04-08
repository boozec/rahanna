package logger

import (
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger = nil

func InitLogger(logFile string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{logFile}
	cfg.ErrorOutputPaths = []string{logFile}

	// Configure lumberjack for log rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100, // megabytes
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.AddSync(lumberjackLogger), // Log only to the file via lumberjack
		cfg.Level,
	)

	logger = zap.New(core)

	return logger
}

func GetLogger() (*zap.Logger, error) {
	if logger == nil {
		return nil, errors.New("You must call `InitLogger()` first.")
	}
	return logger, nil
}
