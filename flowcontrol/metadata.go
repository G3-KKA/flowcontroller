package flowcontrol

import (
	"context"
	"flowcontroller/config"
	"flowcontroller/logger"
)

type ServiceMetadata struct {
	// Not-Empty Config is guaranteed
	cfg config.Config
	// Not-Nil and Initialized Logger is guaranteed
	logger logger.Logger
	// Never Cancelled Context is guaranteed
	ctx context.Context
}

func (mtd *ServiceMetadata) Cfg() config.Config {
	return mtd.cfg
}

func (mtd *ServiceMetadata) Logger() logger.Logger { // returns pointer, not copy, unsafe
	return mtd.logger
}

// May only contain values, never cancelled, never returns an error
func (mtd *ServiceMetadata) Context() context.Context {
	return context.WithoutCancel(mtd.ctx)
}
