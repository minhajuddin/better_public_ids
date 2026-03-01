package bpid

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// Test Definer types (data-carrying structs)
type userIDDef struct {
	OrgID   int64
	UserSeq int64
}

func (userIDDef) Prefix() string { return "user" }

type postIDDef struct {
	BoardID int64
	PostSeq int64
}

func (postIDDef) Prefix() string { return "post" }

// Compile-time interface compliance checks
var (
	_ fmt.Stringer               = ID[userIDDef]{}
	_ encoding.TextMarshaler     = ID[userIDDef]{}
	_ encoding.TextUnmarshaler   = (*ID[userIDDef])(nil)
	_ json.Marshaler             = ID[userIDDef]{}
	_ json.Unmarshaler           = (*ID[userIDDef])(nil)
	_ encoding.BinaryMarshaler   = ID[userIDDef]{}
	_ encoding.BinaryUnmarshaler = (*ID[userIDDef])(nil)
	_ driver.Valuer              = ID[userIDDef]{}
	_ sql.Scanner                = (*ID[userIDDef])(nil)
)

// --- Constructor Tests ---

func TestNew(t *testing.T) {
	data := userIDDef{OrgID: 42, UserSeq: 1001}
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
	id1, _ := New(userIDDef{OrgID: 1, UserSeq: 1})
	id2, _ := New(userIDDef{OrgID: 1, UserSeq: 2})

	if id1.Equal(id2) {
		t.Error("IDs with different data should not be equal")
	}
}

func TestNewSameData(t *testing.T) {
	data := userIDDef{OrgID: 42, UserSeq: 1001}
	id1, _ := New(data)
	id2, _ := New(data)

	if !id1.Equal(id2) {
		t.Error("IDs with same data should be equal (gob is deterministic)")
	}
}

func TestMustNew(t *testing.T) {
	data := userIDDef{OrgID: 42, UserSeq: 1001}
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
	data := userIDDef{OrgID: 42, UserSeq: 1001}
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
		{name: "only prefix", input: "user.", wantErr: ErrDecodingFailed},
		{name: "whitespace prefix", input: " user." + encodeBytes(validID.raw), wantErr: ErrPrefixMismatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := Parse[userIDDef](tt.input)
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
	data := userIDDef{OrgID: 42, UserSeq: 1001}
	id, _ := New(data)
	s := id.String()

	parsed, err := Parse[userIDDef](s)
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
	id, _ := New(userIDDef{OrgID: 42, UserSeq: 1001})
	s := id.String()

	parsed := MustParse[userIDDef](s)
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
	MustParse[userIDDef]("invalid")
}

// --- Method Tests ---

func TestString(t *testing.T) {
	id, _ := New(userIDDef{OrgID: 42, UserSeq: 1001})
	s := id.String()

	if !strings.HasPrefix(s, "user.") {
		t.Errorf("String() = %q, want prefix 'user.'", s)
	}
	if len(s) <= len("user.") {
		t.Errorf("String() = %q, encoded part is empty", s)
	}
}

func TestStringZero(t *testing.T) {
	var id ID[userIDDef]
	if got := id.String(); got != "" {
		t.Errorf("zero ID String() = %q, want %q", got, "")
	}
}

func TestStringDeterministic(t *testing.T) {
	id, _ := New(userIDDef{OrgID: 42, UserSeq: 1001})
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
	data := userIDDef{OrgID: 42, UserSeq: 1001}
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
	var id ID[userIDDef]
	got, err := id.Data()
	if err != nil {
		t.Fatalf("Data on zero ID: %v", err)
	}
	if got != (userIDDef{}) {
		t.Errorf("zero ID Data() = %+v, want zero value", got)
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		name string
		id   ID[userIDDef]
		want bool
	}{
		{name: "zero value", id: ID[userIDDef]{}, want: true},
		{name: "new ID", id: MustNew(userIDDef{OrgID: 1, UserSeq: 1}), want: false},
		{name: "new with zero data", id: MustNew(userIDDef{}), want: false},
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
	data := userIDDef{OrgID: 42, UserSeq: 1001}
	id1, _ := New(data)
	id2, _ := New(data)
	id3, _ := New(userIDDef{OrgID: 99, UserSeq: 2002})

	if !id1.Equal(id2) {
		t.Error("same data IDs should be equal")
	}
	if id1.Equal(id3) {
		t.Error("different data IDs should not be equal")
	}

	var z1, z2 ID[userIDDef]
	if !z1.Equal(z2) {
		t.Error("two zero IDs should be equal")
	}
}

func TestPrefix(t *testing.T) {
	id := MustNew(userIDDef{OrgID: 1, UserSeq: 1})
	if id.Prefix() != "user" {
		t.Errorf("Prefix() = %q, want %q", id.Prefix(), "user")
	}

	pid := MustNew(postIDDef{BoardID: 1, PostSeq: 1})
	if pid.Prefix() != "post" {
		t.Errorf("Prefix() = %q, want %q", pid.Prefix(), "post")
	}

	var zero ID[userIDDef]
	if zero.Prefix() != "user" {
		t.Errorf("zero Prefix() = %q, want %q", zero.Prefix(), "user")
	}
}

// --- Text Marshaling Tests ---

func TestTextMarshalRoundTrip(t *testing.T) {
	id := MustNew(userIDDef{OrgID: 42, UserSeq: 1001})

	data, err := id.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}

	var parsed ID[userIDDef]
	if err := parsed.UnmarshalText(data); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}

	if !id.Equal(parsed) {
		t.Error("text marshal round-trip failed")
	}
}

