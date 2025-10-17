package transport

import "github.com/srand/mqc"

type TransportEvents struct {
	onConnect []func(mqc.Transport)
}

func (te *TransportEvents) OnConnect(f func(mqc.Transport)) {
	te.onConnect = append(te.onConnect, f)
}

func (te *TransportEvents) TriggerConnect(transport mqc.Transport) {
	for _, f := range te.onConnect {
		f(transport)
	}
}
