package bpid

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// Test PublicID types (data-carrying structs)
type testUserID struct {
	OrgID   int64
	UserSeq int64
}

func (testUserID) Prefix() string { return "user" }

type testPostID struct {
	BoardID int64
	PostSeq int64
}

func (testPostID) Prefix() string { return "post" }

// Compile-time interface compliance checks
var (
	_ fmt.Stringer             = ID[testUserID]{}
	_ gob.GobEncoder           = ID[testUserID]{}
	_ gob.GobDecoder           = (*ID[testUserID])(nil)
	_ encoding.TextMarshaler   = ID[testUserID]{}
	_ encoding.TextUnmarshaler = (*ID[testUserID])(nil)
)

// --- Constructor Tests ---

func TestNew(t *testing.T) {
	data := testUserID{OrgID: 42, UserSeq: 1001}
	id, err := New(data)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if id.IsZero() {
		t.Error("New() returned zero ID")
	}
	if id.Prefix() != "user" {
		t.Errorf("Prefix() = %q, want %q", id.Prefix(), "user")
	}
	if !strings.HasPrefix(id.String(), "user.") {
		t.Errorf("String() = %q, want prefix 'user.'", id.String())
	}

	got, err := id.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	if got != data {
		t.Errorf("Data() = %+v, want %+v", got, data)
	}
}

func TestNewDifferentData(t *testing.T) {
	id1, _ := New(testUserID{OrgID: 1, UserSeq: 1})
	id2, _ := New(testUserID{OrgID: 1, UserSeq: 2})

	if id1.Equal(id2) {
		t.Error("IDs with different data should not be equal")
	}
}

func TestNewSameData(t *testing.T) {
	data := testUserID{OrgID: 42, UserSeq: 1001}
	id1, _ := New(data)
	id2, _ := New(data)

	if !id1.Equal(id2) {
		t.Error("IDs with same data should be equal (gob is deterministic)")
	}
}

func TestMustNew(t *testing.T) {
	data := testUserID{OrgID: 42, UserSeq: 1001}
	id := MustNew(data)
	if id.IsZero() {
		t.Error("MustNew returned zero ID")
	}
	got, _ := id.Data()
	if got != data {
		t.Errorf("Data() = %+v, want %+v", got, data)
	}
}

func TestParse(t *testing.T) {
	data := testUserID{OrgID: 42, UserSeq: 1001}
	validID, _ := New(data)
	validStr := validID.String()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "valid", input: validStr},
		{name: "empty string", input: "", wantErr: ErrEmptyString},
		{name: "wrong prefix", input: strings.Replace(validStr, "user.", "post.", 1), wantErr: ErrPrefixMismatch},
		{name: "no separator", input: "user" + "ABCDEF", wantErr: ErrInvalidFormat},
		{name: "invalid encoding", input: "user.!!invalid!!", wantErr: ErrInvalidEncoding},
		{name: "only prefix", input: "user.", wantErr: ErrInvalidFormat},
		{name: "whitespace prefix", input: " user." + encodeBytes(validID.raw), wantErr: ErrPrefixMismatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := Parse[testUserID](tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Parse(%q) = nil error, want %v", tt.input, tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Parse(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}
			got, err := id.Data()
			if err != nil {
				t.Fatalf("Data: %v", err)
			}
			if got != data {
				t.Errorf("Parse(%q) Data() = %+v, want %+v", tt.input, got, data)
			}
		})
	}
}

func TestParseRoundTrip(t *testing.T) {
	data := testUserID{OrgID: 42, UserSeq: 1001}
	id, _ := New(data)
	s := id.String()

	parsed, err := Parse[testUserID](s)
	if err != nil {
		t.Fatalf("Parse(%q) error: %v", s, err)
	}
	if !id.Equal(parsed) {
		t.Error("round-trip failed")
	}
	got, _ := parsed.Data()
	if got != data {
		t.Errorf("round-trip Data() = %+v, want %+v", got, data)
	}
}

func TestMustParse(t *testing.T) {
	id, _ := New(testUserID{OrgID: 42, UserSeq: 1001})
	s := id.String()

	parsed := MustParse[testUserID](s)
	if !id.Equal(parsed) {
		t.Error("MustParse round-trip failed")
	}
}

func TestMustParsePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse with invalid input should panic")
		}
	}()
	MustParse[testUserID]("invalid")
}

// --- Method Tests ---

func TestString(t *testing.T) {
	id, _ := New(testUserID{OrgID: 42, UserSeq: 1001})
	s := id.String()

	if !strings.HasPrefix(s, "user.") {
		t.Errorf("String() = %q, want prefix 'user.'", s)
	}
	if len(s) <= len("user.") {
		t.Errorf("String() = %q, encoded part is empty", s)
	}
}

func TestStringZero(t *testing.T) {
	var id ID[testUserID]
	if got := id.String(); got != "" {
		t.Errorf("zero ID String() = %q, want %q", got, "")
	}
}

func TestStringDeterministic(t *testing.T) {
	id, _ := New(testUserID{OrgID: 42, UserSeq: 1001})
	s1 := id.String()
	s2 := id.String()
	if s1 != s2 {
		t.Errorf("String() not deterministic: %q != %q", s1, s2)
	}

	parts := strings.SplitN(s1, ".", 2)
	if len(parts) != 2 {
		t.Fatalf("String() = %q, want format 'prefix.encoded'", s1)
	}
	if parts[0] != "user" {
		t.Errorf("prefix = %q, want %q", parts[0], "user")
	}
}

func TestData(t *testing.T) {
	data := testUserID{OrgID: 42, UserSeq: 1001}
	id, _ := New(data)

	got, err := id.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	if got != data {
		t.Errorf("Data() = %+v, want %+v", got, data)
	}
}

func TestDataZero(t *testing.T) {
	var id ID[testUserID]
	got, err := id.Data()
	if err != nil {
		t.Fatalf("Data on zero ID: %v", err)
	}
	if got != (testUserID{}) {
		t.Errorf("zero ID Data() = %+v, want zero value", got)
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		name string
		id   ID[testUserID]
		want bool
	}{
		{name: "zero value", id: ID[testUserID]{}, want: true},
		{name: "new ID", id: MustNew(testUserID{OrgID: 1, UserSeq: 1}), want: false},
		{name: "new with zero data", id: MustNew(testUserID{}), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	data := testUserID{OrgID: 42, UserSeq: 1001}
	id1, _ := New(data)
	id2, _ := New(data)
	id3, _ := New(testUserID{OrgID: 99, UserSeq: 2002})

	if !id1.Equal(id2) {
		t.Error("same data IDs should be equal")
	}
	if id1.Equal(id3) {
		t.Error("different data IDs should not be equal")
	}

	var z1, z2 ID[testUserID]
	if !z1.Equal(z2) {
		t.Error("two zero IDs should be equal")
	}
}

func TestPrefix(t *testing.T) {
	id := MustNew(testUserID{OrgID: 1, UserSeq: 1})
	if id.Prefix() != "user" {
		t.Errorf("Prefix() = %q, want %q", id.Prefix(), "user")
	}

	pid := MustNew(testPostID{BoardID: 1, PostSeq: 1})
	if pid.Prefix() != "post" {
		t.Errorf("Prefix() = %q, want %q", pid.Prefix(), "post")
	}

	var zero ID[testUserID]
	if zero.Prefix() != "user" {
		t.Errorf("zero Prefix() = %q, want %q", zero.Prefix(), "user")
	}
}

// --- Gob Encode/Decode Tests ---

func TestGobEncodeRoundTrip(t *testing.T) {
	id := MustNew(testUserID{OrgID: 42, UserSeq: 1001})

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&id); err != nil {
		t.Fatalf("gob.Encode: %v", err)
	}

	var parsed ID[testUserID]
	if err := gob.NewDecoder(&buf).Decode(&parsed); err != nil {
		t.Fatalf("gob.Decode: %v", err)
	}

	if !id.Equal(parsed) {
		t.Error("gob round-trip failed")
	}
	got, _ := parsed.Data()
	if got != (testUserID{OrgID: 42, UserSeq: 1001}) {
		t.Errorf("Data() after gob round-trip = %+v, want {42 1001}", got)
	}
}

func TestGobEncodeZero(t *testing.T) {
	var id ID[testUserID]

	data, err := id.GobEncode()
	if err != nil {
		t.Fatalf("GobEncode: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("zero ID GobEncode returned %d bytes, want 0", len(data))
	}

	var parsed ID[testUserID]
	if err := parsed.GobDecode(data); err != nil {
		t.Fatalf("GobDecode: %v", err)
	}
	if !parsed.IsZero() {
		t.Error("GobDecode(nil) should produce zero ID")
	}
}

func TestGobDecodeInvalidBytes(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "garbage", data: []byte{0xFF, 0xFE, 0xFD}},
		{name: "single byte", data: []byte{0x0A}},
		{name: "truncated", data: []byte{0x00, 0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id ID[testUserID]
			err := id.GobDecode(tt.data)
			if err == nil {
				t.Error("GobDecode should error on invalid gob bytes")
			}
			if !errors.Is(err, ErrDecodingFailed) {
				t.Errorf("error = %v, want ErrDecodingFailed", err)
			}
		})
	}
}

