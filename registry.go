package bpid

import (
	"fmt"
	"strings"
	"sync"
)

const defaultSeparator = "."

// RegistryOption configures a [Registry].
type RegistryOption func(*Registry) error

// WithSeparator sets the separator character used between the prefix and the
// encoded data. Only "." and "~" are allowed.
func WithSeparator(sep string) RegistryOption {
	return func(r *Registry) error {
		if sep != "." && sep != "~" {
			return fmt.Errorf("%w: got %q", ErrInvalidSeparator, sep)
		}
		r.separator = sep
		return nil
	}
}

// WithType registers a [PublicID] type's prefix in the registry.
func WithType[T PublicID]() RegistryOption {
	var zero T
	prefix := zero.Prefix()
	return func(r *Registry) error {
		return r.Register(prefix)
	}
}

// Registry holds a mapping of prefixes for type-agnostic parsing via [Registry.ParseAny].
type Registry struct {
	mu        sync.RWMutex
	separator string
	prefixes  map[string]struct{}
}

// NewRegistry creates a new [Registry] with the given options.
func NewRegistry(opts ...RegistryOption) (*Registry, error) {
	r := &Registry{
		separator: defaultSeparator,
		prefixes:  make(map[string]struct{}),
	}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// MustNewRegistry is like [NewRegistry] but panics on error.
func MustNewRegistry(opts ...RegistryOption) *Registry {
	r, err := NewRegistry(opts...)
	if err != nil {
		panic(fmt.Sprintf("bpid.MustNewRegistry: %v", err))
	}
	return r
}

// Register adds a prefix to the registry.
func (r *Registry) Register(prefix string) error {
	if err := validatePrefix(prefix); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.prefixes[prefix]; exists {
		return fmt.Errorf("%w: %q", ErrDuplicatePrefix, prefix)
	}
	r.prefixes[prefix] = struct{}{}
	return nil
}

// IsRegistered reports whether a prefix has been registered.
func (r *Registry) IsRegistered(prefix string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.prefixes[prefix]
	return ok
}

// ParseAny parses a prefixed ID string without knowing its type. It returns
// the prefix and the raw gob-encoded bytes. The prefix must be registered.
func (r *Registry) ParseAny(s string) (prefix string, rawBytes []byte, err error) {
	if s == "" {
		return "", nil, ErrEmptyString
	}

	r.mu.RLock()
	sep := r.separator
	r.mu.RUnlock()

	prefix, encoded, ok := strings.Cut(s, sep)
	if !ok {
		return "", nil, fmt.Errorf("%w: no separator %q found", ErrInvalidFormat, sep)
	}

	r.mu.RLock()
	_, ok = r.prefixes[prefix]
	r.mu.RUnlock()

	if !ok {
		return "", nil, fmt.Errorf("%w: %q", ErrUnknownPrefix, prefix)
	}

	rawBytes, err = decodeBytes(encoded)
	if err != nil {
		return "", nil, err
	}

	return prefix, rawBytes, nil
}

// Separator returns the registry's separator string.
func (r *Registry) Separator() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.separator
}

// defaultRegistry is the global registry used by [ParseAny] and [RegisterType].
// It is not directly replaceable; use [RegisterType] to add prefixes.
var defaultRegistry = MustNewRegistry()

// DefaultRegistry returns the global registry. Use [RegisterType] to register
// prefixes rather than calling Register on the returned registry directly.
func DefaultRegistry() *Registry {
	return defaultRegistry
}

// RegisterType registers a [PublicID] type's prefix in the global registry.
// This must be called (typically in init or main) before using [ParseAny].
// It is idempotent: registering an already-registered prefix is a no-op.
func RegisterType[T PublicID]() {
	var zero T
	prefix := zero.Prefix()
	if defaultRegistry.IsRegistered(prefix) {
		return
	}
	if err := defaultRegistry.Register(prefix); err != nil {
		panic(fmt.Sprintf("bpid.RegisterType: %v", err))
	}
}

// ParseAny parses a prefixed ID string using the global registry.
func ParseAny(s string) (prefix string, rawBytes []byte, err error) {
	return defaultRegistry.ParseAny(s)
}
