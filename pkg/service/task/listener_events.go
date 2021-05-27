package task

import (
	"github.com/mycontroller-org/backend/v2/pkg/model"
	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	eventML "github.com/mycontroller-org/backend/v2/pkg/model/bus/event"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

type eventWrapper struct {
	Event *eventML.Event
	Tasks []taskML.Config
}

const (
	eventListenerPreQueueLimit   = 1000
	eventListenerPostQueueLimit  = 1000
	eventListenerPreWorkerLimit  = 5
	eventListenerPostWorkerLimit = 1
	eventListenerPreQueueName    = "task_event_listener_pre"
	eventListenerPostQueueName   = "task_event_listener_post"
)

var (
	preEventsQueue          *queueUtils.Queue
	postEventsQueue         *queueUtils.Queue
	preEventsSubscriptionID = int64(0)
	preEventsTopic          = "" // updated dynamically
)

// initEventListener events listener
func initEventListener() error {
	preEventsQueue = queueUtils.New(eventListenerPreQueueName, eventListenerPreQueueLimit, processPreEvent, eventListenerPreWorkerLimit)
	postEventsQueue = queueUtils.New(eventListenerPostQueueName, eventListenerPostQueueLimit, resourcePostProcessor, eventListenerPostWorkerLimit)

	// on message receive add it in to our local queue
	preEventsTopic = mcbus.FormatTopic(mcbus.TopicEventsAll)
	sID, err := mcbus.Subscribe(preEventsTopic, onEventReceive)
	if err != nil {
		return err
	}
	preEventsSubscriptionID = sID
	return nil
}

func closeEventListener() error {
	err := mcbus.Unsubscribe(preEventsTopic, preEventsSubscriptionID)
	if err != nil {
		return err
	}
	preEventsQueue.Close()
	postEventsQueue.Close()
	return nil
}

func onEventReceive(busData *busML.BusData) {
	status := preEventsQueue.Produce(busData)
	if !status {
		zap.L().Warn("Failed to store the event into queue", zap.Any("event", busData))
	}
}

func processPreEvent(item interface{}) {
	busData := item.(*busML.BusData)

	event := &eventML.Event{}
	err := busData.LoadData(event)
	if err != nil {
		zap.L().Warn("Error on convet to target type", zap.Any("topic", busData.Topic), zap.Error(err))
		return
	}

	// supported entity events
	switch event.EntityType {
	case
		model.EntityGateway,
		model.EntityNode,
		model.EntitySource,
		model.EntityField:
		// continue

	default:
		// return do not proceed further
		return
	}

	resourceWrapper := &eventWrapper{Event: event}
	err = resourcePreProcessor(resourceWrapper)
	if err != nil {
		zap.L().Error("Error on executing a resource", zap.Any("resource", resourceWrapper), zap.Error(err))
		return
	}

	if len(resourceWrapper.Tasks) > 0 {
		status := postEventsQueue.Produce(resourceWrapper)
		if !status {
			zap.L().Error("failed to post selected tasks on post processor queue")
		}
	}
}

func resourcePreProcessor(evntWrapper *eventWrapper) error {
	zap.L().Debug("eventWrapper received", zap.Any("eventWrapper", evntWrapper))

	tasks := tasksStore.filterTasks(evntWrapper)
	zap.L().Debug("filtered", zap.Any("numberOftasks", len(tasks)))

	for index := 0; index < len(tasks); index++ {
		task := tasks[index]
		zap.L().Debug("executing task", zap.String("id", task.ID), zap.String("description", task.Description))
		if len(tasks) > 0 {
			evntWrapper.Tasks = tasks
		}
	}
	return nil
}

func resourcePostProcessor(item interface{}) {
	evntWrapper, ok := item.(*eventWrapper)
	if !ok {
		zap.L().Warn("supplied item is not resourceWrapper", zap.Any("item", item))
		return
	}

	zap.L().Debug("resourceWrapper received", zap.String("entityType", evntWrapper.Event.EntityType))

	for index := 0; index < len(evntWrapper.Tasks); index++ {
		task := evntWrapper.Tasks[index]
		executeTask(&task, evntWrapper)
	}
}
