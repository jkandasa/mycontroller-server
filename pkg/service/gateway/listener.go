package service

import (
	q "github.com/jaegertracing/jaeger/pkg/queue"
	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	rsml "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	"go.uber.org/zap"
)

var (
	eventQueue *q.BoundedQueue
	queueSize  = int(50)
	cfg        *Config
)

// Config of gateway service
type Config struct {
	IDs    []string
	Labels cmap.CustomStringMap
}

// Init starts resource server listener
func Init(config cmap.CustomMap) error {
	cfg = &Config{}
	err := utils.MapToStruct(utils.TagNameNone, config, cfg)
	if err != nil {
		return err
	}

	eventQueue = utils.GetQueue("gateway_service", queueSize)

	// on event receive add it in to our local queue
	topic := mcbus.FormatTopic(mcbus.TopicServiceGateway)
	_, err = mcbus.Subscribe(topic, onEvent)
	if err != nil {
		return err
	}

	eventQueue.StartConsumers(1, processEvent)

	// load gateways
	reqEvent := rsml.Event{
		Type:    rsml.TypeGateway,
		Command: rsml.CommandLoadAll,
	}
	topicResourceServer := mcbus.FormatTopic(mcbus.TopicServiceResourceServer)
	return mcbus.Publish(topicResourceServer, reqEvent)
}

// Close the service
func Close() {
	UnloadAll()
	eventQueue.Stop()
}

func onEvent(event *busML.BusData) {
	reqEvent := &rsml.Event{}
	err := event.ToStruct(reqEvent)
	if err != nil {
		zap.L().Warn("Failed to convet to target type", zap.Error(err))
		return
	}
	if reqEvent == nil {
		zap.L().Warn("Received a nil message", zap.Any("event", event))
		return
	}
	zap.L().Debug("Event added into processing queue", zap.Any("event", reqEvent))
	status := eventQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("Failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processEvent from the queue
func processEvent(event interface{}) {
	reqEvent := event.(*rsml.Event)
	zap.L().Debug("Processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsml.TypeGateway {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsml.CommandStart:
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil && helper.IsMine(cfg.IDs, cfg.Labels, gwCfg.ID, gwCfg.Labels) {
			Start(gwCfg)
		}

	case rsml.CommandStop:
		if reqEvent.ID != "" {
			Stop(reqEvent.ID)
			return
		}
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil {
			Stop(gwCfg.ID)
		}

	case rsml.CommandReload:
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil {
			Stop(gwCfg.ID)
			if helper.IsMine(cfg.IDs, cfg.Labels, gwCfg.ID, gwCfg.Labels) {
				Start(gwCfg)
			}
		}

	case rsml.CommandUnloadAll:
		UnloadAll()

	default:
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func getGatewayConfig(reqEvent *rsml.Event) *gwml.Config {
	gwCfg := &gwml.Config{}
	err := reqEvent.ToStruct(gwCfg)
	if err != nil {
		zap.L().Error("Error on data conversion", zap.Error(err))
		return nil
	}
	return gwCfg
}
