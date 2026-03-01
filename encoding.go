package bpid

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
)

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
