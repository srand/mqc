package mqc

var (
	// ErrNoAddress indicates that no address was provided to connect to.
	ErrNoAddress         = &Error{"no address provided"}
	ErrProtocolViolation = &Error{"protocol violation"}
	ErrNilRequest        = &Error{"nil request"}
)

// Error represents an error in the mqc package.
type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}
