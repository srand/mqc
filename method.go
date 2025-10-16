package mqc

// Method represents an RPC method name,
// that is used to identify the method being invoked.
// It is formatted as <service>/<method>.
type Method string

func (m Method) String() string {
	return string(m)
}
