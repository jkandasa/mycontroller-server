package schedule

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// Add scheduler
func Add(cfg *scheduleTY.Config) error {
	return postCommand(cfg, rsTY.CommandAdd)
}

// Remove scheduler
func Remove(cfg *scheduleTY.Config) error {
	return postCommand(cfg, rsTY.CommandRemove)
}

// LoadAll makes schedulers alive
func LoadAll() {
	result, err := List(nil, nil)
	if err != nil {
		zap.L().Error("Failed to get list of schedules", zap.Error(err))
		return
	}
	schedulers := *result.Data.(*[]scheduleTY.Config)
	for index := 0; index < len(schedulers); index++ {
		cfg := schedulers[index]
		if cfg.Enabled {
			err = Add(&cfg)
			if err != nil {
				zap.L().Error("Failed to load a schedule", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
}

// UnloadAll makes stop all schedulers
func UnloadAll() {
	err := postCommand(nil, rsTY.CommandUnloadAll)
	if err != nil {
		zap.L().Error("error on unloadall scheduler command", zap.Error(err))
	}
}

// Enable scheduler
func Enable(ids []string) error {
	schedulers, err := getSchedulerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(schedulers); index++ {
		cfg := schedulers[index]
		if !cfg.Enabled {
			cfg.Enabled = true
			err = SaveAndReload(&cfg)
			if err != nil {
				zap.L().Error("error on enabling a schedule", zap.String("id", cfg.ID), zap.Error(err))
			}
		}
	}
	return nil
}

// Disable scheduler
func Disable(ids []string) error {
	schedulers, err := getSchedulerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(schedulers); index++ {
		cfg := schedulers[index]
		if cfg.Enabled {
			cfg.Enabled = false
			err = Save(&cfg)
			if err != nil {
				return err
			}
			err = Remove(&cfg)
			if err != nil {
				zap.L().Error("error on disabling a schedule", zap.String("id", cfg.ID), zap.Error(err))
			}
		}
	}
	return nil
}

// Reload scheduler
func Reload(ids []string) error {
	schedules, err := getSchedulerEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(schedules); index++ {
		cfg := schedules[index]
		err = Remove(&cfg)
		if err != nil {
			zap.L().Error("error on removing a scheduling", zap.Error(err), zap.String("id", cfg.ID))
		}
		if cfg.Enabled {
			err = Add(&cfg)
			if err != nil {
				zap.L().Error("error on adding a schedule", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
	return nil
}

func postCommand(cfg *scheduleTY.Config, command string) error {
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeScheduler,
		Command: command,
	}
	if cfg != nil {
		reqEvent.ID = cfg.ID
		reqEvent.SetData(cfg)
	}
	topic := mcbus.FormatTopic(mcbus.TopicServiceScheduler)
	return mcbus.Publish(topic, reqEvent)
}

func getSchedulerEntries(ids []string) ([]scheduleTY.Config, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	result, err := List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *result.Data.(*[]scheduleTY.Config), nil
}
