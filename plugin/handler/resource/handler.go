package resource

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	"github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	yamlUtils "github.com/mycontroller-org/server/v2/pkg/utils/yaml"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

const (
	PluginResourceHandler = "resource"

	schedulePrefix = "resource_handler"
)

// ResourceClient struct
type ResourceClient struct {
	HandlerCfg *handlerTY.Config
	store      *store
}

func NewResourcePlugin(config *handlerTY.Config) (handlerTY.Plugin, error) {
	return &ResourceClient{
		HandlerCfg: config,
		store:      &store{mutex: sync.RWMutex{}, handlerID: config.ID, jobs: map[string]JobsConfig{}},
	}, nil
}

func (p *ResourceClient) Name() string {
	return PluginResourceHandler
}

// Start handler implementation
func (c *ResourceClient) Start() error {
	// load handler data from disk
	err := c.store.loadFromDisk(c)
	if err != nil {
		zap.L().Error("failed to load handler data", zap.String("diskLocation", c.store.getName()), zap.Error(err))
		return nil
	}

	return nil
}

// Close handler implementation
func (c *ResourceClient) Close() error {
	c.unloadAll()

	// save jobs to disk location
	err := c.store.saveToDisk()
	if err != nil {
		zap.L().Error("failed to save handler data", zap.String("diskLocation", c.store.getName()), zap.Error(err))
	}
	return nil
}

// State implementation
func (c *ResourceClient) State() *types.State {
	if c.HandlerCfg != nil {
		if c.HandlerCfg.State == nil {
			c.HandlerCfg.State = &types.State{}
		}
		return c.HandlerCfg.State
	}
	return &types.State{}
}

// Post handler implementation
func (c *ResourceClient) Post(data map[string]interface{}) error {
	for name, value := range data {
		stringValue, ok := value.(string)
		if !ok {
			continue
		}

		genericData := handlerTY.GenericData{}
		err := json.Unmarshal([]byte(stringValue), &genericData)
		if err != nil {
			continue
		}

		if !strings.HasPrefix(genericData.Type, handlerTY.DataTypeResource) {
			continue
		}

		rsData := handlerTY.ResourceData{}
		err = yamlUtils.UnmarshalBase64Yaml(genericData.Data, &rsData)
		if err != nil {
			zap.L().Error("error on loading resource data", zap.Error(err), zap.String("name", name), zap.String("input", stringValue))
			continue
		}

		if rsData.PreDelay != "" {
			delayDuration, err := time.ParseDuration(rsData.PreDelay)
			if err != nil {
				return fmt.Errorf("invalid preDelay. name:%s, preDelay:%s", name, rsData.PreDelay)
			}
			if delayDuration > 0 {
				c.store.add(name, rsData)
				c.schedule(name, rsData)
				continue
			}
		}

		zap.L().Debug("about to perform an action", zap.String("rawData", stringValue), zap.Any("finalData", rsData))
		busUtils.PostToResourceService("resource_fake_id", rsData, rsTY.TypeResourceAction, rsTY.CommandSet, "")
	}
	return nil
}

// preDelay scheduler helpers

func (c *ResourceClient) getScheduleTriggerFunc(name string, rsData handlerTY.ResourceData) func() {
	return func() {
		// disable the schedule
		c.unschedule(name)

		// call the resource action
		zap.L().Debug("scheduler triggered. about to perform an action", zap.String("name", name), zap.Any("rsData", rsData))
		busUtils.PostToResourceService("resource_fake_id", rsData, rsTY.TypeResourceAction, rsTY.CommandSet, "")
	}
}

func (c *ResourceClient) schedule(name string, rsData handlerTY.ResourceData) {
	c.unschedule(name) // removes the existing schedule, if any

	schedulerID := c.getScheduleID(name)
	cronSpec := fmt.Sprintf("@every %s", rsData.PreDelay)
	err := coreScheduler.SVC.AddFunc(schedulerID, cronSpec, c.getScheduleTriggerFunc(name, rsData))
	if err != nil {
		zap.L().Error("error on adding schedule", zap.Error(err))
	}
	zap.L().Debug("added a schedule", zap.String("name", name), zap.String("schedulerID", schedulerID), zap.Any("resourceData", rsData))
}

func (c *ResourceClient) unschedule(name string) {
	schedulerID := c.getScheduleID(name)
	coreScheduler.SVC.RemoveFunc(schedulerID)
	zap.L().Debug("removed a schedule", zap.String("name", name), zap.String("schedulerID", schedulerID))
}

func (c *ResourceClient) unloadAll() {
	coreScheduler.SVC.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, c.HandlerCfg.ID))
}

func (c *ResourceClient) getScheduleID(name string) string {
	return fmt.Sprintf("%s_%s_%s", schedulePrefix, c.HandlerCfg.ID, name)
}
