package test

import (
	"github.com/srand/mqc"
	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/mqtt"
	tpc "github.com/srand/mqc/transport/tcp"
)

type Factory interface {
	New() (mqc.Transport, error)
}

type TcpFactory struct {
}

func (f *TcpFactory) New() (mqc.Transport, error) {
	return tpc.NewTransport(transport.WithAddress("localhost:8080"))
}

var _ Factory = (*TcpFactory)(nil)

type UnixFactory struct {
}

func (f *UnixFactory) New() (mqc.Transport, error) {
	return tpc.NewTransport(transport.WithProtocol("unix"), transport.WithAddress("/tmp/mqc.sock"))
}

var _ Factory = (*UnixFactory)(nil)

// MQTT transport factory is not implemented yet
type MqttFactory struct {
}

func (f *MqttFactory) New() (mqc.Transport, error) {
	return mqtt.NewTransport(transport.WithAddress("localhost:1883"))
}

var _ Factory = (*MqttFactory)(nil)
