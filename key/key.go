package key

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/disksing/luson/config"
	"github.com/disksing/luson/util"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type APIKey string

func NewAPIKey(dataDir config.DataDir, logger *util.Logger) (APIKey, error) {
	defer logger.Sync()
	f := path.Join(string(dataDir), fname)
	data, err := ioutil.ReadFile(f)
	if os.IsNotExist(err) {
		logger.Infof("%s not exist, creating", f)
		k := uuid.NewV4().String()
		err = ioutil.WriteFile(f, []byte(k), 0600)
		if err != nil {
			logger.Errorw("failed to persist api key", zap.Error(err))
			return "", err
		}
		logger.Infof("api key saved to %s", f)
		return APIKey(k), nil
	}
	if err != nil {
		logger.Errorw("failed to open api key file", zap.Error(err))
		return "", err
	}
	logger.Info("api key loaded")
	return APIKey(data), nil
}

const fname = "api-key"
