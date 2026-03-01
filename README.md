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

### Define an ID type

Create a struct with exported fields and implement `Prefix()`:

```go
type UserID struct {
    OrgID   int64
    UserSeq int64
}

func (UserID) Prefix() string { return "user" }
```

### Create IDs

```go
id, err := bpid.New(UserID{OrgID: 42, UserSeq: 1001})
// or panic on error:
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

### Zero values

The zero value of `ID[T]` represents "no ID":

```go
var id bpid.ID[UserID]
id.IsZero()  // true
id.String()  // ""
```

JSON marshals as `null`, SQL stores as `NULL`.

### Registry — type-agnostic parsing

Prefixes are auto-registered when you create or parse IDs. Use `ParseAny` to extract prefix and raw bytes without knowing the type:

```go
prefix, rawBytes, err := bpid.ParseAny(s)
// prefix = "user", rawBytes = gob-encoded bytes
```

Custom registries with different separators:

```go
reg := bpid.NewRegistry(bpid.WithSeparator("~"))
reg.Register("user")
prefix, raw, err := reg.ParseAny("user~<encoded>")
```

## Interfaces

Every `ID[T]` implements:

- `fmt.Stringer`
- `encoding.TextMarshaler` / `TextUnmarshaler`
- `encoding/json.Marshaler` / `Unmarshaler`
- `encoding.BinaryMarshaler` / `BinaryUnmarshaler`
- `database/sql/driver.Valuer`
- `database/sql.Scanner`

## Development

```sh
make          # vet + build + test
make test     # go test ./...
make test-v   # verbose tests
make test-race # tests with race detector
make vet      # go vet ./...
make bench    # benchmarks with -benchmem
make fuzz     # all 3 fuzz targets (10s each)
make fuzz FUZZTIME=30s  # override fuzz duration
make clean    # clear test and fuzz caches
```
