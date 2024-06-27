package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var noOpLogger = zap.NewNop()

type Logger struct {
	zap *zap.Logger
}

func NewLogger(logFile string, level zapcore.Level) (*Logger, error) {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(level)
	config.OutputPaths = []string{"stdout"}

	if logFile != "" {
		config.OutputPaths = append(config.OutputPaths, fmt.Sprint("./logs/", logFile))
	}

	logger, err := config.Build(zap.AddCaller())
	if err != nil {
		return nil, err
	}
	return &Logger{zap: logger}, err
}
func NewTestLogger() Logger {
	return Logger{zap: noOpLogger}
}

func (l *Logger) Close() {
	l.zap.Sync()
}

func (l Logger) Debug(msg string, fields ...zap.Field) {
	l.writer().Debug(msg, fields...)
}

func (l Logger) Info(msg string, fields ...zap.Field) {
	l.writer().Info(msg, fields...)
}

func (l Logger) Warn(msg string, fields ...zap.Field) {
	l.writer().Warn(msg, fields...)
}

func (l Logger) Error(msg string, fields ...zap.Field) {
	l.writer().Warn(msg, fields...)
}

func (l Logger) Fatal(msg string, fields ...zap.Field) {
	l.writer().Warn(msg, fields...)
}

func (l Logger) writer() *zap.Logger {
	// If we don't have a logger, initialize a no-op logger. This can be useful for avoiding panics during testing
	if l.zap == nil {
		return noOpLogger
	}
	return l.zap
}
