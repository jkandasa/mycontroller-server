package forwardpayload

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/api/action"
	fpAPI "github.com/mycontroller-org/backend/v2/pkg/api/forward_payload"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	eventML "github.com/mycontroller-org/backend/v2/pkg/model/bus/event"
	"github.com/mycontroller-org/backend/v2/pkg/model/field"
	fpml "github.com/mycontroller-org/backend/v2/pkg/model/forward_payload"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	quickIdUtils "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

var (
	queue          *queueUtils.Queue
	queueSize      = int(1000)
	workerCount    = int(1)
	topic          = ""
	subscriptionID = int64(-1)
)

// Init message process engine
func Init() error {
	queue = queueUtils.New("forward_payload", queueSize, processEvent, workerCount)

	topic = mcbus.FormatTopic(mcbus.TopicEventField)

	// on event receive add it in to local queue
	sID, err := mcbus.Subscribe(topic, onEventReceive)
	if err != nil {
		return err
	}

	subscriptionID = sID
	return nil
}

func onEventReceive(data *busML.BusData) {
	event := &eventML.Event{}
	err := data.ToStruct(event)
	if err != nil {
		zap.L().Warn("Error on convet to target type", zap.Any("topic", data.Topic), zap.Error(err))
		return
	}

	if event.EntityType != model.EntityField || event.Type != eventML.TypeUpdated {
		// this data is not for us
		return
	}

	if event.Entity == nil {
		zap.L().Warn("Received a nil data", zap.Any("event", data))
		return
	}

	field, ok := event.Entity.(field.Field)
	if !ok {
		zap.L().Warn("received non field entity", zap.Any("entityType", event.EntityType), zap.Any("entity", event.Entity))
		return
	}

	zap.L().Debug("Field data added into processing queue", zap.Any("data", field))
	status := queue.Produce(&field)
	if !status {
		zap.L().Warn("error to store the data into queue", zap.Any("data", field))
	}
}

// Close message process engine
func Close() error {
	err := mcbus.Unsubscribe(topic, subscriptionID)
	if err != nil {
		zap.L().Error("error on unsubscription", zap.Error(err), zap.String("topic", topic), zap.Int64("subscriptionId", subscriptionID))
	}
	queue.Close()
	return nil
}

// processEvent from the queue
func processEvent(item interface{}) {
	field := item.(*field.Field)

	quickID, err := quickIdUtils.GetQuickID(*field)
	if err != nil {
		zap.L().Error("unable to get quick id", zap.Error(err), zap.String("gateway", field.GatewayID), zap.String("node", field.NodeID), zap.String("source", field.SourceID), zap.String("field", field.FieldID))
		return
	}

	// fetch mapped filed for this event
	pagination := &stgml.Pagination{Limit: 50}
	filters := []stgml.Filter{
		{Key: ml.KeySrcFieldID, Operator: stgml.OperatorEqual, Value: quickID},
		{Key: ml.KeyEnabled, Operator: stgml.OperatorEqual, Value: true},
	}
	response, err := fpAPI.List(filters, pagination)
	if err != nil {
		zap.L().Error("error getting mapping data from database", zap.Error(err))
		return
	}

	if response.Count == 0 {
		return
	}

	zap.L().Debug("Starting data forwarding", zap.Any("data", field))

	mappings := *response.Data.(*[]fpml.Mapping)
	for index := 0; index < len(mappings); index++ {
		mapping := mappings[index]
		// send payload
		if mapping.SrcFieldID != mapping.DstFieldID {
			err = action.ToFieldByQuickID(mapping.DstFieldID, fmt.Sprintf("%v", field.Current.Value))
			if err != nil {
				zap.L().Error("error on sending payload", zap.Any("mapping", mapping), zap.Error(err))
			} else {
				zap.L().Debug("Data forwarded", zap.Any("mapping", mapping))
			}
		}
	}
}
