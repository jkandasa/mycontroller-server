package bus

import busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"

// CallBackFunc message passed to this func
type CallBackFunc func(data *busTY.BusData)

// Plugin interface
type Plugin interface {
	Name() string
	Close() error
	Publish(topic string, data interface{}) error
	Subscribe(topic string, handler CallBackFunc) (int64, error)
	Unsubscribe(topic string, subscriptionID int64) error
	QueueSubscribe(topic, queueName string, handler CallBackFunc) (int64, error)
	QueueUnsubscribe(topic, queueName string, subscriptionID int64) error
	UnsubscribeAll(topic string) error
}
