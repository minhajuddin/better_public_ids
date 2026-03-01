package bpid

import "errors"

var (
	// ErrInvalidPrefix is returned when a prefix does not match the allowed
	// pattern [a-z0-9][a-z0-9_-]*.
	ErrInvalidPrefix = errors.New("bpid: invalid prefix: must match [a-z0-9][a-z0-9_-]*")

	// ErrPrefixMismatch is returned when parsing a string whose prefix does
	// not match the expected type's prefix.
	ErrPrefixMismatch = errors.New("bpid: prefix mismatch")

	// ErrInvalidFormat is returned when a string does not contain a separator
	// between prefix and encoded data.
	ErrInvalidFormat = errors.New("bpid: invalid format: expected prefix<sep>encoded")

	// ErrInvalidEncoding is returned when the base64url portion cannot be decoded.
	ErrInvalidEncoding = errors.New("bpid: invalid base64url encoding")

	// ErrEmptyString is returned when attempting to parse an empty string.
	ErrEmptyString = errors.New("bpid: cannot parse empty string")

	// ErrUnknownPrefix is returned by [Registry.ParseAny] when the prefix is
	// not registered.
	ErrUnknownPrefix = errors.New("bpid: unknown prefix")

	// ErrDuplicatePrefix is returned when registering a prefix that is already
	// registered.
	ErrDuplicatePrefix = errors.New("bpid: duplicate prefix")

	// ErrEncodingFailed is returned when gob encoding of the data fails.
	ErrEncodingFailed = errors.New("bpid: failed to encode data")

	// ErrDecodingFailed is returned when gob decoding of the data fails.
	ErrDecodingFailed = errors.New("bpid: failed to decode data")

	// ErrInvalidSeparator is returned when an invalid separator is provided.
	ErrInvalidSeparator = errors.New("bpid: invalid separator: must be '.' or '~'")
)
