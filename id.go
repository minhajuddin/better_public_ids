package bpid

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// registrationOnces tracks per-prefix sync.Once instances for lazy registration.
var registrationOnces sync.Map // map[string]*sync.Once

// ensureRegistered lazily registers the prefix for type T in the [DefaultRegistry].
// It is safe for concurrent use and runs at most once per distinct prefix.
func ensureRegistered[T Definer]() string {
	var zero T
	prefix := zero.Prefix()
	once, _ := registrationOnces.LoadOrStore(prefix, &sync.Once{})
	once.(*sync.Once).Do(func() {
		// Best-effort registration. If it fails (e.g., duplicate from manual
		// Register call), that's fine — the prefix is already known.
		_ = DefaultRegistry.Register(prefix)
	})
	return prefix
}

// ID is a type-safe, prefixed identifier parameterized by a [Definer] type.
// The struct's exported fields are serialized using [encoding/gob].
// The zero value represents "no ID" and serializes as an empty string (or JSON null).
type ID[T Definer] struct {
	raw []byte // gob-encoded bytes of T; nil for zero value
}

// New creates a new ID by gob-encoding the provided data.
func New[T Definer](data T) (ID[T], error) {
	ensureRegistered[T]()
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
// bytes are valid gob for type T.
func Parse[T Definer](s string) (ID[T], error) {
	prefix := ensureRegistered[T]()

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
	prefix := ensureRegistered[T]()
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

// MarshalText implements [encoding.TextMarshaler].
// The zero value marshals to an empty byte slice.
func (id ID[T]) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler].
// An empty input sets the ID to the zero value.
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

// MarshalJSON implements [encoding/json.Marshaler].
// The zero value marshals as JSON null.
func (id ID[T]) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(id.String())
}

// UnmarshalJSON implements [encoding/json.Unmarshaler].
// JSON null or empty string sets the ID to the zero value.
func (id *ID[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*id = ID[T]{}
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("bpid: invalid JSON: %w", err)
	}
	if s == "" {
		*id = ID[T]{}
		return nil
	}
	parsed, err := Parse[T](s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// MarshalBinary implements [encoding.BinaryMarshaler].
// Returns the raw gob-encoded bytes. The zero value returns nil.
func (id ID[T]) MarshalBinary() ([]byte, error) {
	if id.IsZero() {
		return nil, nil
	}
	out := make([]byte, len(id.raw))
	copy(out, id.raw)
	return out, nil
}

// UnmarshalBinary implements [encoding.BinaryUnmarshaler].
// Accepts gob-encoded bytes. Empty input sets the ID to the zero value.
func (id *ID[T]) UnmarshalBinary(data []byte) error {
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

// Value implements [database/sql/driver.Valuer].
// Stores the prefixed string representation in the database.
// The zero value returns nil (SQL NULL).
func (id ID[T]) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, nil
	}
	return id.String(), nil
}

// Scan implements [database/sql.Scanner].
// Accepts string, []byte, or nil.
func (id *ID[T]) Scan(src any) error {
	if src == nil {
		*id = ID[T]{}
		return nil
	}
	switch v := src.(type) {
	case string:
		if v == "" {
			*id = ID[T]{}
			return nil
		}
		parsed, err := Parse[T](v)
		if err != nil {
			return err
		}
		*id = parsed
		return nil
	case []byte:
		if len(v) == 0 {
			*id = ID[T]{}
			return nil
		}
		parsed, err := Parse[T](string(v))
		if err != nil {
			return err
		}
		*id = parsed
		return nil
	default:
		return fmt.Errorf("%w: got %T", ErrScanType, src)
	}
}
