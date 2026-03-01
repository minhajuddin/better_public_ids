package bpid

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

var testSigningKey = []byte("test-secret-key-for-signing-1234")

func newTestSignedRegistry(t *testing.T, opts ...SignedRegistryOption) (*SignedRegistry, *Registry) {
	t.Helper()
	r := MustNewRegistry(WithType[testUserID]("user"), WithType[testPostID]("post"))
	sr, err := NewSignedRegistry(r, testSigningKey, opts...)
	if err != nil {
		t.Fatalf("NewSignedRegistry: %v", err)
	}
	return sr, r
}

// --- NewSignedRegistry Tests ---

func TestNewSignedRegistry(t *testing.T) {
	sr, _ := newTestSignedRegistry(t)
	if sr == nil {
		t.Fatal("NewSignedRegistry returned nil")
	}
	if got := sr.Separator(); got != "." {
		t.Errorf("Separator() = %q, want %q", got, ".")
	}
}

func TestNewSignedRegistryErrors(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))

	tests := []struct {
		name       string
		registry   *Registry
		signingKey []byte
	}{
		{name: "nil registry", registry: nil, signingKey: testSigningKey},
		{name: "empty key", registry: r, signingKey: nil},
		{name: "zero-length key", registry: r, signingKey: []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSignedRegistry(tt.registry, tt.signingKey)
			if err == nil {
				t.Fatal("expected error")
			}
			if !errors.Is(err, ErrInvalidKey) {
				t.Errorf("error = %v, want ErrInvalidKey", err)
			}
		})
	}
}

func TestMustNewSignedRegistryPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNewSignedRegistry with empty key should panic")
		}
	}()
	r := MustNewRegistry(WithType[testUserID]("user"))
	MustNewSignedRegistry(r, nil)
}

// --- Serialize / Deserialize Tests ---

func TestSignedSerializeDeserializeRoundTrip(t *testing.T) {
	sr, _ := newTestSignedRegistry(t)
	data := testUserID{OrgID: 42, UserSeq: 1001}

	s, err := SignedSerialize(sr, data)
	if err != nil {
		t.Fatalf("SignedSerialize: %v", err)
	}
	if !strings.HasPrefix(s, "user.") {
		t.Errorf("SignedSerialize() = %q, want prefix 'user.'", s)
	}

	got, err := SignedDeserialize[testUserID](sr, s)
	if err != nil {
		t.Fatalf("SignedDeserialize: %v", err)
	}
	if got != data {
		t.Errorf("round-trip: got %+v, want %+v", got, data)
	}
}

func TestSignedSerializeDeterministic(t *testing.T) {
	sr, _ := newTestSignedRegistry(t)
	data := testUserID{OrgID: 42, UserSeq: 1001}

	s1, _ := SignedSerialize(sr, data)
	s2, _ := SignedSerialize(sr, data)
	if s1 != s2 {
		t.Errorf("SignedSerialize not deterministic: %q != %q", s1, s2)
	}
}

// --- Tampering Tests ---

func TestSignedDeserializeTamperedData(t *testing.T) {
	sr, _ := newTestSignedRegistry(t)
	s, _ := SignedSerialize(sr, testUserID{OrgID: 42, UserSeq: 1001})

	// Find the data portion (between first and last separator) and flip a character.
	sep := sr.Separator()
	lastSep := strings.LastIndex(s, sep)
	firstSep := strings.Index(s, sep)
	if firstSep == lastSep {
		t.Fatal("expected at least two separators in signed string")
	}

	// Tamper with a byte in the data portion.
	dataStart := firstSep + len(sep)
	tampered := []byte(s)
	tampered[dataStart] ^= 0x01
	tamperedStr := string(tampered)

	_, err := SignedDeserialize[testUserID](sr, tamperedStr)
	if err == nil {
		t.Fatal("expected error for tampered data")
	}
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("error = %v, want ErrInvalidSignature", err)
	}
}

func TestSignedDeserializeTamperedSignature(t *testing.T) {
	sr, _ := newTestSignedRegistry(t)
	s, _ := SignedSerialize(sr, testUserID{OrgID: 42, UserSeq: 1001})

	// Flip the last character of the signature.
	tampered := []byte(s)
	tampered[len(tampered)-1] ^= 0x01
	tamperedStr := string(tampered)

	_, err := SignedDeserialize[testUserID](sr, tamperedStr)
	if err == nil {
		t.Fatal("expected error for tampered signature")
	}
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("error = %v, want ErrInvalidSignature", err)
	}
}

