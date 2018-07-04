package services

type Config struct {
	Host        string
	Port        uint32
	Username    string
	Password    string
	Database    string
	ClusterName string
	ClientName  string

	MonitorIntervalMilliseconds int

	ReconnectEnabled              bool
	ReconnectIntervalMilliseconds int
	ReconnectStrategy             ReconnectStrategy
}

type ReconnectStrategy func(svc Service) (successful bool)
