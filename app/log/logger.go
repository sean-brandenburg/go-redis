package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var noOpLogger = zap.NewNop()

type Logger struct {
	zap      *zap.Logger
	metadata []zap.Field
}

// TODO: Would be nice to make a clone function so that we can clone this logger then attach zap fields to it

func NewLogger(logFile string, level zapcore.Level) (*Logger, error) {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(level)
	config.OutputPaths = []string{"stdout"}

	if logFile != "" {
		config.OutputPaths = append(config.OutputPaths, fmt.Sprint("./logs/", logFile))
	}

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return &Logger{
		zap:      logger,
		metadata: []zap.Field{},
	}, err
}

func (l *Logger) AddMetadata(data ...zap.Field) {
	l.metadata = append(l.metadata, data...)
}

func NewNoOpLogger() Logger {
	return Logger{zap: noOpLogger}
}

func (l *Logger) Close() {
	l.zap.Sync()
}

func (l Logger) Debug(msg string, fields ...zap.Field) {
	l.writer().Debug(msg, append(fields, l.metadata...)...)
}

func (l Logger) Info(msg string, fields ...zap.Field) {
	l.writer().Info(msg, append(fields, l.metadata...)...)
}

func (l Logger) Warn(msg string, fields ...zap.Field) {
	l.writer().Warn(msg, append(fields, l.metadata...)...)
}

func (l Logger) Error(msg string, fields ...zap.Field) {
	l.writer().Error(msg, append(fields, l.metadata...)...)
}

func (l Logger) Fatal(msg string, fields ...zap.Field) {
	l.writer().Fatal(msg, append(fields, l.metadata...)...)
}

func (l Logger) writer() *zap.Logger {
	// If we don't have a logger, initialize a no-op logger. This can be useful for avoiding panics during testing
	if l.zap == nil {
		return noOpLogger
	}
	return l.zap
}