func TestSignedDeserializeTamperedPrefix(t *testing.T) {
	sr, _ := newTestSignedRegistry(t)
	s, _ := SignedSerialize(sr, testUserID{OrgID: 42, UserSeq: 1001})

	// Swap "user" → "post" in the prefix.
	tamperedStr := "post" + s[4:]

	_, err := SignedDeserialize[testUserID](sr, tamperedStr)
	if err == nil {
		t.Fatal("expected error for tampered prefix")
	}
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("error = %v, want ErrInvalidSignature", err)
	}
}

func TestSignedDeserializeWrongKey(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))

	srA := MustNewSignedRegistry(r, []byte("key-A-secret"))
	srB := MustNewSignedRegistry(r, []byte("key-B-secret"))

	s, _ := SignedSerialize(srA, testUserID{OrgID: 42, UserSeq: 1001})

	_, err := SignedDeserialize[testUserID](srB, s)
	if err == nil {
		t.Fatal("expected error for wrong key")
	}
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("error = %v, want ErrInvalidSignature", err)
	}
}

func TestSignedDeserializeUnsignedString(t *testing.T) {
	sr, r := newTestSignedRegistry(t)

	// Serialize without signing.
	unsigned, _ := Serialize(r, testUserID{OrgID: 42, UserSeq: 1001})

	_, err := SignedDeserialize[testUserID](sr, unsigned)
	if err == nil {
		t.Fatal("expected error for unsigned string")
	}
	// The unsigned string won't have a valid HMAC at the end, so it should fail.
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("error = %v, want ErrInvalidSignature", err)
	}
}

// --- Prefix Tests ---

func TestSignedRegistryPrefix(t *testing.T) {
	sr, _ := newTestSignedRegistry(t)
	s, _ := SignedSerialize(sr, testUserID{OrgID: 42, UserSeq: 1001})

	prefix, err := sr.Prefix(s)
	if err != nil {
		t.Fatalf("Prefix: %v", err)
	}
	if prefix != "user" {
		t.Errorf("Prefix() = %q, want %q", prefix, "user")
	}
}

func TestSignedRegistryPrefixTampered(t *testing.T) {
	sr, _ := newTestSignedRegistry(t)
	s, _ := SignedSerialize(sr, testUserID{OrgID: 42, UserSeq: 1001})

	// Tamper with the signature.
	tampered := []byte(s)
	tampered[len(tampered)-1] ^= 0x01

	_, err := sr.Prefix(string(tampered))
	if err == nil {
		t.Fatal("expected error for tampered string")
	}
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("error = %v, want ErrInvalidSignature", err)
	}
}

// --- Multiple Types ---

func TestSignedRoundTripMultipleTypes(t *testing.T) {
	sr, _ := newTestSignedRegistry(t)

	userData := testUserID{OrgID: 1, UserSeq: 2}
	postData := testPostID{BoardID: 3, PostSeq: 4}

	userS, err := SignedSerialize(sr, userData)
	if err != nil {
		t.Fatalf("SignedSerialize user: %v", err)
	}
	postS, err := SignedSerialize(sr, postData)
	if err != nil {
		t.Fatalf("SignedSerialize post: %v", err)
	}

	gotUser, err := SignedDeserialize[testUserID](sr, userS)
	if err != nil {
		t.Fatalf("SignedDeserialize user: %v", err)
	}
	if gotUser != userData {
		t.Errorf("user round-trip: got %+v, want %+v", gotUser, userData)
	}

	gotPost, err := SignedDeserialize[testPostID](sr, postS)
	if err != nil {
		t.Fatalf("SignedDeserialize post: %v", err)
	}
	if gotPost != postData {
		t.Errorf("post round-trip: got %+v, want %+v", gotPost, postData)
	}
}

// --- Custom Separator ---

func TestSignedRoundTripCustomSeparator(t *testing.T) {
	r := MustNewRegistry(WithSeparator("~"), WithType[testUserID]("user"))
	sr := MustNewSignedRegistry(r, testSigningKey)
	data := testUserID{OrgID: 42, UserSeq: 1001}

	s, err := SignedSerialize(sr, data)
	if err != nil {
		t.Fatalf("SignedSerialize: %v", err)
	}
	if !strings.HasPrefix(s, "user~") {
		t.Errorf("SignedSerialize() = %q, want prefix 'user~'", s)
	}
	if sr.Separator() != "~" {
		t.Errorf("Separator() = %q, want %q", sr.Separator(), "~")
	}

	got, err := SignedDeserialize[testUserID](sr, s)
	if err != nil {
		t.Fatalf("SignedDeserialize: %v", err)
	}
	if got != data {
		t.Errorf("round-trip with ~ separator: got %+v, want %+v", got, data)
	}
}

