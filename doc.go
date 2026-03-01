// Package bpid provides type-safe, prefixed public identifiers using Go generics.
//
// Each ID type is a plain struct whose exported fields ARE the ID's data,
// serialized using [encoding/gob] and encoded as base64url without padding.
// The prefix is provided at registration time, not on the type itself.
//
// Define an ID type as a plain struct:
//
//	type UserID struct {
//	    OrgID   int64
//	    UserSeq int64
//	}
//
// Create a registry and register your types with their prefixes:
//
//	var registry = bpid.MustNewRegistry(
//	    bpid.WithType[UserID]("user"),
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
