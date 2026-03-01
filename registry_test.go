package bpid

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

type testRegID struct {
	Val int64
}

type testUserID struct {
	OrgID   int64
	UserSeq int64
}

type testPostID struct {
	BoardID int64
	PostSeq int64
}

// --- NewRegistry Tests ---

func TestNewRegistry(t *testing.T) {
	r, err := NewRegistry()
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	if got := r.Separator(); got != "." {
		t.Errorf("NewRegistry().Separator() = %q, want %q", got, ".")
	}
}

func TestNewRegistryWithSeparator(t *testing.T) {
	r, err := NewRegistry(WithSeparator("~"))
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	if got := r.Separator(); got != "~" {
		t.Errorf("Separator() = %q, want %q", got, "~")
	}

	r2, err := NewRegistry(WithSeparator("."))
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	if got := r2.Separator(); got != "." {
		t.Errorf("Separator() = %q, want %q", got, ".")
	}
}

func TestWithSeparatorErrors(t *testing.T) {
	invalids := []string{":", "", "ab", " ", "-", "_", "/", "\\"}
	for _, sep := range invalids {
		t.Run(fmt.Sprintf("sep=%q", sep), func(t *testing.T) {
			_, err := NewRegistry(WithSeparator(sep))
			if err == nil {
				t.Errorf("NewRegistry(WithSeparator(%q)) should error", sep)
			}
			if !errors.Is(err, ErrInvalidSeparator) {
				t.Errorf("error = %v, want ErrInvalidSeparator", err)
			}
		})
	}
}

func TestMustNewRegistry(t *testing.T) {
	r := MustNewRegistry(WithSeparator("~"))
	if got := r.Separator(); got != "~" {
		t.Errorf("Separator() = %q, want %q", got, "~")
	}
}

func TestMustNewRegistryPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNewRegistry with invalid separator should panic")
		}
	}()
	MustNewRegistry(WithSeparator(":"))
}

// --- WithType Tests ---

func TestWithType(t *testing.T) {
	_, err := NewRegistry(WithType[testRegID]("regtest"))
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
}

func TestWithTypeDuplicatePrefix(t *testing.T) {
	_, err := NewRegistry(
		WithType[testRegID]("regtest"),
		WithType[testPostID]("regtest"),
	)
	if err == nil {
		t.Fatal("duplicate prefix should error")
	}
	if !errors.Is(err, ErrDuplicatePrefix) {
		t.Fatalf("error = %v, want ErrDuplicatePrefix", err)
	}
}

func TestWithTypeDuplicateType(t *testing.T) {
	_, err := NewRegistry(
		WithType[testRegID]("regtest"),
		WithType[testRegID]("regtest2"),
	)
	if err == nil {
		t.Fatal("duplicate type should error")
	}
	if !errors.Is(err, ErrDuplicateType) {
		t.Fatalf("error = %v, want ErrDuplicateType", err)
	}
}

// --- Prefix Validation ---

func TestWithTypeInvalidPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{name: "empty prefix", prefix: ""},
		{name: "uppercase prefix", prefix: "User"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRegistry(WithType[testRegID](tt.prefix))
			if err == nil {
				t.Fatal("expected error for invalid prefix")
			}
			if !errors.Is(err, ErrInvalidPrefix) {
				t.Errorf("error = %v, want ErrInvalidPrefix", err)
			}
		})
	}
}

// --- Serialize Tests ---

func TestSerialize(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	data := testUserID{OrgID: 42, UserSeq: 1001}

	s, err := Serialize(r, data)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !strings.HasPrefix(s, "user.") {
		t.Errorf("Serialize() = %q, want prefix 'user.'", s)
	}
	if len(s) <= len("user.") {
		t.Errorf("Serialize() = %q, encoded part is empty", s)
	}
}

func TestSerializeDeterministic(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	data := testUserID{OrgID: 42, UserSeq: 1001}

	s1, _ := Serialize(r, data)
	s2, _ := Serialize(r, data)
	if s1 != s2 {
		t.Errorf("Serialize not deterministic: %q != %q", s1, s2)
	}
}

