package mcbus

import busml "github.com/mycontroller-org/backend/v2/plugin/bus"

// Close func
func Close() error {
	if busClient != nil {
		return busClient.Close()
	}
	return nil
}

// Publish a data to a topic
func Publish(topic string, data interface{}) error {
	return busClient.Publish(topic, data)
}

// Subscribe a topic
func Subscribe(topic string, handler func(event *busml.Event)) error {
	return busClient.Subscribe(topic, handler)
}

// Unsubscribe a topic
func Unsubscribe(topic string) error {
	return busClient.Unsubscribe(topic)
}
