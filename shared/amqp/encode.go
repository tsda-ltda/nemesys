package amqp

import (
	"bytes"

	"github.com/vmihailenco/msgpack/v5"
)

// Encode returns the MessagePack encoding of v using a custom configuration for amqp.
func Encode(v any) ([]byte, error) {
	enc := msgpack.GetEncoder()
	enc.UseArrayEncodedStructs(true)
	var buf bytes.Buffer
	enc.Reset(&buf)

	err := enc.Encode(v)
	b := buf.Bytes()

	msgpack.PutEncoder(enc)

	if err != nil {
		return nil, err
	}
	return b, err
}

// Decode decodes the MessagePack-encoded data and stores the result
// in the value pointed to by v.
func Decode(data []byte, v any) error {
	dec := msgpack.GetDecoder()

	dec.Reset(bytes.NewReader(data))
	err := dec.Decode(v)

	msgpack.PutDecoder(dec)

	return err
}
