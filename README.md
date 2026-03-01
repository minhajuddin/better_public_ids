# bpid — Better Public IDs

Type-safe, prefixed public identifiers for Go using generics. Each ID carries structured data serialized with `encoding/gob` and encoded as base64url. Zero external dependencies.

```
user.Kv-HAwEBBlVzZXJJRAH_iAABAgEFT3JnSUQBBAABB1VzZXJTZXEBBAAAAAn_iAFUAf4H0gA
```

## Install

```sh
go get github.com/minhajuddin/better_public_ids
```

## Quick Start

Define plain structs for your ID types — no interfaces to implement:

```go
type UserID struct {
    OrgID   int64
    UserSeq int64
}

type PostID struct {
    PostNum int64
}
```

Create a registry, register types with prefixes, then serialize and deserialize:

```go
import bpid "github.com/minhajuddin/better_public_ids"

var registry = bpid.MustNewRegistry(
    bpid.WithType[UserID]("user"),
    bpid.WithType[PostID]("post"),
)

// Serialize a struct into a prefixed string.
s, err := bpid.Serialize(registry, UserID{OrgID: 42, UserSeq: 1001})
// s = "user.Kv-HAwEBBlVzZXJJRAH_iAABAgEFT3JnSUQBBAABB1VzZXJTZXEBBAAAAAn_iAFUAf4H0gA"

// Deserialize back into a typed struct.
data, err := bpid.Deserialize[UserID](registry, s)
// data.OrgID = 42, data.UserSeq = 1001
```

## Prefix Extraction

Extract the prefix from a serialized ID for routing or switching before deserializing:

```go
prefix, err := registry.Prefix(s)
// prefix = "user"

switch prefix {
case "user":
    data, err := bpid.Deserialize[UserID](registry, s)
case "post":
    data, err := bpid.Deserialize[PostID](registry, s)
}
```

## Signed IDs

`SignedRegistry` wraps a `Registry` and appends a truncated HMAC-SHA256 signature to each ID, making tampering detectable. The format is `prefix.encoded.signature`.

```go
key := []byte("your-secret-signing-key")
sr := bpid.MustNewSignedRegistry(registry, key)

// Serialize with signature.
s, err := bpid.SignedSerialize(sr, UserID{OrgID: 42, UserSeq: 1001})

// Deserialize — verifies the signature first, returns ErrInvalidSignature on tamper.
data, err := bpid.SignedDeserialize[UserID](sr, s)

// Prefix extraction also verifies the signature.
prefix, err := sr.Prefix(s)
```

Any modification to the prefix, encoded data, or signature causes `ErrInvalidSignature`.

## Key Rotation

Use `WithOldKeys` for zero-downtime key rotation. Old keys can verify existing signatures but won't sign new IDs:

```go
oldKey := []byte("old-secret-key")
newKey := []byte("new-secret-key")

sr := bpid.MustNewSignedRegistry(registry, newKey, bpid.WithOldKeys(oldKey))
// New IDs are signed with newKey.
// Old IDs signed with oldKey still verify.
// Once all old IDs have expired, remove oldKey.
```

## Custom Separators

The default separator is `"."`. You can use `"~"` instead (only `"."` and `"~"` are allowed):

```go
reg := bpid.MustNewRegistry(
    bpid.WithType[UserID]("user"),
    bpid.WithSeparator("~"),
)
// produces: user~Kv-HAwEB...
```

## Encoding Pipeline

```
Struct → gob → base64url (no padding) → "prefix.encoded"
```

Signed variant:

```
Struct → gob → base64url → "prefix.encoded" → HMAC-SHA256 → "prefix.encoded.signature"
```

The encoded data uses Go's `encoding/gob` format, so only Go programs can decode the embedded data. Non-Go consumers can still use IDs as opaque strings — compare, store, and transmit them freely.

## Error Handling

All errors are sentinel values, usable with `errors.Is`:

**Validation errors:**

| Error | Meaning |
|---|---|
| `ErrInvalidPrefix` | Prefix doesn't match `[a-z0-9][a-z0-9_-]*` |
| `ErrInvalidSeparator` | Separator is not `"."` or `"~"` |
| `ErrDuplicatePrefix` | Prefix already registered |
| `ErrDuplicateType` | Type already registered |
| `ErrUnregisteredPrefix` | Prefix not in registry |
| `ErrPrefixMismatch` | String prefix doesn't match expected type |

**Serialization errors:**

| Error | Meaning |
|---|---|
| `ErrEmptyString` | Attempted to deserialize an empty string |
| `ErrInvalidFormat` | Missing separator between prefix and data |
| `ErrInvalidEncoding` | Base64url portion is corrupt |
| `ErrEncodingFailed` | Gob encoding failed |
| `ErrDecodingFailed` | Gob decoding failed |

