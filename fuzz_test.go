package bpid

import (
	"testing"
)

type fuzzUserDef struct {
	A int64
	B string
}

func (fuzzUserDef) Prefix() string { return "fuzzuser" }

func FuzzParse(f *testing.F) {
	id := MustNew(fuzzUserDef{A: 1, B: "hello"})
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
		id, err := Parse[fuzzUserDef](s)
		if err != nil {
			return
		}
		// If parsing succeeded and ID is not zero, verify round-trip.
		if id.IsZero() {
			return
		}
		str := id.String()
		id2, err := Parse[fuzzUserDef](str)
		if err != nil {
			t.Fatalf("round-trip failed: Parse(%q) succeeded, but Parse(%q) failed: %v", s, str, err)
		}
		if !id.Equal(id2) {
			t.Fatalf("round-trip mismatch: %v != %v", id, id2)
		}
	})
}

func FuzzUnmarshalText(f *testing.F) {
	id := MustNew(fuzzUserDef{A: 1, B: "hello"})
	data, _ := id.MarshalText()
	f.Add(data)
	f.Add([]byte(""))
	f.Add([]byte("fuzzuser."))
	f.Add([]byte("wrong.AAAAAAAAAAAAAAAAAAAAAA"))
	f.Add([]byte("\xff\xfe\xfd"))

	f.Fuzz(func(t *testing.T, data []byte) {
		var id ID[fuzzUserDef]
		err := id.UnmarshalText(data)
		if err != nil {
			return
		}
		if id.IsZero() {
			return
		}
		out, err := id.MarshalText()
		if err != nil {
			t.Fatalf("MarshalText failed after successful UnmarshalText: %v", err)
		}
		var id2 ID[fuzzUserDef]
		if err := id2.UnmarshalText(out); err != nil {
			t.Fatalf("round-trip UnmarshalText failed: %v", err)
		}
		if !id.Equal(id2) {
			t.Fatalf("round-trip mismatch")
		}
	})
}

func FuzzUnmarshalJSON(f *testing.F) {
	id := MustNew(fuzzUserDef{A: 1, B: "hello"})
	data, _ := id.MarshalJSON()
	f.Add(data)
	f.Add([]byte(`null`))
	f.Add([]byte(`""`))
	f.Add([]byte(`123`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var id ID[fuzzUserDef]
		err := id.UnmarshalJSON(data)
		if err != nil {
			return
		}
		if id.IsZero() {
			return
		}
		out, err := id.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed after successful UnmarshalJSON: %v", err)
		}
		var id2 ID[fuzzUserDef]
		if err := id2.UnmarshalJSON(out); err != nil {
			t.Fatalf("round-trip UnmarshalJSON failed: %v", err)
		}
		if !id.Equal(id2) {
			t.Fatalf("round-trip mismatch")
		}
	})
}