func TestTextMarshalZero(t *testing.T) {
	var id ID[userIDDef]

	data, err := id.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("zero ID MarshalText = %q, want empty", string(data))
	}

	var parsed ID[userIDDef]
	if err := parsed.UnmarshalText(data); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}
	if !parsed.IsZero() {
		t.Error("UnmarshalText(empty) should produce zero ID")
	}
}

// --- JSON Marshaling Tests ---

func TestJSONMarshalRoundTrip(t *testing.T) {
	id := MustNew(userIDDef{OrgID: 42, UserSeq: 1001})

	data, err := id.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	var parsed ID[userIDDef]
	if err := parsed.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}

	if !id.Equal(parsed) {
		t.Error("JSON marshal round-trip failed")
	}
}

func TestJSONMarshalZero(t *testing.T) {
	var id ID[userIDDef]

	data, err := id.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("zero ID MarshalJSON = %q, want %q", string(data), "null")
	}

	var parsed ID[userIDDef]
	if err := parsed.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON(null): %v", err)
	}
	if !parsed.IsZero() {
		t.Error("UnmarshalJSON(null) should produce zero ID")
	}
}

func TestJSONUnmarshalEmptyString(t *testing.T) {
	var id ID[userIDDef]
	if err := id.UnmarshalJSON([]byte(`""`)); err != nil {
		t.Fatalf("UnmarshalJSON empty string: %v", err)
	}
	if !id.IsZero() {
		t.Error("UnmarshalJSON(\"\") should produce zero ID")
	}
}

func TestJSONUnmarshalInvalidJSON(t *testing.T) {
	var id ID[userIDDef]
	err := id.UnmarshalJSON([]byte(`{}`))
	if err == nil {
		t.Fatal("UnmarshalJSON({}) should error")
	}
}

func TestJSONUnmarshalInvalidID(t *testing.T) {
	postID := MustNew(postIDDef{BoardID: 1, PostSeq: 1})
	data, _ := json.Marshal(postID.String())

	var id ID[userIDDef]
	err := id.UnmarshalJSON(data)
	if err == nil {
		t.Fatal("UnmarshalJSON with wrong prefix should error")
	}
	if !errors.Is(err, ErrPrefixMismatch) {
		t.Fatalf("error = %v, want ErrPrefixMismatch", err)
	}
}

func TestJSONInStruct(t *testing.T) {
	type Payload struct {
		ID   ID[userIDDef] `json:"id"`
		Name string        `json:"name"`
	}

	id := MustNew(userIDDef{OrgID: 42, UserSeq: 1001})
	original := Payload{ID: id, Name: "Alice"}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var decoded Payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if !original.ID.Equal(decoded.ID) {
		t.Error("JSON struct round-trip ID mismatch")
	}
	if original.Name != decoded.Name {
		t.Error("JSON struct round-trip Name mismatch")
	}
}

func TestJSONInStructZeroID(t *testing.T) {
	type Payload struct {
		ID   ID[userIDDef] `json:"id"`
		Name string        `json:"name"`
	}

	original := Payload{Name: "Bob"}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	if !strings.Contains(string(data), `"id":null`) {
		t.Errorf("expected null ID in JSON, got %s", string(data))
	}

	var decoded Payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if !decoded.ID.IsZero() {
		t.Error("decoded ID should be zero")
	}
}

