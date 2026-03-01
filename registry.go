package bpid

import (
	"fmt"
	"strings"
)

const defaultSeparator = "."

// registryConfig accumulates options before building an immutable [Registry].
type registryConfig struct {
	separator string
	prefixes  map[string]struct{}
}

// RegistryOption configures a [Registry].
type RegistryOption func(*registryConfig) error

// WithSeparator sets the separator character used between the prefix and the
// encoded data. Only "." and "~" are allowed.
func WithSeparator(sep string) RegistryOption {
	return func(cfg *registryConfig) error {
		if sep != "." && sep != "~" {
			return fmt.Errorf("%w: got %q", ErrInvalidSeparator, sep)
		}
		cfg.separator = sep
		return nil
	}
}

// WithType registers a [PublicID] type's prefix in the registry.
func WithType[T PublicID]() RegistryOption {
	var zero T
	prefix := zero.Prefix()
	return func(cfg *registryConfig) error {
		if err := validatePrefix(prefix); err != nil {
			return err
		}
		if _, exists := cfg.prefixes[prefix]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicatePrefix, prefix)
		}
		cfg.prefixes[prefix] = struct{}{}
		return nil
	}
}

// Registry is the central type. Immutable after creation, safe for concurrent use.
type Registry struct {
	separator string
	prefixes  map[string]struct{}
}

// NewRegistry creates a new [Registry] with the given options.
func NewRegistry(opts ...RegistryOption) (*Registry, error) {
	cfg := &registryConfig{
		separator: defaultSeparator,
		prefixes:  make(map[string]struct{}),
	}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	return &Registry{
		separator: cfg.separator,
		prefixes:  cfg.prefixes,
	}, nil
}

// MustNewRegistry is like [NewRegistry] but panics on error.
func MustNewRegistry(opts ...RegistryOption) *Registry {
	r, err := NewRegistry(opts...)
	if err != nil {
		panic(fmt.Sprintf("bpid.MustNewRegistry: %v", err))
	}
	return r
}

// Separator returns the registry's separator string.
func (r *Registry) Separator() string {
	return r.separator
}

// Prefix extracts the prefix from a serialized ID string.
// Returns [ErrUnregisteredPrefix] if the prefix is not in this registry.
func (r *Registry) Prefix(s string) (string, error) {
	if s == "" {
		return "", ErrEmptyString
	}
	prefix, _, ok := strings.Cut(s, r.separator)
	if !ok {
		return "", fmt.Errorf("%w: no separator %q found", ErrInvalidFormat, r.separator)
	}
	if _, ok := r.prefixes[prefix]; !ok {
		return "", fmt.Errorf("%w: %q", ErrUnregisteredPrefix, prefix)
	}
	return prefix, nil
}

// Serialize encodes a [PublicID] value into a prefixed string.
func Serialize[T PublicID](r *Registry, data T) (string, error) {
	var zero T
	prefix := zero.Prefix()
	if _, ok := r.prefixes[prefix]; !ok {
		return "", fmt.Errorf("%w: %q", ErrUnregisteredPrefix, prefix)
	}
	raw, err := encodeGob(data)
	if err != nil {
		return "", err
	}
	return prefix + r.separator + encodeBytes(raw), nil
}

// MustSerialize is like [Serialize] but panics on error.
func MustSerialize[T PublicID](r *Registry, data T) string {
	s, err := Serialize(r, data)
	if err != nil {
		panic(fmt.Sprintf("bpid.MustSerialize: %v", err))
	}
	return s
}

// Deserialize decodes a prefixed string back into a [PublicID] value.
func Deserialize[T PublicID](r *Registry, s string) (T, error) {
	var zero T
	prefix := zero.Prefix()
	if _, ok := r.prefixes[prefix]; !ok {
		return zero, fmt.Errorf("%w: %q", ErrUnregisteredPrefix, prefix)
	}
	if s == "" {
		return zero, ErrEmptyString
	}
	gotPrefix, encoded, ok := strings.Cut(s, r.separator)
	if !ok {
		return zero, fmt.Errorf("%w: no separator %q found", ErrInvalidFormat, r.separator)
	}
	if gotPrefix != prefix {
		return zero, fmt.Errorf("%w: expected %q, got %q", ErrPrefixMismatch, prefix, gotPrefix)
	}
	raw, err := decodeBytes(encoded)
	if err != nil {
		return zero, err
	}
	if len(raw) == 0 {
		return zero, fmt.Errorf("%w: empty encoded data", ErrInvalidFormat)
	}
	return decodeGob[T](raw)
}
