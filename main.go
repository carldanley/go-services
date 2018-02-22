package main

import (
	"time"

	"github.com/davecgh/go-spew/spew"
	"gitlab.encrypted.place/open-source/services/pkg"
)

func main() {
	factory := pkg.NewFactory()
	config := pkg.Config{
		Host:     "127.0.0.1",
		Port:     5672,
		Username: "guest",
		Password: "guest",

		ReconnectEnabled:              true,
		ReconnectIntervalMilliseconds: 5000,
		ReconnectStrategy: func(svc pkg.Service) bool {
			if err := svc.Connect(); err != nil {
				return false
			}

			return true
		},
	}

	if err := factory.Register(pkg.ServiceTypeRabbitMQ, config); err != nil {
		panic(err)
	}

	myCallback := func(event pkg.Event) {
		spew.Dump(event)
	}

	factory.Subscribe(myCallback)

	if err := factory.Connect(); err != nil {
		panic(err)
	}

	go func() {
		time.Sleep(time.Second * 5)

		_, err := factory.Get(pkg.ServiceTypeRabbitMQ)
		if err != nil {
			panic(err)
		}

		factory.Unsubscribe(myCallback)
	}()

	defer factory.Disconnect()

	// todo: remove this function
	<-factory.GetEventStream()

}
