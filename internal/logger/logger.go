package logger

import (
	"errors"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger = nil

// Set up a new Zap logger. If `onlyFile` is true, set up the logger to work
// only on file, else prints on stdout
func InitLogger(logFile string, onlyFile bool) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var core zapcore.Core

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.AddSync(lumberjackLogger),
		cfg.Level,
	)

	if onlyFile {
		core = fileCore
	} else {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(cfg.EncoderConfig),
			zapcore.Lock(os.Stdout),
			cfg.Level,
		)
		core = zapcore.NewTee(fileCore, consoleCore)
	}

	logger = zap.New(core)
	return logger
}

// Return the global Zap logger after calling `InitLogger` method
func GetLogger() (*zap.Logger, error) {
	if logger == nil {
		return nil, errors.New("You must call `InitLogger()` first.")
	}
	return logger, nil
}
