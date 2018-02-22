package main

import (
	"fmt"

	"gitlab.encrypted.place/open-source/services/pkg"
)

func main() {
	factory := pkg.NewFactory()
	config := pkg.Config{
		Host: "127.0.0.1",
		Port: 6379,

		ReconnectEnabled: true,
	}

	if err := factory.Register(pkg.ServiceTypeRedis, config); err != nil {
		panic(err)
	}

	config = pkg.Config{
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "root",
		Password: "root",
		Database: "users",

		ReconnectEnabled: true,
	}

	if err := factory.Register(pkg.ServiceTypeGorm, config); err != nil {
		panic(err)
	}

	config = pkg.Config{
		Host:     "127.0.0.1",
		Port:     5672,
		Username: "guest",
		Password: "guest",

		ReconnectEnabled: true,
	}

	if err := factory.Register(pkg.ServiceTypeRabbitMQ, config); err != nil {
		panic(err)
	}

	factory.Subscribe(func(event pkg.Event) {
		var status string

		switch event.Code {
		case pkg.ServiceUnhealthy:
			status = "unhealthy"
		case pkg.ServiceHealthy:
			status = "healthy"
		case pkg.ServiceConnected:
			status = "connected"
		case pkg.ServiceDisconnected:
			status = "disconnected"
		case pkg.ServiceReconnecting:
			status = "reconnecting"
		case pkg.ServiceReconnected:
			status = "reconnected"
		case pkg.ServiceCouldNotConnect:
			status = "could not connect"
		}

		fmt.Printf("[%s]: %s\n", event.ServiceType, status)
	})

	if err := factory.Connect(); err != nil {
		panic(err)
	}

	defer factory.Disconnect()

	// todo: remove this function
	<-factory.GetEventStream()
}
