package mqtt

import (
	"context"
	"errors"
	"strings"

	"github.com/srand/mqc"
	"github.com/srand/mqc/serialization"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type call struct {
	client             mqtt.Client
	method             mqc.Method
	receiver           chan *mqc.Message
	id                 string
	clientControlTopic string
	clientDataTopic    string
	serverControlTopic string
	serverDataTopic    string
	controlTopic       string
	server             bool
	serializer         serialization.Serializer
}

func controlTopic(method mqc.Method, id string) string {
	return "MQC/" + method.String() + "/Control/" + id
}

func sharedControlTopic(method mqc.Method, id string) string {
	return "$share/MQC/MQC/" + method.String() + "/Control/" + id
}

func clientTopic(method mqc.Method, id string, name string) string {
	return "MQC/" + method.String() + "/Client/" + id + "/" + name
}

func serverTopic(method mqc.Method, id string, name string) string {
	return "MQC/" + method.String() + "/Server/" + id + "/" + name
}

func extractTopicId(topic string) string {
	parts := strings.Split(topic, "/")
	if len(parts) < 1 {
		return ""
	}
	return parts[len(parts)-1]
}

func newCall(serializer serialization.Serializer, client mqtt.Client, method mqc.Method, id string, server bool) (*call, error) {
	receiver := make(chan *mqc.Message, 1)
	call := &call{
		client:             client,
		receiver:           receiver,
		method:             method,
		id:                 id,
		clientControlTopic: clientTopic(method, id, "Control"),
		clientDataTopic:    clientTopic(method, id, "Data"),
		serverControlTopic: serverTopic(method, id, "Control"),
		serverDataTopic:    serverTopic(method, id, "Data"),
		controlTopic:       controlTopic(method, id),
		server:             server,
		serializer:         serializer,
	}

	if server {
		if err := call.subscribe(call.clientControlTopic, false); err != nil {
			return nil, err
		}
		if err := call.subscribe(call.clientDataTopic, true); err != nil {
			return nil, err
		}
	} else {
		if err := call.subscribe(call.serverControlTopic, false); err != nil {
			return nil, err
		}
		if err := call.subscribe(call.serverDataTopic, true); err != nil {
			return nil, err
		}
	}
	return call, nil
}

func (c *call) Invoke(ctx context.Context) error {
	msg := mqc.NewCallMessage(c.method)

	payload, err := c.serializer.Marshal(msg)
	if err != nil {
		return err
	}

	// Publish the call message to the invoke topic
	token := c.client.Publish(c.controlTopic, 2, false, payload)
	token.Wait()
	if err := token.Error(); err != nil {
		return err
	}

	// Wait for an acknowledgment
	ackMsg, err := c.Recv(ctx)
	if err != nil {
		return err
	}
	if !ackMsg.IsAck() {
		return mqc.ErrProtocolViolation
	}
	return nil
}

func (c *call) Recv(ctx context.Context) (*mqc.Message, error) {
	select {
	case msg := <-c.receiver:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *call) Send(ctx context.Context, msg *mqc.Message) error {
	if msg == nil {
		return errors.New("message is nil")
	}

	var err error
	var topic string
	var payload []byte

	if msg.IsData() {
		if c.server {
			topic = c.serverDataTopic
		} else {
			topic = c.clientDataTopic
		}

		payload = msg.Data
	} else {
		if c.server {
			topic = c.serverControlTopic
		} else {
			topic = c.clientControlTopic
		}

		payload, err = c.serializer.Marshal(msg)
		if err != nil {
			return err
		}
	}

	token := c.client.Publish(topic, 2, false, payload)

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

func (c *call) subscribe(topic string, data bool) error {
	token := c.client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		var m mqc.Message

		if data {
			m.Type = mqc.MsgTypeData
			m.Data = msg.Payload()
		} else {
			if err := c.serializer.Unmarshal(msg.Payload(), &m); err != nil {
				return
			}
		}
		// Handle incoming messages
		c.receiver <- &m

		msg.Ack()
	})
	if token == nil {
		return errors.New("failed to create subscription token")
	}
	token.Wait()
	return token.Error()
}

func (c *call) unsubscribe(topic string) error {
	token := c.client.Unsubscribe(topic)
	token.Wait()
	return token.Error()
}

func (c *call) Close() error {
	if c.server {
		err := c.unsubscribe(c.clientControlTopic)
		if err != nil {
			return err
		}
		return c.unsubscribe(c.clientDataTopic)
	}
	err := c.unsubscribe(c.serverControlTopic)
	if err != nil {
		return err
	}
	return c.unsubscribe(c.serverDataTopic)
}