// --- Cross-Type Safety Tests ---

func TestParseWrongType(t *testing.T) {
	postID := MustNew(testPostID{BoardID: 1, PostSeq: 1})
	postStr := postID.String()

	_, err := Parse[testUserID](postStr)
	if err == nil {
		t.Fatal("Parse[testUserID] with postID string should error")
	}
	if !errors.Is(err, ErrPrefixMismatch) {
		t.Errorf("error = %v, want ErrPrefixMismatch", err)
	}
}

func TestMultipleTypes(t *testing.T) {
	userID := MustNew(testUserID{OrgID: 1, UserSeq: 1})
	postID := MustNew(testPostID{BoardID: 1, PostSeq: 1})

	if userID.Prefix() == postID.Prefix() {
		t.Error("different types should have different prefixes")
	}

	if strings.HasPrefix(userID.String(), "post.") {
		t.Error("userID should not have post prefix")
	}
	if strings.HasPrefix(postID.String(), "user.") {
		t.Error("postID should not have user prefix")
	}
}

// --- Registration Tests ---

type testAutoRegID struct {
	Val int64
}

func (testAutoRegID) Prefix() string { return "autoreg" }

type testUnregID struct {
	X int64
}

func (testUnregID) Prefix() string { return "unreg" }

func TestNewUnregistered(t *testing.T) {
	// New no longer requires registration — only ParseAny does.
	id, err := New(testUnregID{X: 1})
	if err != nil {
		t.Fatalf("New with unregistered type should succeed: %v", err)
	}
	if id.IsZero() {
		t.Error("New should return non-zero ID")
	}
}

func TestParseUnregistered(t *testing.T) {
	// Parse no longer requires registration — only ParseAny does.
	id, err := New(testUnregID{X: 42})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	parsed, err := Parse[testUnregID](id.String())
	if err != nil {
		t.Fatalf("Parse with unregistered type should succeed: %v", err)
	}
	if !id.Equal(parsed) {
		t.Error("round-trip failed for unregistered type")
	}
}

// --- Zero Value Tests ---

func TestZeroValueBehavior(t *testing.T) {
	var id ID[testUserID]

	if !id.IsZero() {
		t.Error("zero value IsZero() should be true")
	}
	if id.String() != "" {
		t.Errorf("zero value String() = %q, want %q", id.String(), "")
	}
	if id.Prefix() != "user" {
		t.Errorf("zero value Prefix() = %q, want %q", id.Prefix(), "user")
	}

	data, err := id.Data()
	if err != nil {
		t.Fatalf("zero value Data() error: %v", err)
	}
	if data != (testUserID{}) {
		t.Errorf("zero value Data() = %+v, want zero", data)
	}

	var other ID[testUserID]
	if !id.Equal(other) {
		t.Error("two zero values should be equal")
	}
}

// --- Concurrency Tests ---

func TestConcurrentNew(t *testing.T) {
	const n = 100
	ids := make([]ID[testUserID], n)
	errs := make([]error, n)
	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ids[i], errs[i] = New(testUserID{OrgID: int64(i), UserSeq: int64(i * 10)})
		}(i)
	}
	wg.Wait()

	seen := make(map[string]bool, n)
	for i, id := range ids {
		if errs[i] != nil {
			t.Errorf("ids[%d] error: %v", i, errs[i])
			continue
		}
		if id.IsZero() {
			t.Errorf("ids[%d] is zero", i)
		}
		s := id.String()
		if seen[s] {
			t.Errorf("ids[%d] is a duplicate: %s", i, s)
		}
		seen[s] = true
	}
}

