package busutils

import (
	"encoding/json"

	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// PostToHandler send data to handlers
func PostToHandler(handlers []string, data map[string]string) {
	zap.L().Debug("Posting data to handlers", zap.Any("handlers", handlers))

	// remove disabled parameters
	updateData := make(map[string]interface{})
	for name, value := range data {
		genericData := handlerType.GenericData{}
		err := json.Unmarshal([]byte(value), &genericData)
		if err != nil {
			continue
		}

		if converterUtils.ToBool(genericData.Disabled) {
			zap.L().Debug("parameter disabled", zap.String("name", name), zap.String("type", genericData.Type), zap.String("disabled", genericData.Disabled))
			continue
		}
		// update to our new list if the item is not disabled
		updateData[name] = value
	}

	if len(updateData) == 0 {
		return
	}

	for _, handlerID := range handlers {
		if handlerID == "" {
			continue
		}
		msg := &handlerType.MessageWrapper{
			ID:   handlerID,
			Data: updateData,
		}
		err := mcbus.Publish(mcbus.FormatTopic(mcbus.TopicPostMessageNotifyHandler), msg)
		if err != nil {
			zap.L().Error("error on posting data to handler", zap.Error(err), zap.String("handlerID", handlerID))
		}
	}
}
