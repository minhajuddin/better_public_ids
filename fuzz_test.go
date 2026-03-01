package bpid

import (
	"testing"
)

type fuzzUserID struct {
	A int64
	B string
}

func (fuzzUserID) Prefix() string { return "fuzzuser" }

func FuzzParse(f *testing.F) {
	id := MustNew(fuzzUserID{A: 1, B: "hello"})
	f.Add(id.String())
	f.Add("")
	f.Add("fuzzuser.")
	f.Add(".")
	f.Add("noprefix")
	f.Add("fuzzuser.!!invalid!!")
	f.Add("post.AAAAAAAAAAAAAAAAAAAAAA")
	f.Add("a]b]c")
	f.Add("\x00\x01\x02")

	f.Fuzz(func(t *testing.T, s string) {
		id, err := Parse[fuzzUserID](s)
		if err != nil {
			return
		}
		// If parsing succeeded and ID is not zero, verify round-trip.
		if id.IsZero() {
			return
		}
		str := id.String()
		id2, err := Parse[fuzzUserID](str)
		if err != nil {
			t.Fatalf("round-trip failed: Parse(%q) succeeded, but Parse(%q) failed: %v", s, str, err)
		}
		if !id.Equal(id2) {
			t.Fatalf("round-trip mismatch: %v != %v", id, id2)
		}
	})
}
