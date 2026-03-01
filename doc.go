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
// Create a registry and register your types:
//
//	var registry = bpid.MustNewRegistry(
//	    bpid.WithType[UserID](),
//	)
//
// Serialize a struct into a prefixed string:
//
//	s, err := bpid.Serialize(registry, UserID{OrgID: 42, UserSeq: 1001})
//	// s = "user.<base64url(gob(data))>"
//
// Deserialize a prefixed string back into a struct:
//
//	data, err := bpid.Deserialize[UserID](registry, s)
//	// data = UserID{OrgID: 42, UserSeq: 1001}
//
// Extract the prefix for routing or switching:
//
//	prefix, err := registry.Prefix(s)
//	// prefix = "user"
//
// [Registry] is immutable after creation and safe for concurrent use.
// There is no global registry — consumers always create their own.
//
// Note: The encoded data uses Go's [encoding/gob] format, which means only Go
// programs can decode the data embedded in an ID. Non-Go consumers can still
// use IDs as opaque strings (compare, store, transmit) but cannot extract
// the embedded data.
package bpid
