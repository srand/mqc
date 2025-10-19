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
	options     *transport.TransportOptions
	mqttOptions *mqtt.ClientOptions
	mqttClient  mqtt.Client
	serializer  serialization.Serializer
	handlers    map[mqc.Method]mqc.MethodHandler
}

var _ mqc.Transport = (*pahoTransport)(nil)

func NewTransport(options ...transport.TransportOption) (mqc.Transport, error) {
	transportOptions := &transport.TransportOptions{
		ConnectTimeout: time.Second * 5,
		CallTimeout:    time.Second * 5,
	}

	for _, opt := range options {
		if err := opt(transportOptions); err != nil {
			return nil, err
		}
	}

	if len(transportOptions.Addrs) == 0 {
		return nil, mqc.ErrNoAddress
	}

	mqttOptions := mqtt.NewClientOptions()

	for _, addr := range transportOptions.Addrs {
		mqttOptions.AddBroker(addr)
	}

	if transportOptions.TlsConfig != nil {
		mqttOptions.SetTLSConfig(transportOptions.TlsConfig)
	}

	client := mqtt.NewClient(mqttOptions)
	serializer := serialization.NewJSONSerializer()

	return &pahoTransport{
		options:     transportOptions,
		mqttClient:  client,
		mqttOptions: mqttOptions,
		serializer:  serializer,
		handlers:    make(map[mqc.Method]mqc.MethodHandler),
	}, nil
}

func (p *pahoTransport) ensureConnected() error {
	if p.mqttClient.IsConnected() {
		return nil
	}

	if token := p.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	for method := range p.handlers {
		if err := p.subscribe(&method); err != nil {
			return err
		}
	}

	if p.options.OnConnect != nil {
		p.options.OnConnect(p)
	}

	return nil
}

func (p *pahoTransport) Close() error {
	p.mqttClient.Disconnect(0)
	return nil
}

func (p *pahoTransport) Invoke(ctx context.Context, method *mqc.Method) (mqc.Conn, error) {
	if err := p.ensureConnected(); err != nil {
		return nil, err
	}

	if !p.mqttClient.IsConnected() {
		return nil, fmt.Errorf("mqtt client is not connected")
	}

	if method.IsPubSub() {
		return newPubSubConn(p.serializer, p.mqttClient, method)
	}

	conn, err := newConn(p.serializer, p.mqttClient, method, uuid.New().String(), false)
	if err != nil {
		return nil, err
	}

	err = conn.Invoke(ctx)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func (p *pahoTransport) RegisterHandler(method *mqc.Method, handler mqc.MethodHandler) error {
	p.handlers[*method] = handler
	return nil
}

func (p *pahoTransport) UnregisterHandler(method *mqc.Method) error {
	delete(p.handlers, *method)
	return nil
}

func (p *pahoTransport) Dial() error {
	if p.mqttClient.IsConnected() {
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

func (p *pahoTransport) subscribe(method *mqc.Method) error {
	topic := sharedControlTopic(method, "+")

	token := p.mqttClient.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		var m mqc.Message

		if err := p.serializer.Unmarshal(msg.Payload(), &m); err != nil {
			return
		}

		if !m.IsCall() {
			return
		}

		handler, ok := p.handlers[*m.Method()]
		if !ok {
			return
		}

		conn, err := newConn(p.serializer, p.mqttClient, method, extractTopicId(msg.Topic()), true)
		if err != nil {
			return
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Recovered in server goroutine:", r)
				}
			}()

			defer conn.Close()

			ctx, cancel := context.WithTimeout(context.Background(), p.options.CallTimeout)
			defer cancel()

			// Ack the received message
			if err := conn.SendAck(ctx); err != nil {
				return
			}

			ctx, cancel = context.WithTimeout(context.Background(), p.options.CallTimeout)
			defer cancel()

			if err := handler(conn); err != nil {
				conn.SendError(ctx, err)
			} else {
				conn.SendClose(ctx)
			}
		}()
	})
	if token == nil {
		return errors.New("failed to create subscription token")
	}
	token.Wait()
	return token.Error()
}
