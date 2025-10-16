package serialization

import (
	"encoding/json"
	"io"
)

type JSONSerializer struct{}

func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

func (s *JSONSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *JSONSerializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (s *JSONSerializer) NewDecoder(reader io.Reader) Decoder {
	return json.NewDecoder(reader)
}

func (s *JSONSerializer) NewEncoder(writer io.Writer) Encoder {
	return json.NewEncoder(writer)
}
