package handler

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// Start notifyHandler
func Start(cfg *handlerTY.Config) error {
	return postCommand(cfg, rsTY.CommandStart)
}

// Stop notifyHandler
func Stop(cfg *handlerTY.Config) error {
	return postCommand(cfg, rsTY.CommandStop)
}

// LoadAll makes notifyHandlers alive
func LoadAll() {
	result, err := List(nil, nil)
	if err != nil {
		zap.L().Error("Failed to get list of handlers", zap.Error(err))
		return
	}
	handlers := *result.Data.(*[]handlerTY.Config)
	for index := 0; index < len(handlers); index++ {
		cfg := handlers[index]
		if cfg.Enabled {
			err = Start(&cfg)
			if err != nil {
				zap.L().Error("error on load a handler", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
}

// UnloadAll makes stop all notifyHandlers
func UnloadAll() {
	err := postCommand(nil, rsTY.CommandUnloadAll)
	if err != nil {
		zap.L().Error("error on unloadall handlers command", zap.Error(err))
	}
}

// Enable notifyHandler
func Enable(ids []string) error {
	notifyHandlers, err := getNotifyHandlerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(notifyHandlers); index++ {
		cfg := notifyHandlers[index]
		if !cfg.Enabled {
			cfg.Enabled = true
			err = SaveAndReload(&cfg)
			if err != nil {
				zap.L().Error("error on enabling a handler", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
	return nil
}

// Disable notifyHandler
func Disable(ids []string) error {
	notifyHandlers, err := getNotifyHandlerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(notifyHandlers); index++ {
		cfg := notifyHandlers[index]
		err := Stop(&cfg)
		if err != nil {
			zap.L().Error("error on disabling a handler", zap.Error(err), zap.String("id", cfg.ID))
		}
		if cfg.Enabled {
			cfg.Enabled = false
			err = Save(&cfg)
			if err != nil {
				zap.L().Error("error on saving a handler", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
	return nil
}

// Reload notifyHandler
func Reload(ids []string) error {
	notifyHandlers, err := getNotifyHandlerEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(notifyHandlers); index++ {
		cfg := notifyHandlers[index]
		if cfg.Enabled {
			err = postCommand(&cfg, rsTY.CommandReload)
			if err != nil {
				zap.L().Error("error on reload handler command", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
	return nil
}

func postCommand(cfg *handlerTY.Config, command string) error {
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeHandler,
		Command: command,
	}
	if cfg != nil {
		reqEvent.ID = cfg.ID
		reqEvent.SetData(cfg)
	}
	topic := mcbus.FormatTopic(mcbus.TopicServiceHandler)
	return mcbus.Publish(topic, reqEvent)
}

func getNotifyHandlerEntries(ids []string) ([]handlerTY.Config, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	result, err := List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *result.Data.(*[]handlerTY.Config), nil
}
