package main

import (
	"net/http"

	"github.com/disksing/luson/config"
	"github.com/disksing/luson/jsonstore"
	"github.com/disksing/luson/key"
	"github.com/disksing/luson/metastore"
	"github.com/disksing/luson/service"
	"github.com/disksing/luson/util"
	"github.com/gorilla/mux"
	"go.uber.org/dig"
)

func main() {
	c := dig.New()
	_ = c.Provide(config.NewConfig)
	_ = c.Provide(util.NewLogger)
	_ = c.Provide(config.NewDataDir)
	_ = c.Provide(key.NewAPIKey)
	_ = c.Provide(metastore.NewStore)
	_ = c.Provide(jsonstore.NewStore)
	_ = c.Provide(service.NewJServer)
	_ = c.Provide(service.NewRouter)
	_ = c.Invoke(start)
}

func start(logger *util.Logger, router *mux.Router) {
	logger.Info("ready to start")
	logger.Error(http.ListenAndServe(":42195", router))
}
