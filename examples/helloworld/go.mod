module helloworld

go 1.24.6

replace github.com/srand/mqc => ../..

require (
	github.com/srand/mqc v0.0.0-20251016101848-7cc4fef3a4b3
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
)
