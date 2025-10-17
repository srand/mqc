module client_service

go 1.24.6

require (
	github.com/srand/mqc v0.0.0-20251016203617-2db9edc0c605
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
)

replace github.com/srand/mqc => ../..