func TestJSONInStructOmitzero(t *testing.T) {
	type Payload struct {
		ID   ID[userIDDef] `json:"id,omitzero"`
		Name string        `json:"name"`
	}

	original := Payload{Name: "Charlie"}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	if strings.Contains(string(data), `"id"`) {
		t.Errorf("expected omitted ID in JSON, got %s", string(data))
	}
}

func TestJSONNullInPayload(t *testing.T) {
	type Payload struct {
		ID   ID[userIDDef] `json:"id"`
		Name string        `json:"name"`
	}

	data := []byte(`{"id":null,"name":"Dave"}`)
	var decoded Payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if !decoded.ID.IsZero() {
		t.Error("null ID should decode to zero")
	}
	if decoded.Name != "Dave" {
		t.Errorf("Name = %q, want %q", decoded.Name, "Dave")
	}
}

// --- Binary Marshaling Tests ---

func TestBinaryMarshalRoundTrip(t *testing.T) {
	id := MustNew(userIDDef{OrgID: 42, UserSeq: 1001})

	data, err := id.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("MarshalBinary returned empty bytes for non-zero ID")
	}

	var parsed ID[userIDDef]
	if err := parsed.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}

	if !id.Equal(parsed) {
		t.Error("binary marshal round-trip failed")
	}
}

func TestBinaryMarshalZero(t *testing.T) {
	var id ID[userIDDef]

	data, err := id.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("zero ID MarshalBinary returned %d bytes, want 0", len(data))
	}

	var parsed ID[userIDDef]
	if err := parsed.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	if !parsed.IsZero() {
		t.Error("UnmarshalBinary(nil) should produce zero ID")
	}
}

func TestBinaryUnmarshalInvalidGob(t *testing.T) {
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
			var id ID[userIDDef]
			err := id.UnmarshalBinary(tt.data)
			if err == nil {
				t.Error("UnmarshalBinary should error on invalid gob bytes")
			}
			if !errors.Is(err, ErrDecodingFailed) {
				t.Errorf("error = %v, want ErrDecodingFailed", err)
			}
		})
	}
}

// --- SQL Tests ---

func TestSQLValueAndScan(t *testing.T) {
	id := MustNew(userIDDef{OrgID: 42, UserSeq: 1001})

	val, err := id.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}

	s, ok := val.(string)
	if !ok {
		t.Fatalf("Value() type = %T, want string", val)
	}
	if s != id.String() {
		t.Errorf("Value() = %q, want %q", s, id.String())
	}

	var scanned ID[userIDDef]
	if err := scanned.Scan(s); err != nil {
		t.Fatalf("Scan(string): %v", err)
	}
	if !id.Equal(scanned) {
		t.Error("Value/Scan round-trip failed")
	}
}

func TestSQLValueZero(t *testing.T) {
	var id ID[userIDDef]
	val, err := id.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if val != nil {
		t.Errorf("zero ID Value() = %v, want nil", val)
	}
}

func TestSQLScanNil(t *testing.T) {
	id := MustNew(userIDDef{OrgID: 1, UserSeq: 1})
	if err := id.Scan(nil); err != nil {
		t.Fatalf("Scan(nil): %v", err)
	}
	if !id.IsZero() {
		t.Error("Scan(nil) should produce zero ID")
	}
}

func TestSQLScanString(t *testing.T) {
	original := MustNew(userIDDef{OrgID: 42, UserSeq: 1001})
	s := original.String()

	var id ID[userIDDef]
	if err := id.Scan(s); err != nil {
		t.Fatalf("Scan(string): %v", err)
	}
	if !original.Equal(id) {
		t.Error("Scan(string) mismatch")
	}
}

func TestSQLScanBytes(t *testing.T) {
	original := MustNew(userIDDef{OrgID: 42, UserSeq: 1001})
	s := original.String()

	var id ID[userIDDef]
	if err := id.Scan([]byte(s)); err != nil {
		t.Fatalf("Scan([]byte): %v", err)
	}
	if !original.Equal(id) {
		t.Error("Scan([]byte) mismatch")
	}
}

func TestSQLScanEmptyString(t *testing.T) {
	var id ID[userIDDef]
	if err := id.Scan(""); err != nil {
		t.Fatalf("Scan empty string: %v", err)
	}
	if !id.IsZero() {
		t.Error("Scan(\"\") should produce zero ID")
	}
}

