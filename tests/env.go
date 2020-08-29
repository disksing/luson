package tests

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/disksing/luson/config"
	"github.com/disksing/luson/jsonstore"
	"github.com/disksing/luson/key"
	"github.com/disksing/luson/metastore"
	"github.com/disksing/luson/service"
	"github.com/disksing/luson/util"
	"github.com/gorilla/mux"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type Env struct {
	Conf    *config.Config
	dataDir string
	server  *http.Server
	addr    string
}

func NewEnv() (*Env, error) {
	c := dig.New()
	_ = c.Provide(newMockConfig)
	_ = c.Provide(util.NewLogger)
	_ = c.Provide(config.NewDataDir)
	_ = c.Provide(newMockAPIKey)
	_ = c.Provide(metastore.NewStore)
	_ = c.Provide(jsonstore.NewStore)
	_ = c.Provide(service.NewJServer)
	_ = c.Provide(service.NewRouter)

	env := &Env{}
	err := c.Invoke(func(conf *config.Config, router *mux.Router, logger *zap.SugaredLogger) error {
		env.Conf = conf
		env.dataDir = conf.DataDir
		lis, err := net.Listen("tcp", ":0")
		if err != nil {
			return err
		}
		env.addr = lis.Addr().String()
		server := &http.Server{Handler: router}
		go logger.Debug(server.Serve(lis))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return env, nil
}

func (env *Env) Close() {
	env.server.Close()
	os.RemoveAll(env.dataDir)
}

func newMockConfig() (*config.Config, error) {
	dataDir, err := ioutil.TempDir("", "luson_test_****")
	if err != nil {
		return nil, err
	}
	return &config.Config{
		DataDir:       dataDir,
		JSONCacheSize: 10,
		MetaCacheSize: 32,
		DefaultAccess: config.Protected,
	}, nil
}

func newMockAPIKey() key.APIKey {
	return MockAPIKey
}

const MockAPIKey = "11112222-3333-4444-aaaa-bbbbccccdddd"
