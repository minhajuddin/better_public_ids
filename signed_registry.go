package bpid

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strings"
)

const signatureBytes = 9 // truncated HMAC-SHA256 length (9 bytes → 12 base64url chars)

type signedRegistryConfig struct {
	oldKeys [][]byte
}

// SignedRegistryOption configures a [SignedRegistry].
type SignedRegistryOption func(*signedRegistryConfig) error

// WithOldKeys adds verification-only keys for key rotation. These keys can
// verify existing signatures but will not be used to sign new IDs.
func WithOldKeys(keys ...[]byte) SignedRegistryOption {
	return func(cfg *signedRegistryConfig) error {
		for _, k := range keys {
			if len(k) == 0 {
				return ErrInvalidKey
			}
		}
		cfg.oldKeys = append(cfg.oldKeys, keys...)
		return nil
	}
}

// SignedRegistry wraps a [Registry] and appends a truncated HMAC-SHA256
// signature to each serialized ID, making tampering detectable.
type SignedRegistry struct {
	registry   *Registry
	signingKey []byte   // signs new IDs
	oldKeys    [][]byte // can verify but won't sign
}

// NewSignedRegistry creates a new [SignedRegistry] that signs IDs produced by r.
func NewSignedRegistry(r *Registry, signingKey []byte, opts ...SignedRegistryOption) (*SignedRegistry, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: registry must not be nil", ErrInvalidKey)
	}
	if len(signingKey) == 0 {
		return nil, ErrInvalidKey
	}

	cfg := &signedRegistryConfig{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	// Copy keys to prevent caller mutation.
	sk := make([]byte, len(signingKey))
	copy(sk, signingKey)

	oldKeys := make([][]byte, len(cfg.oldKeys))
	for i, k := range cfg.oldKeys {
		oldKeys[i] = make([]byte, len(k))
		copy(oldKeys[i], k)
	}

	return &SignedRegistry{
		registry:   r,
		signingKey: sk,
		oldKeys:    oldKeys,
	}, nil
}

// MustNewSignedRegistry is like [NewSignedRegistry] but panics on error.
func MustNewSignedRegistry(r *Registry, signingKey []byte, opts ...SignedRegistryOption) *SignedRegistry {
	sr, err := NewSignedRegistry(r, signingKey, opts...)
	if err != nil {
		panic(fmt.Sprintf("bpid.MustNewSignedRegistry: %v", err))
	}
	return sr
}

// Separator returns the underlying registry's separator string.
func (sr *SignedRegistry) Separator() string {
	return sr.registry.Separator()
}

// Prefix verifies the signature on s, then extracts and returns the prefix.
func (sr *SignedRegistry) Prefix(s string) (string, error) {
	payload, err := sr.verifyAndStripSignature(s)
	if err != nil {
		return "", err
	}
	return sr.registry.Prefix(payload)
}

// sign computes a truncated HMAC-SHA256 of payload using key.
func sign(key []byte, payload string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	sum := mac.Sum(nil)
	return encodeBytes(sum[:signatureBytes])
}

// verifyAndStripSignature splits s into payload and signature, verifies the
// signature against the signing key and any old keys, and returns the payload.
func (sr *SignedRegistry) verifyAndStripSignature(s string) (string, error) {
	sep := sr.registry.separator
	idx := strings.LastIndex(s, sep)
	if idx < 0 {
		return "", fmt.Errorf("%w: no separator found", ErrInvalidFormat)
	}

	payload := s[:idx]
	candidate := s[idx+len(sep):]

	// Fast path: try the current signing key first.
	if hmac.Equal([]byte(sign(sr.signingKey, payload)), []byte(candidate)) {
		return payload, nil
	}

	// Slow path: try old keys in order.
	for _, key := range sr.oldKeys {
		if hmac.Equal([]byte(sign(key, payload)), []byte(candidate)) {
			return payload, nil
		}
	}

	return "", ErrInvalidSignature
}

// SignedSerialize encodes data into a signed, prefixed string.
func SignedSerialize[T any](sr *SignedRegistry, data T) (string, error) {
	unsigned, err := Serialize(sr.registry, data)
	if err != nil {
		return "", err
	}
	sig := sign(sr.signingKey, unsigned)
	return unsigned + sr.registry.separator + sig, nil
}

// MustSignedSerialize is like [SignedSerialize] but panics on error.
func MustSignedSerialize[T any](sr *SignedRegistry, data T) string {
	s, err := SignedSerialize(sr, data)
	if err != nil {
		panic(fmt.Sprintf("bpid.MustSignedSerialize: %v", err))
	}
	return s
}

// SignedDeserialize verifies the signature on s, then decodes it back into a
// value of type T.
func SignedDeserialize[T any](sr *SignedRegistry, s string) (T, error) {
	var zero T
	payload, err := sr.verifyAndStripSignature(s)
	if err != nil {
		return zero, err
	}
	return Deserialize[T](sr.registry, payload)
}
