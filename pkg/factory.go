package pkg

import (
	"errors"
	"fmt"
	"reflect"
)

type Factory struct {
	events             EventStream
	registeredServices map[string]Service
	eventCallbacks     []EventCallback
}

func NewFactory() *Factory {
	return &Factory{
		events:             make(EventStream),
		registeredServices: map[string]Service{},
		eventCallbacks:     []EventCallback{},
	}
}

func (f *Factory) Register(svcType string, config Config) error {
	var service Service

	switch svcType {
	case ServiceTypeRabbitMQ:
		service = &RabbitMQ{}
	default:
		return errors.New("Unrecognized service")
	}

	service.SetConfig(config)
	service.Subscribe(f.listenToServiceEvents)
	f.registeredServices[svcType] = service

	return nil
}

func (f *Factory) Unregister(svcType string) error {
	service, ok := f.registeredServices[svcType]
	if !ok {
		return errors.New("Service not registered")
	}

	service.Unsubscribe(f.listenToServiceEvents)
	defer service.Disconnect()
	delete(f.registeredServices, svcType)
	return nil
}

func (f *Factory) Subscribe(callback EventCallback) {
	f.eventCallbacks = append(f.eventCallbacks, callback)
}

func (f *Factory) Unsubscribe(callback EventCallback) {
	callbacks := []EventCallback{}
	f1 := reflect.ValueOf(callback)
	p1 := f1.Pointer()

	for _, tmp := range f.eventCallbacks {
		f2 := reflect.ValueOf(tmp)
		p2 := f2.Pointer()

		if p1 == p2 {
			continue
		}

		callbacks = append(callbacks, callback)
	}

	f.eventCallbacks = callbacks
}

func (f *Factory) Connect() error {
	for _, service := range f.registeredServices {
		if err := service.Connect(); err != nil {
			return err
		}
	}

	return nil
}

func (f *Factory) Disconnect() error {
	for _, service := range f.registeredServices {
		if err := service.Disconnect(); err != nil {
			return err
		}
	}

	return nil
}

func (f *Factory) Get(svcType string) (Service, error) {
	service, ok := f.registeredServices[svcType]
	if !ok {
		return nil, errors.New("Service not registered")
	}

	return service, nil
}

func (f *Factory) GetEventStream() EventStream {
	return f.events
}

func (f *Factory) listenToServiceEvents(event Event) {
	switch event.Code {
	case ServiceUnhealthy:
		fmt.Println("unhealthy")
	case ServiceHealthy:
		fmt.Println("healthy")
	case ServiceConnected:
		fmt.Println("connected")
	case ServiceDisconnected:
		fmt.Println("disconnected")
	case ServiceReconnecting:
		fmt.Println("reconnecting")
	case ServiceReconnected:
		fmt.Println("reconnected")
	case ServiceCouldNotConnect:
		fmt.Println("could not connect")
	}
}
