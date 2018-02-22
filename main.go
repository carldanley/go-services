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
		// Username: "guest",
		// Password: "guest",
		// Database: "users",

		ReconnectEnabled:              true,
		ReconnectIntervalMilliseconds: 5000,
	}

	if err := factory.Register(pkg.ServiceTypeRedis, config); err != nil {
		panic(err)
	}

	factory.Subscribe(func(event pkg.Event) {
		switch event.Code {
		case pkg.ServiceUnhealthy:
			fmt.Println("unhealthy")
		case pkg.ServiceHealthy:
			fmt.Println("healthy")
		case pkg.ServiceConnected:
			fmt.Println("connected")
		case pkg.ServiceDisconnected:
			fmt.Println("disconnected")
		case pkg.ServiceReconnecting:
			fmt.Println("reconnecting")
		case pkg.ServiceReconnected:
			fmt.Println("reconnected")
		case pkg.ServiceCouldNotConnect:
			fmt.Println("could not connect")
		}
	})

	if err := factory.Connect(); err != nil {
		panic(err)
	}

	defer factory.Disconnect()

	// todo: remove this function
	<-factory.GetEventStream()
}
