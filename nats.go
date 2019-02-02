package services

import (
	"fmt"
	"reflect"
	"time"

	nats "github.com/nats-io/go-nats"
)

const ServiceTypeNATS = "nats"

type NATS struct {
	Config

	healthy         bool
	connected       bool
	reconnecting    bool
	shouldReconnect bool

	eventCallbacks []EventCallback
	connection     *nats.Conn
}

func (n *NATS) SetConfig(config Config) {
	n.Config = config
	n.shouldReconnect = config.ReconnectEnabled
}

func (n *NATS) Connect() error {
	connectionString := fmt.Sprintf(
		"nats://%s:%s@%s:%d",
		n.Config.Username,
		n.Config.Password,
		n.Config.Host,
		n.Config.Port,
	)

	conn, err := nats.Connect(
		connectionString,
		nats.DisconnectHandler(n.onDisconnected),
		nats.MaxReconnects(0),
	)

	if err != nil {
		n.dispatchEvent(Event{
			ServiceType: ServiceTypeNATS,
			Code:        ServiceCouldNotConnect,
		})

		return err
	}

	// cache the nats connection
	n.connection = conn

	// reset the value of `shouldReconnect` back to what the configuration states
	n.shouldReconnect = n.Config.ReconnectEnabled

	// let everyone know we've connected
	n.connected = true
	n.dispatchEvent(Event{
		ServiceType: ServiceTypeNATS,
		Code:        ServiceConnected,
	})

	// let everyone know we're healthy
	n.healthy = true
	n.dispatchEvent(Event{
		ServiceType: ServiceTypeNATS,
		Code:        ServiceHealthy,
	})

	return nil
}

func (n *NATS) dispatchEvent(event Event) {
	for _, callback := range n.eventCallbacks {
		callback(event)
	}
}

func (n *NATS) onDisconnected(nc *nats.Conn) {
	// let everyone know we're unhealthy
	n.healthy = false
	n.dispatchEvent(Event{
		ServiceType: ServiceTypeNATS,
		Code:        ServiceUnhealthy,
	})

	// let everyone know we've disconnected
	n.connected = false
	n.dispatchEvent(Event{
		ServiceType: ServiceTypeNATS,
		Code:        ServiceDisconnected,
	})

	go n.tryToReconnect()
}

func (n *NATS) tryToReconnect() {
	if !n.shouldReconnect || n.IsReconnecting() {
		return
	}

	// let everyone know we're reconnecting
	n.reconnecting = true
	n.dispatchEvent(Event{
		ServiceType: ServiceTypeNATS,
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

func (n *NATS) Disconnect() error {
	// stop any reconnections from happening
	n.shouldReconnect = false

	// make sure we had an active connection
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

func (n *NATS) GetClient() interface{} {
	return n.connection
}

func (n *NATS) Subscribe(callback EventCallback) {
	n.eventCallbacks = append(n.eventCallbacks, callback)
}

func (n *NATS) Unsubscribe(callback EventCallback) {
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

func (n *NATS) IsHealthy() bool {
	return n.healthy
}

func (n *NATS) IsConnected() bool {
	return n.connected
}

func (n *NATS) IsReconnecting() bool {
	return n.reconnecting
}
