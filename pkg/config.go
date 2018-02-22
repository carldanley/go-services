package pkg

type Config struct {
	Host     string
	Port     uint32
	Username string
	Password string

	ReconnectEnabled              bool
	ReconnectIntervalMilliseconds int
	ReconnectStrategy             ReconnectStrategy
}

type ReconnectStrategy func(svc Service) (successful bool)

func DefaultReconnectionStrategy(svc Service) (successful bool) {
	if err := svc.Connect(); err != nil {
		return false
	}

	return true
}