func TestConcurrentParse(t *testing.T) {
	id := MustNew(testUserID{OrgID: 42, UserSeq: 1001})
	s := id.String()

	const n = 100
	results := make([]ID[testUserID], n)
	errs := make([]error, n)
	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = Parse[testUserID](s)
		}(i)
	}
	wg.Wait()

	for i := range n {
		if errs[i] != nil {
			t.Errorf("Parse[%d] error: %v", i, errs[i])
			continue
		}
		if !results[i].Equal(id) {
			t.Errorf("Parse[%d] mismatch", i)
		}
	}
}

func TestConcurrentMixedOperations(t *testing.T) {
	const n = 50
	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id, err := New(testUserID{OrgID: int64(i), UserSeq: int64(i * 10)})
			if err != nil {
				t.Errorf("goroutine %d New error: %v", i, err)
				return
			}
			s := id.String()
			parsed, err := Parse[testUserID](s)
			if err != nil {
				t.Errorf("goroutine %d Parse error: %v", i, err)
				return
			}
			if !id.Equal(parsed) {
				t.Errorf("goroutine %d Parse mismatch", i)
			}
		}(i)
	}
	wg.Wait()
}

// --- Text/JSON Serialization Tests ---

func TestMarshalText(t *testing.T) {
	id := MustNew(testUserID{OrgID: 42, UserSeq: 1001})
	text, err := id.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	if string(text) != id.String() {
		t.Errorf("MarshalText() = %q, want %q", text, id.String())
	}
}

func TestMarshalTextZero(t *testing.T) {
	var id ID[testUserID]
	text, err := id.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	if string(text) != "" {
		t.Errorf("zero ID MarshalText() = %q, want %q", text, "")
	}
}

func TestUnmarshalText(t *testing.T) {
	id := MustNew(testUserID{OrgID: 42, UserSeq: 1001})
	text, _ := id.MarshalText()

	var parsed ID[testUserID]
	if err := parsed.UnmarshalText(text); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}
	if !id.Equal(parsed) {
		t.Error("text round-trip failed")
	}
}

func TestUnmarshalTextEmpty(t *testing.T) {
	var id ID[testUserID]
	if err := id.UnmarshalText([]byte{}); err != nil {
		t.Fatalf("UnmarshalText(empty): %v", err)
	}
	if !id.IsZero() {
		t.Error("UnmarshalText(empty) should produce zero ID")
	}
}

func TestUnmarshalTextInvalid(t *testing.T) {
	var id ID[testUserID]
	err := id.UnmarshalText([]byte("invalid"))
	if err == nil {
		t.Fatal("UnmarshalText with invalid input should error")
	}
}

func TestJSONRoundTrip(t *testing.T) {
	type wrapper struct {
		ID ID[testUserID] `json:"id"`
	}

	original := wrapper{ID: MustNew(testUserID{OrgID: 42, UserSeq: 1001})}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	// JSON should contain the string representation
	want := fmt.Sprintf(`{"id":%q}`, original.ID.String())
	if string(data) != want {
		t.Errorf("json.Marshal = %s, want %s", data, want)
	}

	var parsed wrapper
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if !original.ID.Equal(parsed.ID) {
		t.Error("JSON round-trip failed")
	}

	got, _ := parsed.ID.Data()
	if got != (testUserID{OrgID: 42, UserSeq: 1001}) {
		t.Errorf("Data() after JSON round-trip = %+v", got)
	}
}

func TestJSONZeroValue(t *testing.T) {
	type wrapper struct {
		ID ID[testUserID] `json:"id"`
	}

	original := wrapper{} // zero ID
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(data) != `{"id":""}` {
		t.Errorf("json.Marshal(zero) = %s, want %s", data, `{"id":""}`)
	}

	var parsed wrapper
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if !parsed.ID.IsZero() {
		t.Error("JSON round-trip of zero ID should produce zero ID")
	}
}
