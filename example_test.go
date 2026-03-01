package bpid_test

import (
	"fmt"

	bpid "github.com/minhajuddin/better_public_ids"
)

type UserID struct {
	OrgID   int64
	UserSeq int64
}

type PostID struct {
	PostNum int64
}

var exampleRegistry = bpid.MustNewRegistry(
	bpid.WithType[UserID]("user"),
	bpid.WithType[PostID]("post"),
)

func ExampleSerialize() {
	s, err := bpid.Serialize(exampleRegistry, UserID{OrgID: 42, UserSeq: 1001})
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
	// Output:
	// user.Kv-HAwEBBlVzZXJJRAH_iAABAgEFT3JnSUQBBAABB1VzZXJTZXEBBAAAAAn_iAFUAf4H0gA
}

func ExampleMustSerialize() {
	s := bpid.MustSerialize(exampleRegistry, UserID{OrgID: 42, UserSeq: 1001})
	fmt.Println(s)
	// Output:
	// user.Kv-HAwEBBlVzZXJJRAH_iAABAgEFT3JnSUQBBAABB1VzZXJTZXEBBAAAAAn_iAFUAf4H0gA
}

func ExampleDeserialize() {
	s := bpid.MustSerialize(exampleRegistry, UserID{OrgID: 42, UserSeq: 1001})
	data, err := bpid.Deserialize[UserID](exampleRegistry, s)
	if err != nil {
		panic(err)
	}
	fmt.Println(data.OrgID)
	fmt.Println(data.UserSeq)
	// Output:
	// 42
	// 1001
}

func ExampleDeserialize_roundTrip() {
	original := UserID{OrgID: 42, UserSeq: 1001}
	s, err := bpid.Serialize(exampleRegistry, original)
	if err != nil {
		panic(err)
	}

	data, err := bpid.Deserialize[UserID](exampleRegistry, s)
	if err != nil {
		panic(err)
	}
	fmt.Println(data.OrgID)
	fmt.Println(data.UserSeq)
	// Output:
	// 42
	// 1001
}

func ExampleRegistry_Prefix() {
	s := bpid.MustSerialize(exampleRegistry, UserID{OrgID: 42, UserSeq: 1001})
	prefix, err := exampleRegistry.Prefix(s)
	if err != nil {
		panic(err)
	}
	fmt.Println(prefix)
	// Output:
	// user
}

func ExampleNewRegistry() {
	reg := bpid.MustNewRegistry(
		bpid.WithType[UserID]("user"),
		bpid.WithSeparator("~"),
	)

	fmt.Println(reg.Separator())
	// Output:
	// ~
}
