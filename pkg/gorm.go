package pkg

import (
	"reflect"
)

const ServiceTypeGorm = "gorm"

type Gorm struct {
	Config

	healthy      bool
	connected    bool
	reconnecting bool

	serviceEvents  EventStream
	eventCallbacks []EventCallback
}

func (g *Gorm) SetConfig(config Config) {
	g.Config = config
}

func (g *Gorm) Connect() error {
	return nil
}

func (g *Gorm) Disconnect() error {
	return nil
}

func (g *Gorm) GetClient() interface{} {
	return nil
}

func (g *Gorm) Subscribe(callback EventCallback) {
	g.eventCallbacks = append(g.eventCallbacks, callback)
}

func (g *Gorm) Unsubscribe(callback EventCallback) {
	callbacks := []EventCallback{}
	f1 := reflect.ValueOf(callback)
	p1 := f1.Pointer()

	for _, tmp := range g.eventCallbacks {
		f2 := reflect.ValueOf(tmp)
		p2 := f2.Pointer()

		if p1 == p2 {
			continue
		}

		callbacks = append(callbacks, callback)
	}

	g.eventCallbacks = callbacks
}

func (g *Gorm) IsHealthy() bool {
	return g.healthy
}

func (g *Gorm) IsConnected() bool {
	return g.connected
}

func (g *Gorm) IsReconnecting() bool {
	return g.reconnecting
}
