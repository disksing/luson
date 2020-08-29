package config

import (
	"flag"
	"os"

	"go.uber.org/zap"
)

var dataDir = flag.String("data-dir", "data", "data directory")
var jsonCache = flag.Int("json-cache", 100, "json cache size, default 100M")
var metaCache = flag.Int("meta-cache", 1024, "meta cache limit, default 1024")

type Config struct {
	DataDir       string
	JSONCacheSize int
	MetaCacheSize int
}

func NewConfig() *Config {
	flag.Parse()
	return &Config{
		DataDir:       *dataDir,
		JSONCacheSize: *jsonCache,
		MetaCacheSize: *metaCache,
	}
}

type DataDir string

func NewDataDir(conf *Config, logger *zap.SugaredLogger) (DataDir, error) {
	_, err := os.Stat(conf.DataDir)
	if os.IsNotExist(err) {
		logger.Infow("data dir not exist, creating")
		err = os.Mkdir(conf.DataDir, os.ModePerm)
		if err != nil {
			logger.Errorf("failed to create data dir %s", *dataDir, zap.Error(err))
			return "", err
		}
		logger.Info("data dir created")
	}
	return DataDir(conf.DataDir), nil
}
