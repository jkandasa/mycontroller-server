package allinone

import (
	"github.com/mycontroller-org/backend/v2/pkg/init/common"
	"github.com/mycontroller-org/backend/v2/pkg/init/core"
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	fwdplSVC "github.com/mycontroller-org/backend/v2/pkg/service/forward_payload"
	gwService "github.com/mycontroller-org/backend/v2/pkg/service/gateway"
	msgProcessor "github.com/mycontroller-org/backend/v2/pkg/service/message_processor"
	mts "github.com/mycontroller-org/backend/v2/pkg/service/metrics"
	resourceSVC "github.com/mycontroller-org/backend/v2/pkg/service/resource"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	"go.uber.org/zap"
)

// Init func
func Init(handlerFunc func()) {
	common.InitBasicServices(wrapHandlerFunc(handlerFunc), closeServices)
}

func initServices() {
	stg.Init() // storage
	mts.Init() // metrics

	core.StartupJobs(&cfg.CFG.StartupJobs)
	core.UpdateInitialUser()

	// start message processing engine
	msgProcessor.Init()

	// init resource server
	err := resourceSVC.Init()
	if err != nil {
		zap.L().Fatal("Error on init resource service listener", zap.Error(err))
	}

	// init payload forward service
	err = fwdplSVC.Init()
	if err != nil {
		zap.L().Fatal("Error on init forward payload service", zap.Error(err))
	}

	// init gateway listener
	err = gwService.Init(cfg.CFG.Gateway)
	if err != nil {
		zap.L().Fatal("Failed to init gateway service listener", zap.Error(err))
	}
}

func wrapHandlerFunc(handlerFunc func()) func() {
	return func() {
		initServices()
		if handlerFunc != nil {
			go handlerFunc()
		}
	}
}

func closeServices() {
	// close forward payload service
	fwdplSVC.Close()

	// close gateway service
	gwService.Close()

	// close resource service
	resourceSVC.Close()

	// stop engine
	zap.L().Debug("Closing message process engine")
	msgProcessor.Close()

	// Close storage and metric database
	if stg.SVC != nil {
		err := stg.SVC.Close()
		if err != nil {
			zap.L().Error("Failed to close storage database")
		}
	}
	if mts.SVC != nil {
		if mts.SVC != nil {
			err := mts.SVC.Close()
			if err != nil {
				zap.L().Error("Failed to close metrics database")
			}
		}
	}
}
