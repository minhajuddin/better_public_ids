package bpid

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
)

// Codec marshals and unmarshals values to and from bytes.
// Implementations must be safe for concurrent use.
type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

// GobCodec is the default [Codec] that uses [encoding/gob].
// A fresh encoder/decoder is created on every call, so type descriptors are
// always included and the output is deterministic for equal inputs.
type GobCodec struct{}

// Marshal encodes v using gob.
func (GobCodec) Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncodingFailed, err)
	}
	return buf.Bytes(), nil
}

// Unmarshal decodes gob-encoded data into v.
func (GobCodec) Unmarshal(data []byte, v any) error {
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(v); err != nil {
		return fmt.Errorf("%w: %v", ErrDecodingFailed, err)
	}
	return nil
}

// encodeGob serializes data using encoding/gob. A fresh encoder is created
// each time to ensure type descriptors are always included, making the
// output deterministic for equal inputs.
func encodeGob[T any](data T) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncodingFailed, err)
	}
	return buf.Bytes(), nil
}

// decodeGob deserializes gob-encoded bytes back into a value of type T.
func decodeGob[T any](raw []byte) (T, error) {
	var result T
	dec := gob.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&result); err != nil {
		return result, fmt.Errorf("%w: %v", ErrDecodingFailed, err)
	}
	return result, nil
}

// encodeBytes encodes raw bytes to base64url without padding.
func encodeBytes(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// decodeBytes decodes a base64url string (without padding) into raw bytes.
func decodeBytes(s string) ([]byte, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidEncoding, err)
	}
	return b, nil
}