// --- Concurrency ---

func TestConcurrentSignedOperations(t *testing.T) {
	sr, _ := newTestSignedRegistry(t)
	const n = 50
	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			data := testUserID{OrgID: int64(i), UserSeq: int64(i * 10)}
			s, err := SignedSerialize(sr, data)
			if err != nil {
				t.Errorf("goroutine %d SignedSerialize error: %v", i, err)
				return
			}
			got, err := SignedDeserialize[testUserID](sr, s)
			if err != nil {
				t.Errorf("goroutine %d SignedDeserialize error: %v", i, err)
				return
			}
			if got != data {
				t.Errorf("goroutine %d round-trip mismatch: got %+v, want %+v", i, got, data)
			}
		}(i)
	}
	wg.Wait()
}

// --- Key Rotation Tests ---

func TestKeyRotation(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	keyA := []byte("key-A-original-secret")
	keyB := []byte("key-B-rotated-secret")

	// Phase 1: sign with keyA.
	srA := MustNewSignedRegistry(r, keyA)
	oldID, err := SignedSerialize(srA, testUserID{OrgID: 42, UserSeq: 1001})
	if err != nil {
		t.Fatalf("SignedSerialize with keyA: %v", err)
	}

	// Phase 2: rotate to keyB, keep keyA for verification.
	srB := MustNewSignedRegistry(r, keyB, WithOldKeys(keyA))

	// Old IDs signed with keyA should still verify.
	got, err := SignedDeserialize[testUserID](srB, oldID)
	if err != nil {
		t.Fatalf("SignedDeserialize old ID after rotation: %v", err)
	}
	if got != (testUserID{OrgID: 42, UserSeq: 1001}) {
		t.Errorf("round-trip after rotation: got %+v, want {42, 1001}", got)
	}

	// New IDs are signed with keyB.
	newID, err := SignedSerialize(srB, testUserID{OrgID: 99, UserSeq: 2002})
	if err != nil {
		t.Fatalf("SignedSerialize with keyB: %v", err)
	}
	got2, err := SignedDeserialize[testUserID](srB, newID)
	if err != nil {
		t.Fatalf("SignedDeserialize new ID: %v", err)
	}
	if got2 != (testUserID{OrgID: 99, UserSeq: 2002}) {
		t.Errorf("round-trip new ID: got %+v, want {99, 2002}", got2)
	}
}

func TestKeyRotationDroppedKey(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	keyA := []byte("key-A-original-secret")
	keyB := []byte("key-B-rotated-secret")

	// Sign with keyA.
	srA := MustNewSignedRegistry(r, keyA)
	oldID, _ := SignedSerialize(srA, testUserID{OrgID: 42, UserSeq: 1001})

	// Phase 3: drop keyA entirely.
	srB := MustNewSignedRegistry(r, keyB)

	_, err := SignedDeserialize[testUserID](srB, oldID)
	if err == nil {
		t.Fatal("expected error for dropped key")
	}
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("error = %v, want ErrInvalidSignature", err)
	}
}

func TestWithOldKeysEmptyKeyError(t *testing.T) {
	r := MustNewRegistry(WithType[testUserID]("user"))
	_, err := NewSignedRegistry(r, testSigningKey, WithOldKeys([]byte{}))
	if err == nil {
		t.Fatal("expected error for empty old key")
	}
	if !errors.Is(err, ErrInvalidKey) {
		t.Errorf("error = %v, want ErrInvalidKey", err)
	}
}

// --- Benchmarks ---

var benchSignedRegistry = MustNewSignedRegistry(
	MustNewRegistry(WithType[benchUserID]("benchuser")),
	[]byte("bench-signing-key-1234567890"),
)

func BenchmarkSignedSerialize(b *testing.B) {
	for b.Loop() {
		_, _ = SignedSerialize(benchSignedRegistry, benchUserID{OrgID: 42, UserSeq: 1001})
	}
}

func BenchmarkSignedDeserialize(b *testing.B) {
	s := MustSignedSerialize(benchSignedRegistry, benchUserID{OrgID: 42, UserSeq: 1001})
	b.ResetTimer()
	for b.Loop() {
		_, _ = SignedDeserialize[benchUserID](benchSignedRegistry, s)
	}
}
