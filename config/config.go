package config

import (
	"flag"
	"os"

	"github.com/disksing/luson/util"
	"go.uber.org/zap"
)

var dataDir = flag.String("data-dir", "data", "data directory")
var jsonCache = flag.Int("json-cache", 100, "json cache size, default 100M")
var metaCache = flag.Int("meta-cache", 1024, "meta cache limit, default 1024")
var defaultAccess = flag.String("default-access", "protected", "public/protected/private")

const (
	Public    string = "public"    // everyone can read/write
	Protected string = "protected" // everyone can read, write with api key
	Private   string = "private"   // read/write with api key
)

func ValidateAccess(s string) bool {
	return s == Public || s == Protected || s == Private
}

type Config struct {
	DataDir       string
	JSONCacheSize int
	MetaCacheSize int
	DefaultAccess string
}

func NewConfig() *Config {
	flag.Parse()

	if !ValidateAccess(*defaultAccess) {
		*defaultAccess = Protected
	}

	return &Config{
		DataDir:       *dataDir,
		JSONCacheSize: *jsonCache,
		MetaCacheSize: *metaCache,
		DefaultAccess: *defaultAccess,
	}
}

type DataDir string

func NewDataDir(conf *Config, logger *util.Logger) (DataDir, error) {
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
