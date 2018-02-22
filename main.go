package main

import (
	"gitlab.encrypted.place/open-source/services/pkg"
)

func main() {
	factory := pkg.NewFactory()
	config := pkg.Config{
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "root",
		Password: "root",

		ReconnectEnabled:              true,
		ReconnectIntervalMilliseconds: 5000,
	}

	if err := factory.Register(pkg.ServiceTypeGorm, config); err != nil {
		panic(err)
	}

	if err := factory.Connect(); err != nil {
		panic(err)
	}

	defer factory.Disconnect()

	// todo: remove this function
	<-factory.GetEventStream()
}
