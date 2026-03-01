# bpid — Better Public IDs

Type-safe, prefixed public identifiers for Go using generics. Each ID carries structured data serialized with `encoding/gob` and encoded as base64url. Zero external dependencies.

```
user.H4sIAAAAAAAAA6tWKkktLlGyUlAqS8wpTtVRSs7PS8nMS1eqBQBHnKYcHAAAAA
```

## Install

```sh
go get github.com/minhajuddin/better_public_ids
```

## Usage

### 1. Define ID types

Create structs with exported fields and implement `Prefix()`:

```go
type UserID struct {
    OrgID   int64
    UserSeq int64
}

func (UserID) Prefix() string { return "user" }

type PostID struct {
    BoardID int64
    PostSeq int64
}

func (PostID) Prefix() string { return "post" }
```

### 2. Register all types in a registry

```go
func init() {
    bpid.DefaultRegistry = bpid.MustNewRegistry(
        bpid.WithType[UserID](),
        bpid.WithType[PostID](),
    )
}
```

### 3. Create and use IDs

```go
id, err := bpid.New(UserID{OrgID: 42, UserSeq: 1001})
id := bpid.MustNew(UserID{OrgID: 42, UserSeq: 1001})
```

### Access data

```go
data, err := id.Data()
fmt.Println(data.OrgID)   // 42
fmt.Println(data.UserSeq) // 1001
```

### String representation and parsing

```go
s := id.String() // "user.<base64url(gob(data))>"

parsed, err := bpid.Parse[UserID](s)
fmt.Println(id.Equal(parsed)) // true
```

### Type-agnostic parsing

```go
prefix, rawBytes, err := bpid.ParseAny(s)
// prefix = "user", rawBytes = gob-encoded bytes
```

### Zero values

The zero value of `ID[T]` represents "no ID":

```go
var id bpid.ID[UserID]
id.IsZero()  // true
id.String()  // ""
```

### Custom registries

```go
reg := bpid.MustNewRegistry(
    bpid.WithType[UserID](),
    bpid.WithSeparator("~"),
)
```

## Interfaces

Every `ID[T]` implements:

- `fmt.Stringer`
- `encoding/gob.GobEncoder` / `GobDecoder`

## Development

```sh
make          # vet + build + test
make test     # go test ./...
make test-v   # verbose tests
make test-race # tests with race detector
make vet      # go vet ./...
make bench    # benchmarks with -benchmem
make fuzz     # fuzz FuzzParse (10s)
make fuzz FUZZTIME=30s  # override fuzz duration
make clean    # clear test and fuzz caches
```
