package bpid

import (
	"fmt"
	"strings"
	"sync"
)

const defaultSeparator = "."

// RegistryOption configures a [Registry].
type RegistryOption func(*Registry)

// WithSeparator sets the separator character used between the prefix and the
// encoded data. Only "." and "~" are allowed. Panics on invalid separator.
func WithSeparator(sep string) RegistryOption {
	if sep != "." && sep != "~" {
		panic(fmt.Sprintf("bpid: invalid separator %q: must be '.' or '~'", sep))
	}
	return func(r *Registry) {
		r.separator = sep
	}
}

// Registry holds a mapping of prefixes for type-agnostic parsing via [Registry.ParseAny].
type Registry struct {
	mu        sync.RWMutex
	separator string
	prefixes  map[string]struct{}
}

// NewRegistry creates a new [Registry] with the given options.
func NewRegistry(opts ...RegistryOption) *Registry {
	r := &Registry{
		separator: defaultSeparator,
		prefixes:  make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(r)
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

// DefaultRegistry is the global registry. ID types automatically register
// their prefixes here on first use.
var DefaultRegistry = NewRegistry()

// Register adds a prefix to the [DefaultRegistry].
func Register(prefix string) error {
	return DefaultRegistry.Register(prefix)
}

// ParseAny parses a prefixed ID string using the [DefaultRegistry].
func ParseAny(s string) (prefix string, rawBytes []byte, err error) {
	return DefaultRegistry.ParseAny(s)
}
