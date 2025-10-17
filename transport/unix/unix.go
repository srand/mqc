package unix

import (
	"github.com/srand/mqc"
	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/tcp"
)

func NewTransport(options ...transport.TransportOption) (mqc.Transport, error) {
	transportOptions := []transport.TransportOption{}
	transportOptions = append(transportOptions, options...)
	transportOptions = append(transportOptions, transport.WithProtocol("unix"))
	return tcp.NewTransport(transportOptions...)
}
