package pkg

import (
	"reflect"
)

const ServiceTypeRedis = "redis"

type Redis struct {
	Config

	healthy      bool
	connected    bool
	reconnecting bool

	serviceEvents  EventStream
	eventCallbacks []EventCallback
}

func (r *Redis) SetConfig(config Config) {
	r.Config = config
}

func (r *Redis) Connect() error {
	return nil
}

func (r *Redis) Disconnect() error {
	return nil
}

func (r *Redis) GetClient() interface{} {
	return nil
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
