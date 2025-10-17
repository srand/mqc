package mqtt

import (
	"context"
	"errors"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/srand/mqc"
	"github.com/srand/mqc/serialization"
	"github.com/srand/mqc/transport"
)

type pahoTransport struct {
	transport.TransportEvents

	dialOptions *transport.DialOptions
	client      mqtt.Client
	options     *mqtt.ClientOptions
	serializer  serialization.Serializer
	handlers    map[mqc.Method]mqc.MethodHandler
}

var _ mqc.Transport = (*pahoTransport)(nil)

func NewTransport(options ...transport.DialOption) (mqc.Transport, error) {
	dialOptions := &transport.DialOptions{
		ConnectTimeout: time.Second * 5,
		CallTimeout:    time.Second * 5,
	}

	for _, opt := range options {
		if err := opt(dialOptions); err != nil {
			return nil, err
		}
	}

	if len(dialOptions.Addrs) == 0 {
		return nil, mqc.ErrNoAddress
	}

	mqttOptions := mqtt.NewClientOptions()

	for _, addr := range dialOptions.Addrs {
		mqttOptions.AddBroker(addr)
	}

	if dialOptions.TlsConfig != nil {
		mqttOptions.SetTLSConfig(dialOptions.TlsConfig)
	}

	client := mqtt.NewClient(mqttOptions)
	serializer := serialization.NewJSONSerializer()

	return &pahoTransport{
		dialOptions: dialOptions,
		client:      client,
		options:     mqttOptions,
		serializer:  serializer,
		handlers:    make(map[mqc.Method]mqc.MethodHandler),
	}, nil
}

func (p *pahoTransport) ensureConnected() error {
	if p.client.IsConnected() {
		return nil
	}

	if token := p.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	for method := range p.handlers {
		if err := p.subscribe(method); err != nil {
			return err
		}
	}

	p.TriggerConnect(p)

	return nil
}

func (p *pahoTransport) Close() error {
	p.client.Disconnect(0)
	return nil
}

func (p *pahoTransport) Invoke(ctx context.Context, method mqc.Method) (mqc.Conn, error) {
	if err := p.ensureConnected(); err != nil {
		return nil, err
	}

	call, err := newCallConn(p.serializer, p.client, method, uuid.New().String(), false)
	if err != nil {
		return nil, err
	}

	err = call.Invoke(ctx)
	if err != nil {
		call.Close()
		return nil, err
	}

	return call, nil
}

func (p *pahoTransport) RegisterHandler(method mqc.Method, handler mqc.MethodHandler) error {
	p.handlers[method] = handler
	return nil
}

func (p *pahoTransport) UnregisterHandler(method mqc.Method) error {
	delete(p.handlers, method)
	return nil
}

func (p *pahoTransport) Dial() error {
	if p.client.IsConnected() {
		return fmt.Errorf("transport is already connected")
	}
	return p.ensureConnected()
}

func (p *pahoTransport) Serve() error {
	if err := p.ensureConnected(); err != nil {
		return err
	}

	select {}
}

func (p *pahoTransport) Serializer() serialization.Serializer {
	return p.serializer
}

func (p *pahoTransport) subscribe(method mqc.Method) error {
	topic := sharedControlTopic(method, "+")

	token := p.client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		var m mqc.Message

		if err := p.serializer.Unmarshal(msg.Payload(), &m); err != nil {
			return
		}

		if !m.IsCall() {
			return
		}

		handler, ok := p.handlers[m.Method()]
		if !ok {
			return
		}

		call, err := newCallConn(p.serializer, p.client, method, extractTopicId(msg.Topic()), true)
		if err != nil {
			return
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Recovered in server goroutine:", r)
				}
			}()

			defer call.Close()

			ctx, cancel := context.WithTimeout(context.Background(), p.dialOptions.CallTimeout)
			defer cancel()

			// Ack the received message
			if err := call.SendAck(ctx); err != nil {
				return
			}

			ctx, cancel = context.WithTimeout(context.Background(), p.dialOptions.CallTimeout)
			defer cancel()

			err := handler(call)
			if err != nil {
				call.SendError(ctx, err)
			} else {
				call.SendClose(ctx)
			}
		}()
	})
	if token == nil {
		return errors.New("failed to create subscription token")
	}
	token.Wait()
	return token.Error()
}
