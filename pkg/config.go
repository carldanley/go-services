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
