module websocket

go 1.24.6

replace github.com/srand/mqc => ../..

require (
	github.com/srand/mqc v0.0.0-20251016101848-7cc4fef3a4b3
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/hashicorp/yamux v0.1.2 // indirect
	golang.org/x/net v0.44.0 // indirect
)
