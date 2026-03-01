package bpid

import (
	"bytes"
	"fmt"
	"strings"
)

// ID is a type-safe, prefixed identifier parameterized by a [PublicID] type.
// The struct's exported fields are serialized using [encoding/gob].
// The zero value represents "no ID" and serializes as an empty string.
type ID[T PublicID] struct {
	raw []byte // gob-encoded bytes of T; nil for zero value
}

// New creates a new ID by gob-encoding the provided data.
func New[T PublicID](data T) (ID[T], error) {
	raw, err := encodeGob(data)
	if err != nil {
		return ID[T]{}, err
	}
	return ID[T]{raw: raw}, nil
}

// MustNew is like [New] but panics on error.
func MustNew[T PublicID](data T) ID[T] {
	id, err := New(data)
	if err != nil {
		panic(fmt.Sprintf("bpid.MustNew: %v", err))
	}
	return id
}

// Parse parses a prefixed ID string like "user.<base64url(gob(data))>".
// It validates that the prefix matches type T's prefix. The separator is
// always "." regardless of any custom [Registry] separator.
func Parse[T PublicID](s string) (ID[T], error) {
	var zero T
	prefix := zero.Prefix()

	if s == "" {
		return ID[T]{}, ErrEmptyString
	}

	gotPrefix, encoded, ok := strings.Cut(s, defaultSeparator)
	if !ok {
		return ID[T]{}, fmt.Errorf("%w: no separator %q found in %q", ErrInvalidFormat, defaultSeparator, s)
	}

	if gotPrefix != prefix {
		return ID[T]{}, fmt.Errorf("%w: expected %q, got %q", ErrPrefixMismatch, prefix, gotPrefix)
	}

	raw, err := decodeBytes(encoded)
	if err != nil {
		return ID[T]{}, err
	}

	if len(raw) == 0 {
		return ID[T]{}, fmt.Errorf("%w: empty encoded data", ErrInvalidFormat)
	}

	return ID[T]{raw: raw}, nil
}

// MustParse is like [Parse] but panics on error.
func MustParse[T PublicID](s string) ID[T] {
	id, err := Parse[T](s)
	if err != nil {
		panic(fmt.Sprintf("bpid.MustParse: %v", err))
	}
	return id
}

// Data returns the decoded data. For a zero ID, it returns the zero value of T.
func (id ID[T]) Data() (T, error) {
	if id.IsZero() {
		var zero T
		return zero, nil
	}
	return decodeGob[T](id.raw)
}

// String returns the prefixed string representation using "." as the separator.
// For the zero value, it returns "".
func (id ID[T]) String() string {
	if id.IsZero() {
		return ""
	}
	var zero T
	return zero.Prefix() + defaultSeparator + encodeBytes(id.raw)
}

// IsZero reports whether the ID is the zero value (no data set).
func (id ID[T]) IsZero() bool {
	return len(id.raw) == 0
}

// Equal reports whether two IDs have the same underlying data.
func (id ID[T]) Equal(other ID[T]) bool {
	return bytes.Equal(id.raw, other.raw)
}

// Prefix returns the string prefix for this ID type.
func (id ID[T]) Prefix() string {
	var zero T
	return zero.Prefix()
}

// MarshalText implements [encoding.TextMarshaler].
// This enables JSON, YAML, and TOML serialization automatically.
func (id ID[T]) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler].
// This enables JSON, YAML, and TOML deserialization automatically.
func (id *ID[T]) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*id = ID[T]{}
		return nil
	}
	parsed, err := Parse[T](string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// GobEncode implements [encoding/gob.GobEncoder].
// Returns a copy of the raw gob-encoded bytes. The zero value returns nil.
func (id ID[T]) GobEncode() ([]byte, error) {
	if id.IsZero() {
		return nil, nil
	}
	out := make([]byte, len(id.raw))
	copy(out, id.raw)
	return out, nil
}

// GobDecode implements [encoding/gob.GobDecoder].
// Accepts gob-encoded bytes. Empty input sets the ID to the zero value.
func (id *ID[T]) GobDecode(data []byte) error {
	if len(data) == 0 {
		*id = ID[T]{}
		return nil
	}
	// Validate the bytes decode correctly
	if _, err := decodeGob[T](data); err != nil {
		return err
	}
	id.raw = make([]byte, len(data))
	copy(id.raw, data)
	return nil
}
