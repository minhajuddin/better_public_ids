package bpid_test

import (
	"encoding/json"
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

func ExampleID_MarshalJSON() {
	type User struct {
		ID   bpid.ID[UserIDDef] `json:"id"`
		Name string             `json:"name"`
	}

	id := bpid.MustNew(UserIDDef{OrgID: 42, UserSeq: 1001})
	user := User{
		ID:   id,
		Name: "Alice",
	}
	data, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}

	// Verify it round-trips
	var user2 User
	if err := json.Unmarshal(data, &user2); err != nil {
		panic(err)
	}
	fmt.Println(user.ID.Equal(user2.ID))
	fmt.Println(user2.Name)
	// Output:
	// true
	// Alice
}

func ExampleID_MarshalJSON_zero() {
	type User struct {
		ID   bpid.ID[UserIDDef] `json:"id"`
		Name string             `json:"name"`
	}
	user := User{Name: "Bob"}
	data, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	// Output:
	// {"id":null,"name":"Bob"}
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