func TestSerializeUnregistered(t *testing.T) {
	r := MustNewRegistry(WithType[testPostID]("post"))
	_, err := Serialize(r, testUserID{OrgID: 1, UserSeq: 1})
	if err == nil {
		t.Fatal("Serialize with unregistered type should error")
	}
	if !errors.Is(err, ErrUnregisteredPrefix) {
		t.Errorf("error = %v, want ErrUnregisteredPrefix", err)
	}
}

func TestMustSerialize(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	data := testUserID{OrgID: 42, UserSeq: 1001}
	s := MustSerialize(r, data)
	if !strings.HasPrefix(s, "user.") {
		t.Errorf("MustSerialize() = %q, want prefix 'user.'", s)
	}
}

func TestMustSerializePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustSerialize with unregistered type should panic")
		}
	}()
	r := MustNewRegistry()
	MustSerialize(r, testUserID{OrgID: 1, UserSeq: 1})
}

// --- Deserialize Tests ---

func TestDeserialize(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	data := testUserID{OrgID: 42, UserSeq: 1001}
	s, _ := Serialize(r, data)

	got, err := Deserialize[testUserID](r, s)
	if err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got != data {
		t.Errorf("Deserialize() = %+v, want %+v", got, data)
	}
}

func TestDeserializeErrors(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"), WithType[testPostID]("post"))
	validS, _ := Serialize(r, testUserID{OrgID: 42, UserSeq: 1001})

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "empty string", input: "", wantErr: ErrEmptyString},
		{name: "no separator", input: "userABCDEF", wantErr: ErrInvalidFormat},
		{name: "wrong prefix", input: strings.Replace(validS, "user.", "post.", 1), wantErr: ErrPrefixMismatch},
		{name: "invalid encoding", input: "user.!!invalid!!", wantErr: ErrInvalidEncoding},
		{name: "only prefix", input: "user.", wantErr: ErrInvalidFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Deserialize[testUserID](r, tt.input)
			if err == nil {
				t.Fatalf("Deserialize(%q) = nil error, want %v", tt.input, tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Deserialize(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestDeserializeUnregistered(t *testing.T) {
	r := MustNewRegistry(WithType[testPostID]("post"))
	_, err := Deserialize[testUserID](r, "user.AAAA")
	if err == nil {
		t.Fatal("Deserialize with unregistered type should error")
	}
	if !errors.Is(err, ErrUnregisteredPrefix) {
		t.Errorf("error = %v, want ErrUnregisteredPrefix", err)
	}
}

// --- Round-trip Tests ---

func TestSerializeDeserializeRoundTrip(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	data := testUserID{OrgID: 42, UserSeq: 1001}

	s, err := Serialize(r, data)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	got, err := Deserialize[testUserID](r, s)
	if err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got != data {
		t.Errorf("round-trip: got %+v, want %+v", got, data)
	}
}

func TestRoundTripMultipleTypes(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"), WithType[testPostID]("post"))

	userData := testUserID{OrgID: 1, UserSeq: 2}
	postData := testPostID{BoardID: 3, PostSeq: 4}

	userS, _ := Serialize(r, userData)
	postS, _ := Serialize(r, postData)

	if strings.HasPrefix(userS, "post.") {
		t.Error("userS should not have post prefix")
	}
	if strings.HasPrefix(postS, "user.") {
		t.Error("postS should not have user prefix")
	}

	gotUser, err := Deserialize[testUserID](r, userS)
	if err != nil {
		t.Fatalf("Deserialize user: %v", err)
	}
	if gotUser != userData {
		t.Errorf("user round-trip: got %+v, want %+v", gotUser, userData)
	}

	gotPost, err := Deserialize[testPostID](r, postS)
	if err != nil {
		t.Fatalf("Deserialize post: %v", err)
	}
	if gotPost != postData {
		t.Errorf("post round-trip: got %+v, want %+v", gotPost, postData)
	}
}

// --- Registry.Prefix Tests ---

func TestRegistryPrefix(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"), WithType[testPostID]("post"))
	s, _ := Serialize(r, testUserID{OrgID: 42, UserSeq: 1001})

	prefix, err := r.Prefix(s)
	if err != nil {
		t.Fatalf("Prefix: %v", err)
	}
	if prefix != "user" {
		t.Errorf("Prefix() = %q, want %q", prefix, "user")
	}
}

func TestRegistryPrefixErrors(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "empty string", input: "", wantErr: ErrEmptyString},
		{name: "no separator", input: "userABCDEF", wantErr: ErrInvalidFormat},
		{name: "unregistered", input: "unknown.ABCDEF", wantErr: ErrUnregisteredPrefix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.Prefix(tt.input)
			if err == nil {
				t.Fatalf("Prefix(%q) = nil error, want %v", tt.input, tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Prefix(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// --- Custom Separator Tests ---

func TestSerializeDeserializeCustomSeparator(t *testing.T) {
	r := MustNewRegistry(WithSeparator("~"), WithType[testUserID]("user"))
	data := testUserID{OrgID: 42, UserSeq: 1001}

	s, err := Serialize(r, data)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !strings.HasPrefix(s, "user~") {
		t.Errorf("Serialize() = %q, want prefix 'user~'", s)
	}

	got, err := Deserialize[testUserID](r, s)
	if err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got != data {
		t.Errorf("round-trip with ~ separator: got %+v, want %+v", got, data)
	}

	// Prefix extraction should work with ~
	prefix, err := r.Prefix(s)
	if err != nil {
		t.Fatalf("Prefix: %v", err)
	}
	if prefix != "user" {
		t.Errorf("Prefix() = %q, want %q", prefix, "user")
	}
}

func TestDeserializeWrongSeparator(t *testing.T) {
	rDot := MustNewRegistry(WithType[testUserID]("user"))
	rTilde := MustNewRegistry(WithSeparator("~"), WithType[testUserID]("user"))

	s, _ := Serialize(rDot, testUserID{OrgID: 42, UserSeq: 1001})

	// Try deserializing dot-serialized string with tilde registry
	_, err := Deserialize[testUserID](rTilde, s)
	if err == nil {
		t.Fatal("Deserialize with wrong separator should fail")
	}
}

// --- Concurrency Tests ---

// --- Inspect Tests ---

func TestInspectEmptyRegistry(t *testing.T) {
	r := MustNewRegistry()
	got := r.Inspect()
	want := `bpid.Registry(separator=".", types=0)`
	if got != want {
		t.Errorf("Inspect() = %q, want %q", got, want)
	}
}

func TestInspectWithTypes(t *testing.T) {
	r := MustNewRegistry(
		WithType[testUserID]("user"),
		WithType[testPostID]("post"),
	)
	got := r.Inspect()

	// Should contain key info.
	if !strings.Contains(got, "types=2") {
		t.Errorf("Inspect() = %q, want types=2", got)
	}
	if !strings.Contains(got, "user→") {
		t.Errorf("Inspect() = %q, want user→ entry", got)
	}
	if !strings.Contains(got, "post→") {
		t.Errorf("Inspect() = %q, want post→ entry", got)
	}
	// Entries should be sorted by prefix: post before user.
	postIdx := strings.Index(got, "post→")
	userIdx := strings.Index(got, "user→")
	if postIdx > userIdx {
		t.Errorf("Inspect() entries not sorted: post at %d, user at %d", postIdx, userIdx)
	}
}

func TestInspectCustomSeparator(t *testing.T) {
	r := MustNewRegistry(WithSeparator("~"))
	got := r.Inspect()
	if !strings.Contains(got, `separator="~"`) {
		t.Errorf("Inspect() = %q, want separator=\"~\"", got)
	}
}

// --- Concurrency Tests ---

func TestConcurrentSerialize(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	const n = 100
	results := make([]string, n)
	errs := make([]error, n)
	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = Serialize(r, testUserID{OrgID: int64(i), UserSeq: int64(i * 10)})
		}(i)
	}
	wg.Wait()

	seen := make(map[string]bool, n)
	for i, s := range results {
		if errs[i] != nil {
			t.Errorf("Serialize[%d] error: %v", i, errs[i])
			continue
		}
		if seen[s] {
			t.Errorf("Serialize[%d] duplicate: %s", i, s)
		}
		seen[s] = true
	}
}

func TestConcurrentDeserialize(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	data := testUserID{OrgID: 42, UserSeq: 1001}
	s, _ := Serialize(r, data)

	const n = 100
	results := make([]testUserID, n)
	errs := make([]error, n)
	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = Deserialize[testUserID](r, s)
		}(i)
	}
	wg.Wait()

	for i := range n {
		if errs[i] != nil {
			t.Errorf("Deserialize[%d] error: %v", i, errs[i])
			continue
		}
		if results[i] != data {
			t.Errorf("Deserialize[%d] = %+v, want %+v", i, results[i], data)
		}
	}
}

func TestConcurrentMixedOperations(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	const n = 50
	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			data := testUserID{OrgID: int64(i), UserSeq: int64(i * 10)}
			s, err := Serialize(r, data)
			if err != nil {
				t.Errorf("goroutine %d Serialize error: %v", i, err)
				return
			}
			got, err := Deserialize[testUserID](r, s)
			if err != nil {
				t.Errorf("goroutine %d Deserialize error: %v", i, err)
				return
			}
			if got != data {
				t.Errorf("goroutine %d round-trip mismatch: got %+v, want %+v", i, got, data)
			}
		}(i)
	}
	wg.Wait()
}

// --- Custom Codec Tests ---

// jsonCodec implements Codec using encoding/json.
type jsonCodec struct{}

func (jsonCodec) Marshal(v any) ([]byte, error)          { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v any) error      { return json.Unmarshal(data, v) }

func TestWithCodecJSON(t *testing.T) {
	r, err := NewRegistry(
		WithCodec(jsonCodec{}),
		WithType[testUserID]("user"),
	)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	data := testUserID{OrgID: 42, UserSeq: 1001}
	s, err := Serialize(r, data)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !strings.HasPrefix(s, "user.") {
		t.Errorf("Serialize() = %q, want prefix 'user.'", s)
	}

	got, err := Deserialize[testUserID](r, s)
	if err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got != data {
		t.Errorf("round-trip with JSON codec: got %+v, want %+v", got, data)
	}
}

func TestWithCodecDefaultIsGob(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	if _, ok := r.Codec().(GobCodec); !ok {
		t.Errorf("default codec = %T, want GobCodec", r.Codec())
	}
}

func TestWithCodecAccessor(t *testing.T) {
	c := jsonCodec{}
	r, err := NewRegistry(WithCodec(c), WithType[testUserID]("user"))
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	if r.Codec() != c {
		t.Errorf("Codec() returned different instance")
	}
}

func TestWithCodecNilErrors(t *testing.T) {
	_, err := NewRegistry(WithCodec(nil))
	if err == nil {
		t.Fatal("WithCodec(nil) should error")
	}
}

func TestWithCodecIncompatibleRoundTrip(t *testing.T) {
	// Serialize with JSON, try to deserialize with gob — should fail.
	rJSON, _ := NewRegistry(WithCodec(jsonCodec{}), WithType[testUserID]("user"))
	rGob := MustNewRegistry(WithType[testUserID]("user"))

	s, err := Serialize(rJSON, testUserID{OrgID: 1, UserSeq: 2})
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	_, err = Deserialize[testUserID](rGob, s)
	if err == nil {
		t.Fatal("Deserialize with mismatched codec should fail")
	}
}
