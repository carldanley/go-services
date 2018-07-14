# services

## About

`services` is a package that helps unify the connection management interface for common services used within golang web applications. The idea was to monitor and consume these connections in a way that makes it easy to deduce the overall health of an application during runtime (whether it's currently maintaining active connections to it's dependency services).

## Currently Supported Services

1. Redis
1. RabbitMQ
1. Gorm (MySQL)
1. NATS
1. NATS-streaming
