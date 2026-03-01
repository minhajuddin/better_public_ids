package bpid

import "testing"

type fuzzUserID struct {
	A int64
	B string
}

func (fuzzUserID) Prefix() string { return "fuzzuser" }

var fuzzRegistry = MustNewRegistry(WithType[fuzzUserID]())

func FuzzDeserialize(f *testing.F) {
	s := MustSerialize(fuzzRegistry, fuzzUserID{A: 1, B: "hello"})
	f.Add(s)
	f.Add("")
	f.Add("fuzzuser.")
	f.Add(".")
	f.Add("noprefix")
	f.Add("fuzzuser.!!invalid!!")
	f.Add("post.AAAAAAAAAAAAAAAAAAAAAA")
	f.Add("a]b]c")
	f.Add("\x00\x01\x02")

	f.Fuzz(func(t *testing.T, s string) {
		data, err := Deserialize[fuzzUserID](fuzzRegistry, s)
		if err != nil {
			return
		}
		// If deserialization succeeded, verify round-trip.
		str, err := Serialize(fuzzRegistry, data)
		if err != nil {
			t.Fatalf("round-trip failed: Deserialize(%q) succeeded, but Serialize failed: %v", s, err)
		}
		data2, err := Deserialize[fuzzUserID](fuzzRegistry, str)
		if err != nil {
			t.Fatalf("round-trip failed: Serialize produced %q, but Deserialize failed: %v", str, err)
		}
		if data != data2 {
			t.Fatalf("round-trip mismatch: %+v != %+v", data, data2)
		}
	})
}
