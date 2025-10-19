package mqtt

import (
	"context"
	"errors"
	"io"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/srand/mqc"
	"github.com/srand/mqc/serialization"
)

type pubsubConn struct {
	client     mqtt.Client
	method     mqc.Method
	receiver   chan *mqc.Message
	topic      string
	serializer serialization.Serializer
	err        error
}

var _ mqc.Conn = (*pubsubConn)(nil)

func pubsubTopic(method *mqc.Method) string {
	return "MQC/" + method.Name
}

func newPubSubConn(serializer serialization.Serializer, client mqtt.Client, method *mqc.Method) (*pubsubConn, error) {
	receiver := make(chan *mqc.Message, 1)
	pc := &pubsubConn{
		client:     client,
		method:     *method,
		receiver:   receiver,
		topic:      pubsubTopic(method),
		serializer: serializer,
	}
	if method.IsConsumer() {
		if err := pc.subscribe(pc.topic, false); err != nil {
			return nil, err
		}
	}
	return pc, nil
}

func (c *pubsubConn) Recv(ctx context.Context) ([]byte, error) {
	if c.err != nil {
		return nil, c.err
	}

	var msg *mqc.Message

	select {
	case msg = <-c.receiver:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if msg == nil || msg.IsClose() {
		return nil, io.EOF
	}

	if msg.IsError() {
		c.err = msg.Error()
		return nil, c.err
	}

	return msg.DataBytes(), nil
}

func (c *pubsubConn) Send(ctx context.Context, data []byte) error {
	if data == nil {
		return errors.New("data is nil")
	}

	if c.err != nil {
		return c.err
	}

	return c.publish(ctx, c.topic, data)
}

func (c *pubsubConn) SendClose(ctx context.Context) error {
	return errors.ErrUnsupported
}

func (c *pubsubConn) SendMethod(ctx context.Context, method mqc.Method) error {
	return errors.ErrUnsupported
}

func (c *pubsubConn) publish(ctx context.Context, topic string, data []byte) error {
	token := c.client.Publish(topic, 0, false, data)

	done := make(chan error)
	go func() {
		token.Wait()
		if err := token.Error(); err != nil {
			done <- err
		}
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (c *pubsubConn) subscribe(topic string, data bool) error {
	token := c.client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		m := mqc.Message{
			Type: mqc.Message_DATA,
			Data: msg.Payload(),
		}

		if m.IsError() {
			c.err = m.Error()
		}

		c.receiver <- &m
	})
	if token == nil {
		return errors.New("failed to create subscription token")
	}
	token.Wait()
	return token.Error()
}

func (c *pubsubConn) unsubscribe(topic string) error {
	token := c.client.Unsubscribe(topic)
	token.Wait()
	return token.Error()
}

func (c *pubsubConn) Close() error {
	return c.unsubscribe(c.topic)
}
