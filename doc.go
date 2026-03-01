// Package bpid provides type-safe, prefixed public identifiers using Go generics.
//
// Each ID type is defined by creating a struct that implements the [Definer]
// interface. The struct's exported fields ARE the ID's data, serialized using
// [encoding/gob] and encoded as base64url without padding.
//
// All types must be registered in a [Registry] before use:
//
//	type UserID struct {
//	    OrgID   int64
//	    UserSeq int64
//	}
//	func (UserID) Prefix() string { return "user" }
//
//	func init() {
//	    bpid.DefaultRegistry = bpid.MustNewRegistry(
//	        bpid.WithType[UserID](),
//	    )
//	}
//
//	id, err := bpid.New(UserID{OrgID: 42, UserSeq: 1001})
//	fmt.Println(id) // "user.<base64url(gob(data))>"
//
// IDs implement [fmt.Stringer], [encoding/gob.GobEncoder], and [encoding/gob.GobDecoder].
//
// The [DefaultRegistry] is used by top-level functions like [New], [Parse],
// and [ParseAny]. Custom [Registry] instances can be created for testing
// or isolation.
package bpid