func TestSQLScanEmptyBytes(t *testing.T) {
	var id ID[userIDDef]
	if err := id.Scan([]byte{}); err != nil {
		t.Fatalf("Scan empty bytes: %v", err)
	}
	if !id.IsZero() {
		t.Error("Scan([]byte{}) should produce zero ID")
	}
}

func TestSQLScanUnsupportedType(t *testing.T) {
	var id ID[userIDDef]
	err := id.Scan(123)
	if err == nil {
		t.Fatal("Scan(int) should error")
	}
	if !errors.Is(err, ErrScanType) {
		t.Errorf("error = %v, want ErrScanType", err)
	}
}

func TestSQLScanInvalidID(t *testing.T) {
	postID := MustNew(postIDDef{BoardID: 1, PostSeq: 1})
	var id ID[userIDDef]
	err := id.Scan(postID.String())
	if err == nil {
		t.Fatal("Scan with wrong prefix should error")
	}
	if !errors.Is(err, ErrPrefixMismatch) {
		t.Errorf("error = %v, want ErrPrefixMismatch", err)
	}
}

// --- Cross-Type Safety Tests ---

func TestParseWrongType(t *testing.T) {
	postID := MustNew(postIDDef{BoardID: 1, PostSeq: 1})
	postStr := postID.String()

	_, err := Parse[userIDDef](postStr)
	if err == nil {
		t.Fatal("Parse[userIDDef] with postID string should error")
	}
	if !errors.Is(err, ErrPrefixMismatch) {
		t.Errorf("error = %v, want ErrPrefixMismatch", err)
	}
}

func TestMultipleTypes(t *testing.T) {
	userID := MustNew(userIDDef{OrgID: 1, UserSeq: 1})
	postID := MustNew(postIDDef{BoardID: 1, PostSeq: 1})

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

// --- Auto-Registration Tests ---

type autoRegTestDef struct {
	Val int64
}

func (autoRegTestDef) Prefix() string { return "autoreg" }

func TestAutoRegistration(t *testing.T) {
	freshReg := NewRegistry()

	_, _, err := freshReg.ParseAny("autoreg.AAAAAAAAAAAAAAAAAAAAAA")
	if !errors.Is(err, ErrUnknownPrefix) {
		t.Fatalf("expected ErrUnknownPrefix on fresh registry, got %v", err)
	}

	id := MustNew(autoRegTestDef{Val: 42})

	prefix, _, err := DefaultRegistry.ParseAny(id.String())
	if err != nil {
		t.Fatalf("ParseAny after auto-registration: %v", err)
	}
	if prefix != "autoreg" {
		t.Errorf("prefix = %q, want %q", prefix, "autoreg")
	}
}

// --- Zero Value Tests ---

func TestZeroValueBehavior(t *testing.T) {
	var id ID[userIDDef]

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
	if data != (userIDDef{}) {
		t.Errorf("zero value Data() = %+v, want zero", data)
	}

	var other ID[userIDDef]
	if !id.Equal(other) {
		t.Error("two zero values should be equal")
	}
}

// --- Concurrency Tests ---

func TestConcurrentNew(t *testing.T) {
	const n = 100
	ids := make([]ID[userIDDef], n)
	errs := make([]error, n)
	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ids[i], errs[i] = New(userIDDef{OrgID: int64(i), UserSeq: int64(i * 10)})
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
	id := MustNew(userIDDef{OrgID: 42, UserSeq: 1001})
	s := id.String()

	const n = 100
	results := make([]ID[userIDDef], n)
	errs := make([]error, n)
	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = Parse[userIDDef](s)
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
			id, err := New(userIDDef{OrgID: int64(i), UserSeq: int64(i * 10)})
			if err != nil {
				t.Errorf("goroutine %d New error: %v", i, err)
				return
			}
			s := id.String()
			parsed, err := Parse[userIDDef](s)
			if err != nil {
				t.Errorf("goroutine %d Parse error: %v", i, err)
				return
			}
			if !id.Equal(parsed) {
				t.Errorf("goroutine %d Parse mismatch", i)
			}
			data, err := id.MarshalJSON()
			if err != nil {
				t.Errorf("goroutine %d MarshalJSON error: %v", i, err)
				return
			}
			var unmarshaled ID[userIDDef]
			if err := unmarshaled.UnmarshalJSON(data); err != nil {
				t.Errorf("goroutine %d UnmarshalJSON error: %v", i, err)
				return
			}
			if !id.Equal(unmarshaled) {
				t.Errorf("goroutine %d JSON round-trip mismatch", i)
			}
		}(i)
	}
	wg.Wait()
}
