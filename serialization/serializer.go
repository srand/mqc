package serialization

import "io"

type Decoder interface {
	Decode(v any) error
}

type Encoder interface {
	Encode(v any) error
}

type Serializer interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error

	NewDecoder(reader io.Reader) Decoder
	NewEncoder(writer io.Writer) Encoder
}
