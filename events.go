package services

type Event struct {
	ServiceType string
	Code        int
	Error       error
}

type EventCallback func(Event)

type EventStream chan Event

const (
	ServiceUnhealthy       = 1
	ServiceHealthy         = 2
	ServiceConnected       = 3
	ServiceDisconnected    = 4
	ServiceReconnecting    = 5
	ServiceReconnected     = 6
	ServiceCouldNotConnect = 7
)
