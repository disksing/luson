package main

import (
	"net/http"

	"github.com/disksing/luson/config"
	"github.com/disksing/luson/jsonstore"
	"github.com/disksing/luson/key"
	"github.com/disksing/luson/metastore"
	"github.com/disksing/luson/service"
	"github.com/gorilla/mux"
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
	c.Provide(service.NewJServer)
	c.Provide(service.NewRouter)
	if err := c.Invoke(start); err != nil {
		newLogger().Info("failed to start", err)
	}
}

func newLogger() *zap.SugaredLogger {
	p, _ := zap.NewDevelopment()
	return p.Sugar()
}

func start(logger *zap.SugaredLogger, router *mux.Router) {
	logger.Info("ready to start")
	http.ListenAndServe(":42195", router)
}
