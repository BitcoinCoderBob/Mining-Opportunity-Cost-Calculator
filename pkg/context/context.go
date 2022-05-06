package context

import (
	"Mining-Profitability/cmd/calc"
	"Mining-Profitability/cmd/externaldata"
	"Mining-Profitability/pkg/config"
	util "Mining-Profitability/pkg/utils"
	"context"

	"github.com/sirupsen/logrus"
)

type AppContext struct {
	Logger       *logrus.Logger
	Calc         calc.Interface
	Utils        util.Interface
	ExternalData externaldata.Interface
	Ctx          context.Context
}

func New(cfg *config.Config, logger *logrus.Logger) (*AppContext, context.CancelFunc, error) {
	logger.Debug("setting up context")
	calc := calc.New(cfg)
	externalData := externaldata.New(cfg)
	utils := util.New()
	ctx, cancel := context.WithCancel(context.Background())

	return &AppContext{
		Logger:       logger,
		Calc:         calc,
		Utils:        utils,
		ExternalData: externalData,
		Ctx:          ctx,
	}, cancel, nil
}
