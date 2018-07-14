package services

import (
	"fmt"
	"reflect"
	"time"

	"github.com/garyburd/redigo/redis"
)

const ServiceTypeRedis = "redis"

type Redis struct {
	Config

	healthy         bool
	connected       bool
	reconnecting    bool
	shouldReconnect bool

	eventCallbacks []EventCallback
	connection     redis.Conn
}

func (r *Redis) SetConfig(config Config) {
	r.Config = config
	r.shouldReconnect = config.ReconnectEnabled
}

func (r *Redis) Connect() error {
	connection, err := redis.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", r.Config.Host, r.Config.Port),
		redis.DialPassword(r.Config.Password),
	)

	if err != nil {
		r.dispatchEvent(Event{
			ServiceType: ServiceTypeRedis,
			Code:        ServiceCouldNotConnect,
		})

		return err
	}

	// cache the connection
	r.connection = connection

	// reset the value of `shouldReconnect` back to what the configuration states
	r.shouldReconnect = r.Config.ReconnectEnabled

	// let everyone know we've connected
	r.connected = true
	r.dispatchEvent(Event{
		ServiceType: ServiceTypeRedis,
		Code:        ServiceConnected,
	})

	// let everyone know we're healthy
	r.healthy = true
	r.dispatchEvent(Event{
		ServiceType: ServiceTypeRedis,
		Code:        ServiceHealthy,
	})

	// begin monitoring the database connection
	go r.monitorConnection()

	return nil
}

func (r *Redis) Disconnect() error {
	// stop any reconnections from happening
	r.shouldReconnect = false

	// make sure we had an active connection
	if r.connection == nil {
		return nil
	}

	// close the connection
	err := r.connection.Close()

	// reset some variables
	r.connection = nil
	r.reconnecting = false

	// let everyone know we're unhealthy
	r.healthy = false
	r.dispatchEvent(Event{
		ServiceType: ServiceTypeRedis,
		Code:        ServiceUnhealthy,
	})

	// let everyone know we've disconnected
	r.connected = false
	r.dispatchEvent(Event{
		ServiceType: ServiceTypeRedis,
		Code:        ServiceDisconnected,
	})

	return err
}

func (r *Redis) GetClient() interface{} {
	return r.connection
}

func (r *Redis) Subscribe(callback EventCallback) {
	r.eventCallbacks = append(r.eventCallbacks, callback)
}

func (r *Redis) Unsubscribe(callback EventCallback) {
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

func (r *Redis) IsHealthy() bool {
	return r.healthy
}

func (r *Redis) IsConnected() bool {
	return r.connected
}

func (r *Redis) IsReconnecting() bool {
	return r.reconnecting
}

func (r *Redis) dispatchEvent(event Event) {
	for _, callback := range r.eventCallbacks {
		callback(event)
	}
}

func (r *Redis) monitorConnection() {
	if r.connection == nil {
		return
	}

	if _, err := r.connection.Do("PING"); err != nil {
		// first disconnect
		r.Disconnect()

		// begin trying to reconnect
		go r.tryToReconnect()
	} else {
		interval := r.Config.MonitorIntervalMilliseconds
		if interval == 0 {
			interval = 1000
		}

		time.Sleep(time.Millisecond * time.Duration(interval))

		go r.monitorConnection()
	}
}

func (r *Redis) tryToReconnect() {
	if !r.shouldReconnect || r.IsReconnecting() {
		return
	}

	// let everyone know we're reconnecting
	r.reconnecting = true
	r.dispatchEvent(Event{
		ServiceType: ServiceTypeRedis,
		Code:        ServiceReconnecting,
	})

	// try the reconnecting strategy
	callback := r.Config.ReconnectStrategy
	if callback == nil {
		callback = func(svc Service) bool {
			if err := r.Connect(); err != nil {
				return false
			}

			if _, err := r.GetClient().(redis.Conn).Do("PING"); err != nil {
				return false
			}

			return true
		}
	}

	successful := callback(r)

	// if we weren't successful, attempt to reschedule things
	if !successful && r.Config.ReconnectEnabled {
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
