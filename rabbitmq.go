package services

import (
	"fmt"
	"reflect"
	"time"

	"github.com/streadway/amqp"
)

const ServiceTypeRabbitMQ = "rabbitmq"

type RabbitMQ struct {
	Config

	healthy         bool
	connected       bool
	reconnecting    bool
	shouldReconnect bool

	connectionEvents chan *amqp.Error

	connection     *amqp.Connection
	eventCallbacks []EventCallback
}

func (r *RabbitMQ) SetConfig(config Config) {
	r.Config = config
	r.shouldReconnect = config.ReconnectEnabled
}

func (r *RabbitMQ) Connect() error {
	var credentials string

	// if we already have a connection, return
	if r.connection != nil {
		return nil
	}

	// be sure to include the credentials (if specified)
	if r.Config.Username != "" && r.Config.Password != "" {
		credentials = fmt.Sprintf("%s:%s@", r.Config.Username, r.Config.Password)
	}

	// format the amqp protocol url and try to connect
	url := fmt.Sprintf("amqp://%s%s:%d/", credentials, r.Config.Host, r.Config.Port)
	connection, err := amqp.Dial(url)
	if err != nil {
		r.dispatchEvent(Event{
			ServiceType: ServiceTypeRabbitMQ,
			Code:        ServiceCouldNotConnect,
		})

		return err
	}

	// cache the connection for later
	r.connection = connection

	// reset the value of `shouldReconnect` back to what the configuration states
	r.shouldReconnect = r.Config.ReconnectEnabled

	// make sure we have a way to process connection-related events
	// cleanup the channels
	if r.connectionEvents == nil {
		r.connectionEvents = make(chan *amqp.Error, 1)
	} else if _, ok := (<-r.connectionEvents); !ok {
		r.connectionEvents = make(chan *amqp.Error, 1)
	}

	// let the connection know we want to be informed when there are issues
	r.connection.NotifyClose(r.connectionEvents)

	// begin processing connection events
	go r.processConnectionEvents()

	// let everyone know we've connected
	r.connected = true
	r.dispatchEvent(Event{
		ServiceType: ServiceTypeRabbitMQ,
		Code:        ServiceConnected,
	})

	// let everyone know we're healthy
	r.healthy = true
	r.dispatchEvent(Event{
		ServiceType: ServiceTypeRabbitMQ,
		Code:        ServiceHealthy,
	})

	return nil
}

func (r *RabbitMQ) Disconnect() error {
	// stop any reconnections from happening
	r.shouldReconnect = false

	// skip doing work if we don't have a connection
	if r.connection == nil {
		return nil
	}

	// close the connection
	err := r.connection.Close()

	// clean up some leftovers
	r.connection = nil
	r.reconnecting = false

	// let everyone know we're unhealthy
	r.healthy = false
	r.dispatchEvent(Event{
		ServiceType: ServiceTypeRabbitMQ,
		Code:        ServiceUnhealthy,
	})

	// let everyone know we've disconnected
	r.connected = false
	r.dispatchEvent(Event{
		ServiceType: ServiceTypeRabbitMQ,
		Code:        ServiceDisconnected,
	})

	return err
}

func (r *RabbitMQ) GetClient() interface{} {
	return r.connection
}

func (r *RabbitMQ) Subscribe(callback EventCallback) {
	r.eventCallbacks = append(r.eventCallbacks, callback)
}

func (r *RabbitMQ) Unsubscribe(callback EventCallback) {
	callbacks := []EventCallback{}
	f1 := reflect.ValueOf(callback)
	p1 := f1.Pointer()

	for _, tmp := range r.eventCallbacks {
		f2 := reflect.ValueOf(tmp)
		p2 := f2.Pointer()

		if p1 == p2 {
			continue
		}

		callbacks = append(callbacks, callback)
	}

	r.eventCallbacks = callbacks
}

func (r *RabbitMQ) IsHealthy() bool {
	return r.healthy
}

func (r *RabbitMQ) IsConnected() bool {
	return r.connected
}

func (r *RabbitMQ) IsReconnecting() bool {
	return r.reconnecting
}

func (r *RabbitMQ) processConnectionEvents() {
	// possible events found here:
	// https://godoc.org/github.com/streadway/amqp#pkg-constants

	for _ = range r.connectionEvents {
		// first, disconnect from existing connections
		r.Disconnect()

		// since it wasn't a forced disconnect, put `shouldReconnect` back
		r.shouldReconnect = r.Config.ReconnectEnabled

		// begin trying to reconnect
		go r.tryToReconnect()
	}
}

func (r *RabbitMQ) dispatchEvent(event Event) {
	for _, callback := range r.eventCallbacks {
		callback(event)
	}
}

func (r *RabbitMQ) tryToReconnect() {
	if !r.shouldReconnect || r.IsReconnecting() {
		return
	}

	// let everyone know we're reconnecting
	r.reconnecting = true
	r.dispatchEvent(Event{
		ServiceType: ServiceTypeRabbitMQ,
		Code:        ServiceReconnecting,
	})

	// try the reconnection strategy
	callback := r.Config.ReconnectStrategy
	if callback == nil {
		callback = func(svc Service) bool {
			if err := svc.Connect(); err != nil {
				return false
			}

			return true
		}
	}

	successful := callback(r)

	// if we weren't successful, attempt to reschedule things
	if !successful {
		// calculate when to start the next reconnect
		interval := r.Config.ReconnectIntervalMilliseconds
		if interval == 0 {
			interval = 1000
		}

		time.Sleep(time.Millisecond * time.Duration(interval))
		r.reconnecting = false
		go r.tryToReconnect()
	} else {
		r.reconnecting = false
	}
}
