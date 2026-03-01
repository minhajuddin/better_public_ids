// Package bpid provides type-safe, prefixed public identifiers using Go generics.
//
// Each ID type is defined by creating a struct that implements the [Definer]
// interface. The struct's exported fields ARE the ID's data, serialized using
// [encoding/gob] and encoded as base64url without padding:
//
//	type UserID struct {
//	    OrgID   int64
//	    UserSeq int64
//	}
//	func (UserID) Prefix() string { return "user" }
//
//	id, err := bpid.New(UserID{OrgID: 42, UserSeq: 1001})
//	fmt.Println(id) // "user.<base64url(gob(data))>"
//
// IDs implement [fmt.Stringer], [encoding.TextMarshaler], [encoding.TextUnmarshaler],
// [encoding/json.Marshaler], [encoding/json.Unmarshaler], [encoding.BinaryMarshaler],
// [encoding.BinaryUnmarshaler], [database/sql.Scanner], and [database/sql/driver.Valuer].
//
// A global [DefaultRegistry] automatically tracks registered prefixes, enabling
// type-agnostic parsing via [ParseAny]. Custom [Registry] instances can be created
// for testing or isolation.
package bpid
