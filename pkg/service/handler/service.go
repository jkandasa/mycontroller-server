package handler

import (
	"fmt"
	"time"

	commonStore "github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	handlerPlugin "github.com/mycontroller-org/server/v2/plugin/handler"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// StartHandler notify handlers
func StartHandler(cfg *handlerTY.Config) error {
	if handlersStore.Get(cfg.ID) != nil {
		return fmt.Errorf("a service is in running state. id:%s", cfg.ID)
	}
	if !cfg.Enabled { // this handler is not enabled
		return nil
	}
	zap.L().Debug("starting a handler", zap.Any("id", cfg.ID))
	state := types.State{Since: time.Now()}

	handler, err := loadHandler(cfg)
	if err != nil {
		return err
	}
	err = handler.Start()
	if err != nil {
		zap.L().Error("unable to start a handler service", zap.Any("id", cfg.ID), zap.Error(err))
		state.Message = err.Error()
		state.Status = types.StatusDown
	} else {
		state.Message = "started successfully"
		state.Status = types.StatusUp
		handlersStore.Add(cfg.ID, handler)
	}

	busUtils.SetHandlerState(cfg.ID, state)
	return nil
}

// StopHandler a handler
func StopHandler(id string) error {
	zap.L().Debug("stopping a handler", zap.Any("id", id))
	handler := handlersStore.Get(id)
	if handler != nil {
		err := handler.Close()
		state := types.State{
			Status:  types.StatusDown,
			Since:   time.Now(),
			Message: "stopped by request",
		}
		if err != nil {
			zap.L().Error("failed to stop handler service", zap.String("id", id), zap.Error(err))
			state.Message = err.Error()
		}
		busUtils.SetHandlerState(id, state)
		handlersStore.Remove(id)
	}
	return nil
}

// ReloadHandler a handler
func ReloadHandler(gwCfg *handlerTY.Config) error {
	err := StopHandler(gwCfg.ID)
	if err != nil {
		return err
	}
	utils.SmartSleep(1 * time.Second)
	return StartHandler(gwCfg)
}

// UnloadAll stops all handlers
func UnloadAll() {
	ids := handlersStore.ListIDs()
	for _, id := range ids {
		err := StopHandler(id)
		if err != nil {
			zap.L().Error("error on stopping a handler", zap.String("id", id))
		}
	}
}

func loadHandler(cfg *handlerTY.Config) (handlerTY.Plugin, error) {
	// descrypt the secrets, tokens
	err := cloneUtil.UpdateSecrets(cfg, commonStore.CFG.Secret, "", false, cloneUtil.DefaultSpecialKeys)
	if err != nil {
		return nil, err
	}

	handler, err := handlerPlugin.Create(cfg.Type, cfg)
	if err != nil {
		return nil, err
	}
	return handler, nil
}
