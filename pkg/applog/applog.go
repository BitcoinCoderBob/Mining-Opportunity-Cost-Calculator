package applog

import (
	"Mining-Profitability/pkg/config"

	"github.com/sirupsen/logrus"
)

func New(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	if cfg.Environment == config.EnvProduction {
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}
