package bpid

import (
	"bytes"
	"fmt"
	"strings"
)

// ID is a type-safe, prefixed identifier parameterized by a [Definer] type.
// The struct's exported fields are serialized using [encoding/gob].
// The zero value represents "no ID" and serializes as an empty string.
type ID[T Definer] struct {
	raw []byte // gob-encoded bytes of T; nil for zero value
}

// New creates a new ID by gob-encoding the provided data.
// The type's prefix must be registered in [DefaultRegistry].
func New[T Definer](data T) (ID[T], error) {
	var zero T
	prefix := zero.Prefix()
	if !DefaultRegistry.IsRegistered(prefix) {
		return ID[T]{}, fmt.Errorf("%w: %q", ErrUnknownPrefix, prefix)
	}
	raw, err := encodeGob(data)
	if err != nil {
		return ID[T]{}, err
	}
	return ID[T]{raw: raw}, nil
}

// MustNew is like [New] but panics on error.
func MustNew[T Definer](data T) ID[T] {
	id, err := New(data)
	if err != nil {
		panic(fmt.Sprintf("bpid.MustNew: %v", err))
	}
	return id
}

// Parse parses a prefixed ID string like "user.<base64url(gob(data))>".
// It validates that the prefix matches type T's prefix and that the encoded
// bytes are valid gob for type T. The type's prefix must be registered in [DefaultRegistry].
func Parse[T Definer](s string) (ID[T], error) {
	var zero T
	prefix := zero.Prefix()

	if !DefaultRegistry.IsRegistered(prefix) {
		return ID[T]{}, fmt.Errorf("%w: %q", ErrUnknownPrefix, prefix)
	}

	if s == "" {
		return ID[T]{}, ErrEmptyString
	}

	sep := DefaultRegistry.Separator()
	gotPrefix, encoded, ok := strings.Cut(s, sep)
	if !ok {
		return ID[T]{}, fmt.Errorf("%w: no separator %q found in %q", ErrInvalidFormat, sep, s)
	}

	if gotPrefix != prefix {
		return ID[T]{}, fmt.Errorf("%w: expected %q, got %q", ErrPrefixMismatch, prefix, gotPrefix)
	}

	raw, err := decodeBytes(encoded)
	if err != nil {
		return ID[T]{}, err
	}

	// Validate that the bytes actually decode to a T (catch corruption early)
	if _, err := decodeGob[T](raw); err != nil {
		return ID[T]{}, err
	}

	return ID[T]{raw: raw}, nil
}

// MustParse is like [Parse] but panics on error.
func MustParse[T Definer](s string) ID[T] {
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

// String returns the prefixed string representation.
// For the zero value, it returns "".
func (id ID[T]) String() string {
	if id.IsZero() {
		return ""
	}
	var zero T
	prefix := zero.Prefix()
	sep := DefaultRegistry.Separator()
	return prefix + sep + encodeBytes(id.raw)
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
