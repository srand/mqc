module loadbalancer

go 1.24.6

require (
	github.com/srand/mqc v0.0.0-20251016101848-7cc4fef3a4b3
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/eclipse/paho.mqtt.golang v1.5.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
)

replace github.com/srand/mqc => ../..
