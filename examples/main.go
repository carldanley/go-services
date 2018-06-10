package main

import (
	"fmt"

	"github.com/carldanley/go-services"
)

func registerRedis(factory *services.Factory) {
	config := services.Config{
		Host: "127.0.0.1",
		Port: 6379,

		ReconnectEnabled: true,
	}

	if err := factory.Register(services.ServiceTypeRedis, config); err != nil {
		panic(err)
	}
}

func registerRabbitMQ(factory *services.Factory) {
	config := services.Config{
		Host:     "127.0.0.1",
		Port:     5672,
		Username: "guest",
		Password: "guest",

		ReconnectEnabled: true,
	}

	if err := factory.Register(services.ServiceTypeRabbitMQ, config); err != nil {
		panic(err)
	}
}

func registerGorm(factory *services.Factory) {
	config := services.Config{
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "root",
		Password: "root",
		Database: "users",

		ReconnectEnabled: true,
	}

	if err := factory.Register(services.ServiceTypeGorm, config); err != nil {
		panic(err)
	}
}

func showEvents(event services.Event) {
	var status string

	switch event.Code {
	case services.ServiceUnhealthy:
		status = "unhealthy"
	case services.ServiceHealthy:
		status = "healthy"
	case services.ServiceConnected:
		status = "connected"
	case services.ServiceDisconnected:
		status = "disconnected"
	case services.ServiceReconnecting:
		status = "reconnecting"
	case services.ServiceReconnected:
		status = "reconnected"
	case services.ServiceCouldNotConnect:
		status = "could not connect"
	}

	fmt.Printf("[%s]: %s\n", event.ServiceType, status)
}

func main() {
	factory := services.NewFactory()

	registerGorm(factory)
	registerRabbitMQ(factory)
	registerRedis(factory)

	factory.Subscribe(showEvents)

	if err := factory.Connect(); err != nil {
		panic(err)
	}

	defer factory.Disconnect()

	// todo: remove this function
	<-factory.GetEventStream()
}
