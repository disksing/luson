package main

import (
	"github.com/disksing/luson/config"
	"github.com/disksing/luson/jsonstore"
	"github.com/disksing/luson/key"
	"github.com/disksing/luson/metastore"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func main() {
	c := dig.New()
	c.Provide(config.NewConfig)
	c.Provide(newLogger)
	c.Provide(config.NewDataDir)
	c.Provide(key.NewAPIKey)
	c.Provide(metastore.NewStore)
	c.Provide(jsonstore.NewStore)
	c.Invoke(start)
}

func newLogger() *zap.SugaredLogger {
	p, _ := zap.NewDevelopment()
	return p.Sugar()
}

func start(logger *zap.SugaredLogger, key key.APIKey) {
	logger.Infof("key is %v", key)
	logger.Info("started.")
}
