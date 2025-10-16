package serialization

var (
	ErrInvalidMessage = &SerializationError{Msg: "invalid message"}
)

type SerializationError struct {
	Msg string
}

func (e *SerializationError) Error() string {
	return e.Msg
}
