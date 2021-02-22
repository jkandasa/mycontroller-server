package status

import (
	"github.com/mycontroller-org/backend/v2/pkg/model"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	scheduleML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"go.uber.org/zap"
)

// SetGatewayState send gateway status into bus
func SetGatewayState(id string, state model.State) {
	PostData(id, state, rsML.TypeGateway, rsML.CommandUpdateState)
}

// SetHandlerState send handler status into bus
func SetHandlerState(id string, state model.State) {
	PostData(id, state, rsML.TypeNotifyHandler, rsML.CommandUpdateState)
}

// SetTaskState send handler status into bus
func SetTaskState(id string, state taskML.State) {
	PostData(id, state, rsML.TypeTask, rsML.CommandUpdateState)
}

// SetScheduleState send handler status into bus
func SetScheduleState(id string, state scheduleML.State) {
	PostData(id, state, rsML.TypeScheduler, rsML.CommandUpdateState)
}

// DisableSchedule sends id to resource service
func DisableSchedule(id string) {
	PostData(id, id, rsML.TypeScheduler, rsML.CommandDisable)
}

// DisableTask sends id to resource service
func DisableTask(id string) {
	PostData(id, id, rsML.TypeTask, rsML.CommandDisable)
}

// PostData to resource service
func PostData(id string, data interface{}, serviceType string, command string) {
	event := &rsML.Event{
		Type:    serviceType,
		Command: command,
		ID:      id,
	}
	event.SetData(data)
	topic := mcbus.FormatTopic(mcbus.TopicServiceResourceServer)
	err := mcbus.Publish(topic, event)
	if err != nil {
		zap.L().Error("failed to post an event", zap.String("topic", topic), zap.Any("event", event))
	}
}
