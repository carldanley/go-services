package services

import (
	"fmt"
	"reflect"
	"time"

	"github.com/nats-io/go-nats-streaming"
)

const ServiceTypeNATSStreaming = "nats-streaming"

type NATSStreaming struct {
	Config

	healthy      bool
	connected    bool
	reconnecting bool

	eventCallbacks []EventCallback
	connection     stan.Conn
}

func (n *NATSStreaming) SetConfig(config Config) {
	n.Config = config
}

func (n *NATSStreaming) Connect() error {
	connectionString := fmt.Sprintf(
		"nats://%s:%s@%s:%d",
		n.Config.Username,
		n.Config.Password,
		n.Config.Host,
		n.Config.Port,
	)

	reconnectInterval := n.Config.ReconnectIntervalMilliseconds
	if reconnectInterval == 0 {
		reconnectInterval = 1000
	}

	conn, err := stan.Connect(
		n.Config.ClusterName,
		n.Config.ClientName,
		stan.NatsURL(connectionString),
		stan.Pings(1, 3),
		stan.SetConnectionLostHandler(n.onDisconnected),
	)

	if err != nil {
		n.dispatchEvent(Event{
			ServiceType: ServiceTypeNATSStreaming,
			Code:        ServiceCouldNotConnect,
		})

		return err
	}

	// cache the nats connection
	n.connection = conn

	// let everyone know we've connected
	n.connected = true
	n.dispatchEvent(Event{
		ServiceType: ServiceTypeNATSStreaming,
		Code:        ServiceConnected,
	})

	// let everyone know we're healthy
	n.healthy = true
	n.dispatchEvent(Event{
		ServiceType: ServiceTypeNATSStreaming,
		Code:        ServiceHealthy,
	})

	return nil
}

func (n *NATSStreaming) dispatchEvent(event Event) {
	for _, callback := range n.eventCallbacks {
		callback(event)
	}
}

func (n *NATSStreaming) onDisconnected(_ stan.Conn, reason error) {
	// let everyone know we're unhealthy
	n.healthy = false
	n.dispatchEvent(Event{
		ServiceType: ServiceTypeNATSStreaming,
		Code:        ServiceUnhealthy,
	})

	// let everyone know we've disconnected
	n.connected = false
	n.dispatchEvent(Event{
		ServiceType: ServiceTypeNATSStreaming,
		Code:        ServiceDisconnected,
	})

	if n.Config.ReconnectEnabled {
		go n.tryToReconnect()
	}
}

func (n *NATSStreaming) tryToReconnect() {
	if n.IsReconnecting() {
		return
	}

	// let everyone know we're reconnecting
	n.reconnecting = true
	n.dispatchEvent(Event{
		ServiceType: ServiceTypeNATSStreaming,
		Code:        ServiceReconnecting,
	})

	// try the reconnecting strategy
	callback := n.Config.ReconnectStrategy
	if callback == nil {
		callback = func(svc Service) bool {
			if err := n.Connect(); err != nil {
				return false
			}

			return true
		}
	}

	successful := callback(n)

	// if we weren't successful, attempt to reschedule things
	if !successful && n.Config.ReconnectEnabled {
		// calculate when to start the next reconnect
		interval := n.Config.ReconnectIntervalMilliseconds
		if interval == 0 {
			interval = 1000
		}

		time.Sleep(time.Millisecond * time.Duration(interval))
		n.reconnecting = false
		go n.tryToReconnect()
	} else {
		n.reconnecting = false
	}
}

func (n *NATSStreaming) Disconnect() error {
	if n.connection == nil {
		return nil
	}

	// close the connection
	n.connection.Close()

	// reset some variables
	n.connection = nil
	n.reconnecting = false

	return nil
}

func (n *NATSStreaming) GetClient() interface{} {
	return n.connection
}

func (n *NATSStreaming) Subscribe(callback EventCallback) {
	n.eventCallbacks = append(n.eventCallbacks, callback)
}

func (n *NATSStreaming) Unsubscribe(callback EventCallback) {
	callbacks := []EventCallback{}
	f1 := reflect.ValueOf(callback)
	p1 := f1.Pointer()

	for _, tmp := range n.eventCallbacks {
		f2 := reflect.ValueOf(tmp)
		p2 := f2.Pointer()

		if p1 == p2 {
			continue
		}

		callbacks = append(callbacks, callback)
	}

	n.eventCallbacks = callbacks
}

func (n *NATSStreaming) IsHealthy() bool {
	return n.healthy
}

func (n *NATSStreaming) IsConnected() bool {
	return n.connected
}

func (n *NATSStreaming) IsReconnecting() bool {
	return n.reconnecting
}
