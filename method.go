package mqc

import "fmt"

const (
	// MethodTypeUnary represents a unary RPC method.
	MethodTypeUnary = 0
	// MethodTypeServerStream represents a server streaming RPC method.
	MethodTypeServerStream = 1
	// MethodTypeClientStream represents a client streaming RPC method.
	MethodTypeClientStream = 2
	// MethodTypeBidiStream represents a bidirectional streaming RPC method.
	MethodTypeBidiStream = 3
	// MethodTypePublisher represents a publisher pub-sub method.
	MethodTypePublisher = 4
	// MethodTypeConsumer represents a consumer pub-sub method.
	MethodTypeConsumer = 5
)

// Method represents an RPC method name,
// that is used to identify the method being invoked.
// It is formatted as <service>/<method>.
type Method struct {
	Name string
	Type int
}

func NewMethod(name string, methodType int) *Method {
	return &Method{Name: name, Type: methodType}
}

func NewMethodFromString(s string) *Method {
	var m Method
	fmt.Sscanf(s, "%s/%d", &m.Name, &m.Type)
	return &m
}

func (m *Method) String() string {
	return fmt.Sprintf("%s/%d", m.Name, m.Type)
}

func (m *Method) IsPubSub() bool {
	return m.Type == MethodTypePublisher || m.Type == MethodTypeConsumer
}

func (m *Method) IsUnary() bool {
	return m.Type == MethodTypeUnary
}

func (m *Method) IsServerStream() bool {
	return m.Type == MethodTypeServerStream
}

func (m *Method) IsClientStream() bool {
	return m.Type == MethodTypeClientStream
}

func (m *Method) IsBidiStream() bool {
	return m.Type == MethodTypeBidiStream
}

func (m *Method) IsPublisher() bool {
	return m.Type == MethodTypePublisher
}

func (m *Method) IsConsumer() bool {
	return m.Type == MethodTypeConsumer
}
