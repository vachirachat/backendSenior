module worker

go 1.15

replace common => ../common

replace backendSenior => ../backend

replace proxySenior => ../proxy-server

require (
	backendSenior v0.0.0-00010101000000-000000000000
	common v0.0.0-00010101000000-000000000000
	github.com/disintegration/imaging v1.6.2
	github.com/go-resty/resty/v2 v2.5.0
	github.com/streadway/amqp v1.0.0
	proxySenior v0.0.0-00010101000000-000000000000
)
