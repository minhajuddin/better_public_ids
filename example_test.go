package bpid_test

import (
	"fmt"

	bpid "github.com/minhajuddin/better_public_ids"
)

type UserIDDef struct {
	OrgID   int64
	UserSeq int64
}

func (UserIDDef) Prefix() string { return "user" }

type PostIDDef struct {
	PostNum int64
}

func (PostIDDef) Prefix() string { return "post" }

func ExampleNew() {
	id, err := bpid.New(UserIDDef{OrgID: 42, UserSeq: 1001})
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
	id := bpid.MustNew(UserIDDef{OrgID: 42, UserSeq: 1001})
	fmt.Println(id.Prefix())
	fmt.Println(id.IsZero())
	// Output:
	// user
	// false
}

func ExampleID_Data() {
	id := bpid.MustNew(UserIDDef{OrgID: 42, UserSeq: 1001})
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
	id := bpid.MustNew(UserIDDef{OrgID: 42, UserSeq: 1001})
	s := id.String()

	parsed, err := bpid.Parse[UserIDDef](s)
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
	// Ensure types are registered by using them
	_ = bpid.MustNew(UserIDDef{OrgID: 1, UserSeq: 1})
	_ = bpid.MustNew(PostIDDef{PostNum: 1})

	id := bpid.MustNew(UserIDDef{OrgID: 42, UserSeq: 1001})
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
	var id bpid.ID[UserIDDef]
	fmt.Println(id.IsZero())
	fmt.Println(id.String())
	// Output:
	// true
	//
}

func ExampleNewRegistry() {
	reg := bpid.NewRegistry(bpid.WithSeparator("~"))
	_ = reg.Register("user")

	// Build an ID and extract the encoded portion from the default-separator string
	id := bpid.MustNew(UserIDDef{OrgID: 10, UserSeq: 20})
	fullStr := id.String() // "user.<encoded>"
	encodedPart := fullStr[len("user."):]

	// ParseAny with the ~ separator
	prefix, _, err := reg.ParseAny("user~" + encodedPart)
	if err != nil {
		panic(err)
	}
	fmt.Println(prefix)
	fmt.Println(reg.Separator())
	// Output:
	// user
	// ~
}
