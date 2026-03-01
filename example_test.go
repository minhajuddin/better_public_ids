package bpid_test

import (
	"fmt"

	bpid "github.com/minhajuddin/better_public_ids"
)

type UserID struct {
	OrgID   int64
	UserSeq int64
}

func (UserID) Prefix() string { return "user" }

type PostID struct {
	PostNum int64
}

func (PostID) Prefix() string { return "post" }

func init() {
	bpid.DefaultRegistry = bpid.MustNewRegistry(
		bpid.WithType[UserID](),
		bpid.WithType[PostID](),
	)
}

func ExampleNew() {
	id, err := bpid.New(UserID{OrgID: 42, UserSeq: 1001})
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Prefix())
	fmt.Println(id.IsZero())
	// Output:
	// user
	// false
}

func ExampleMustNew() {
	id := bpid.MustNew(UserID{OrgID: 42, UserSeq: 1001})
	fmt.Println(id.Prefix())
	fmt.Println(id.IsZero())
	// Output:
	// user
	// false
}

func ExampleID_Data() {
	id := bpid.MustNew(UserID{OrgID: 42, UserSeq: 1001})
	data, err := id.Data()
	if err != nil {
		panic(err)
	}
	fmt.Println(data.OrgID)
	fmt.Println(data.UserSeq)
	// Output:
	// 42
	// 1001
}

func ExampleParse_roundTrip() {
	id := bpid.MustNew(UserID{OrgID: 42, UserSeq: 1001})
	s := id.String()

	parsed, err := bpid.Parse[UserID](s)
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Equal(parsed))

	data, _ := parsed.Data()
	fmt.Println(data.OrgID)
	fmt.Println(data.UserSeq)
	// Output:
	// true
	// 42
	// 1001
}

func ExampleParseAny() {
	id := bpid.MustNew(UserID{OrgID: 42, UserSeq: 1001})
	s := id.String()

	prefix, rawBytes, err := bpid.ParseAny(s)
	if err != nil {
		panic(err)
	}
	fmt.Println(prefix)
	fmt.Println(len(rawBytes) > 0)
	// Output:
	// user
	// true
}

func ExampleID_IsZero() {
	var id bpid.ID[UserID]
	fmt.Println(id.IsZero())
	fmt.Println(id.String())
	// Output:
	// true
	//
}

func ExampleNewRegistry() {
	reg := bpid.MustNewRegistry(
		bpid.WithType[UserID](),
		bpid.WithSeparator("~"),
	)

	fmt.Println(reg.IsRegistered("user"))
	fmt.Println(reg.Separator())
	// Output:
	// true
	// ~
}
