package services

type Service interface {
	SetConfig(Config)

	Connect() error
	Disconnect() error

	GetClient() interface{}

	Subscribe(EventCallback)
	Unsubscribe(EventCallback)

	IsHealthy() bool
	IsConnected() bool
	IsReconnecting() bool
}
