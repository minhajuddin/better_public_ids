// Package bpid provides type-safe, prefixed public identifiers using Go generics.
//
// Each ID type is defined by creating a struct that implements the [PublicID]
// interface. The struct's exported fields ARE the ID's data, serialized using
// [encoding/gob] and encoded as base64url without padding.
//
// Define an ID type by implementing [PublicID]:
//
//	type UserID struct {
//	    OrgID   int64
//	    UserSeq int64
//	}
//	func (UserID) Prefix() string { return "user" }
//
// Create and parse IDs directly — no registration required:
//
//	id, err := bpid.New(UserID{OrgID: 42, UserSeq: 1001})
//	fmt.Println(id) // "user.<base64url(gob(data))>"
//
// For type-agnostic parsing with [ParseAny], register types first:
//
//	func init() {
//	    bpid.RegisterType[UserID]()
//	}
//
// IDs implement [fmt.Stringer], [encoding.TextMarshaler], [encoding.TextUnmarshaler],
// [encoding/gob.GobEncoder], and [encoding/gob.GobDecoder]. The TextMarshaler/TextUnmarshaler
// implementations enable automatic JSON, YAML, and TOML support.
//
// The global registry ([DefaultRegistry]) is used by [ParseAny] for type-agnostic
// parsing. Register types with [RegisterType] before using [ParseAny]. The typed
// functions [New] and [Parse] do not require registration.
//
// Note: The encoded data uses Go's [encoding/gob] format, which means only Go
// programs can decode the data embedded in an ID. Non-Go consumers can still
// use IDs as opaque strings (compare, store, transmit) but cannot extract
// the embedded data.
package bpid
