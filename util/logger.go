package util

import "go.uber.org/zap"

func NewLogger() *zap.SugaredLogger {
	p, _ := zap.NewDevelopment()
	return p.Sugar()
}