**Signature errors:**

| Error | Meaning |
|---|---|
| `ErrInvalidSignature` | HMAC signature doesn't match any known key |
| `ErrInvalidKey` | Signing or verification key is empty |

## API Reference

### Registry

```go
func NewRegistry(opts ...RegistryOption) (*Registry, error)
func MustNewRegistry(opts ...RegistryOption) *Registry

func WithType[T any](prefix string) RegistryOption
func WithSeparator(sep string) RegistryOption

func (*Registry) Prefix(s string) (string, error)
func (*Registry) Separator() string
func (*Registry) Inspect() string
```

### Serialize / Deserialize

```go
func Serialize[T any](r *Registry, data T) (string, error)
func MustSerialize[T any](r *Registry, data T) string
func Deserialize[T any](r *Registry, s string) (T, error)
```

### Signed Registry

```go
func NewSignedRegistry(r *Registry, signingKey []byte, opts ...SignedRegistryOption) (*SignedRegistry, error)
func MustNewSignedRegistry(r *Registry, signingKey []byte, opts ...SignedRegistryOption) *SignedRegistry

func WithOldKeys(keys ...[]byte) SignedRegistryOption

func (*SignedRegistry) Prefix(s string) (string, error)
func (*SignedRegistry) Separator() string
func (*SignedRegistry) Inspect() string

func SignedSerialize[T any](sr *SignedRegistry, data T) (string, error)
func MustSignedSerialize[T any](sr *SignedRegistry, data T) string
func SignedDeserialize[T any](sr *SignedRegistry, s string) (T, error)
```

## Full Example

A runnable example lives in [`example/main.go`](example/main.go). It registers three ID types — int64, UUID, and string-based — then demonstrates serialization, signed IDs, tamper detection, and key rotation.

Run it with:

```sh
go run ./example/
```

Output:

```
Registry: bpid.Registry(separator=".", types=3, registered=[inv→main.InviteID, order→main.OrderID, sess→main.SessionID])

OrderID serialized:   order.LH8DAQEHT3JkZXJJRAH_gAABAgEGU2hvcElEAQQAAQhPcmRlclNlcQEEAAAACf-AAVQB_gfSAA
OrderID deserialized: ShopID=42 OrderSeq=1001

SessionID serialized:   sess.If-BAwEBCVNlc3Npb25JRAH_ggABAQEEVVVJRAH_hAAAABn_gwEBAQlbMTZddWludDgB_4QAAQYBIAAAH_-CARBr_6f_uBD_nf-tEf_R_4D_tAD_wE__1DD_yAA
SessionID deserialized: UUID=6ba7b810-9dad-11d1-80b4-00c04fd430c8

InviteID serialized:   inv.Lf-FAwEBCEludml0ZUlEAf-GAAECAQlXb3Jrc3BhY2UBDAABBENvZGUBDAAAABX_hgEJYWNtZS1jb3JwAQV4SzltUQA
InviteID deserialized: Workspace="acme-corp" Code="xK9mQ"

--- Prefix extraction ---
  order.LH8DAQEHT3JkZX...  →  prefix="order"
  sess.If-BAwEBCVNlc3N...  →  prefix="sess"
  inv.Lf-FAwEBCEludml0...  →  prefix="inv"

========================================
  Signed Registry
========================================

SignedRegistry: bpid.SignedRegistry(signingKey=6d792d73..., oldKeys=0, registry=bpid.Registry(separator=".", types=3, registered=[inv→main.InviteID, order→main.OrderID, sess→main.SessionID]))

Signed OrderID:   order.LH8DAQEHT3JkZXJJRAH_gAABAgEGU2hvcElEAQQAAQhPcmRlclNlcQEEAAAACf-AAVQB_gfSAA.D3OA5qQMWZ6z
Deserialized:     ShopID=42 OrderSeq=1001

--- Tamper detection ---
Tampered ID rejected: bpid: invalid signature

--- Key rotation ---
Signed with old key: inv.Lf-FAwEBCEludml0ZUlEAf-GAA...
Old ID still valid:  Workspace="acme-corp" Code="xK9mQ"
Signed with new key: inv.Lf-FAwEBCEludml0ZUlEAf-GAA...
New ID valid:        Workspace="acme-corp" Code="xK9mQ"
```

## Development

```sh
make              # vet + build + test
make test         # go test ./...
make test-v       # verbose tests
make test-race    # tests with race detector
make vet          # go vet ./...
make bench        # benchmarks with -benchmem
make fuzz         # fuzz FuzzDeserialize (10s)
make fuzz FUZZTIME=30s  # override fuzz duration
make clean        # clear test and fuzz caches
```
