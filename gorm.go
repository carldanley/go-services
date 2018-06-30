package services

import (
	"fmt"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const ServiceTypeGorm = "gorm"

type Gorm struct {
	Config

	healthy      bool
	connected    bool
	reconnecting bool

	eventCallbacks []EventCallback

	db *gorm.DB
}

func (g *Gorm) SetConfig(config Config) {
	g.Config = config
}

func (g *Gorm) Connect() error {
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		g.Config.Username,
		g.Config.Password,
		g.Config.Host,
		g.Config.Port,
		g.Config.Database,
	)

	db, err := gorm.Open("mysql", connectionString)
	if err != nil {
		g.dispatchEvent(Event{
			ServiceType: ServiceTypeGorm,
			Code:        ServiceCouldNotConnect,
		})

		return err
	}

	// cache the gorm database
	g.db = db

	// let everyone know we've connected
	g.connected = true
	g.dispatchEvent(Event{
		ServiceType: ServiceTypeGorm,
		Code:        ServiceConnected,
	})

	// let everyone know we're healthy
	g.healthy = true
	g.dispatchEvent(Event{
		ServiceType: ServiceTypeGorm,
		Code:        ServiceHealthy,
	})

	// begin monitoring the database connection
	go g.monitorConnection()

	return nil
}

func (g *Gorm) Disconnect() error {
	if g.db == nil {
		return nil
	}

	// close the connection
	err := g.db.Close()

	// reset some variables
	g.db = nil
	g.reconnecting = false

	// let everyone know we're unhealthy
	g.healthy = false
	g.dispatchEvent(Event{
		ServiceType: ServiceTypeGorm,
		Code:        ServiceUnhealthy,
	})

	// let everyone know we've disconnected
	g.connected = false
	g.dispatchEvent(Event{
		ServiceType: ServiceTypeGorm,
		Code:        ServiceDisconnected,
	})

	return err
}

func (g *Gorm) GetClient() interface{} {
	return g.db
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

func (g *Gorm) dispatchEvent(event Event) {
	for _, callback := range g.eventCallbacks {
		callback(event)
	}
}

func (g *Gorm) monitorConnection() {
	if g.db == nil {
		return
	}

	interval := g.Config.MonitorIntervalMilliseconds
	if interval == 0 {
		interval = 1000
	}

	time.Sleep(time.Millisecond * time.Duration(interval))

	if _, err := g.db.DB().Exec("DO 1;"); err != nil {
		// first, disconnect
		g.Disconnect()

		// begin trying to reconnect
		go g.tryToReconnect()
	} else {
		go g.monitorConnection()
	}
}

func (g *Gorm) tryToReconnect() {
	if g.IsReconnecting() {
		return
	}

	// let everyone know we're reconnecting
	g.reconnecting = true
	g.dispatchEvent(Event{
		ServiceType: ServiceTypeGorm,
		Code:        ServiceReconnecting,
	})

	// try the reconnecting strategy
	callback := g.Config.ReconnectStrategy
	if callback == nil {
		callback = func(svc Service) bool {
			if err := g.Connect(); err != nil {
				return false
			}

			if _, err := g.GetClient().(*gorm.DB).DB().Exec("DO 1;"); err != nil {
				return false
			}

			return true
		}
	}

	successful := callback(g)

	// if we weren't successful, attempt to reschedule things
	if !successful && g.Config.ReconnectEnabled {
		// calculate when to start the next reconnect
		interval := g.Config.ReconnectIntervalMilliseconds
		if interval == 0 {
			interval = 1000
		}

		time.Sleep(time.Millisecond * time.Duration(interval))
		g.reconnecting = false
		go g.tryToReconnect()
	} else {
		g.reconnecting = false
	}
}
