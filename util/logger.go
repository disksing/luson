package util

import "go.uber.org/zap"

// Logger wraps zap.SugaredLogger.
type Logger struct {
	*zap.SugaredLogger
}

// Sync flushes cache to disk.
func (l *Logger) Sync() {
	_ = l.SugaredLogger.Sync()
}

// NewLogger creates a Logger.
func NewLogger() *Logger {
	p, _ := zap.NewDevelopment()
	return &Logger{p.Sugar()}
}
