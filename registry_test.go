package bpid

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"
)

type registryTestData struct {
	Val int64
}

func (registryTestData) Prefix() string { return "regtest" }

// makeTestEncodedString creates a valid encoded ID string for registry tests.
func makeTestEncodedString(prefix string) (string, []byte) {
	data := registryTestData{Val: 42}
	raw, _ := encodeGob(data)
	return prefix + "." + encodeBytes(raw), raw
}

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

func TestWithType(t *testing.T) {
	r, err := NewRegistry(WithType[registryTestData]())
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	if !r.IsRegistered("regtest") {
		t.Error("regtest should be registered")
	}
}

func TestWithTypeDuplicate(t *testing.T) {
	_, err := NewRegistry(
		WithType[registryTestData](),
		WithType[registryTestData](),
	)
	if err == nil {
		t.Fatal("duplicate WithType should error")
	}
	if !errors.Is(err, ErrDuplicatePrefix) {
		t.Fatalf("error = %v, want ErrDuplicatePrefix", err)
	}
}

func TestIsRegistered(t *testing.T) {
	r := MustNewRegistry(WithType[registryTestData]())
	if !r.IsRegistered("regtest") {
		t.Error("regtest should be registered")
	}
	if r.IsRegistered("unknown") {
		t.Error("unknown should not be registered")
	}
}

func TestRegistryRegister(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		wantErr error
	}{
		{name: "simple", prefix: "user", wantErr: nil},
		{name: "with hyphen", prefix: "my-prefix", wantErr: nil},
		{name: "with underscore", prefix: "a_b_c", wantErr: nil},
		{name: "alphanumeric", prefix: "abc123", wantErr: nil},
		{name: "single char", prefix: "u", wantErr: nil},
		{name: "uppercase", prefix: "User", wantErr: ErrInvalidPrefix},
		{name: "empty", prefix: "", wantErr: ErrInvalidPrefix},
		{name: "contains dot", prefix: "user.name", wantErr: ErrInvalidPrefix},
		{name: "contains space", prefix: "user name", wantErr: ErrInvalidPrefix},
		{name: "special chars", prefix: "user!", wantErr: ErrInvalidPrefix},
		{name: "starts with hyphen", prefix: "-user", wantErr: ErrInvalidPrefix},
		{name: "starts with underscore", prefix: "_user", wantErr: ErrInvalidPrefix},
		{name: "contains tilde", prefix: "user~name", wantErr: ErrInvalidPrefix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := MustNewRegistry()
			err := r.Register(tt.prefix)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Register(%q) = nil, want %v", tt.prefix, tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Register(%q) = %v, want %v", tt.prefix, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Register(%q) unexpected error: %v", tt.prefix, err)
			}
		})
	}
}

func TestRegistryRegisterDuplicate(t *testing.T) {
	r := MustNewRegistry()
	if err := r.Register("user"); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	err := r.Register("user")
	if err == nil {
		t.Fatal("second Register = nil, want ErrDuplicatePrefix")
	}
	if !errors.Is(err, ErrDuplicatePrefix) {
		t.Fatalf("second Register = %v, want ErrDuplicatePrefix", err)
	}
}

func TestRegistryParseAny(t *testing.T) {
	validUserStr, validUserRaw := makeTestEncodedString("user")
	validPostStr, _ := makeTestEncodedString("post")

	r := MustNewRegistry()
	if err := r.Register("user"); err != nil {
		t.Fatal(err)
	}
	if err := r.Register("post"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		input      string
		wantPrefix string
		wantRaw    []byte
		wantErr    error
	}{
		{
			name:       "valid user",
			input:      validUserStr,
			wantPrefix: "user",
			wantRaw:    validUserRaw,
		},
		{
			name:       "valid post",
			input:      validPostStr,
			wantPrefix: "post",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: ErrEmptyString,
		},
		{
			name:    "no separator",
			input:   "userABCDEFGHIJKLMNOPQRSTU",
			wantErr: ErrInvalidFormat,
		},
		{
			name:    "unknown prefix",
			input:   "unknown." + encodeBytes(validUserRaw),
			wantErr: ErrUnknownPrefix,
		},
		{
			name:    "invalid encoding",
			input:   "user.!!invalid!!",
			wantErr: ErrInvalidEncoding,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, rawBytes, err := r.ParseAny(tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("ParseAny(%q) = nil error, want %v", tt.input, tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParseAny(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseAny(%q) unexpected error: %v", tt.input, err)
			}
			if prefix != tt.wantPrefix {
				t.Errorf("ParseAny(%q) prefix = %q, want %q", tt.input, prefix, tt.wantPrefix)
			}
			if tt.wantRaw != nil && !bytes.Equal(rawBytes, tt.wantRaw) {
				t.Errorf("ParseAny(%q) raw bytes mismatch", tt.input)
			}
		})
	}
}

func TestRegistryParseAnyWithCustomSeparator(t *testing.T) {
	data := registryTestData{Val: 42}
	raw, _ := encodeGob(data)
	encoded := encodeBytes(raw)

	r := MustNewRegistry(WithSeparator("~"))
	if err := r.Register("user"); err != nil {
		t.Fatal(err)
	}

	// Should parse with ~ separator
	prefix, rawBytes, err := r.ParseAny("user~" + encoded)
	if err != nil {
		t.Fatalf("ParseAny with ~ separator: %v", err)
	}
	if prefix != "user" {
		t.Errorf("prefix = %q, want %q", prefix, "user")
	}
	if !bytes.Equal(rawBytes, raw) {
		t.Error("raw bytes mismatch")
	}

	// Should fail with . separator
	_, _, err = r.ParseAny("user." + encoded)
	if err == nil {
		t.Fatal("ParseAny with . separator should fail for ~ registry")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("error = %v, want ErrInvalidFormat", err)
	}
}

func TestRegistryConcurrency(t *testing.T) {
	r := MustNewRegistry()
	var wg sync.WaitGroup
	errs := make(chan error, 200)

	// Spawn goroutines registering different prefixes
	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			prefix := fmt.Sprintf("prefix%d", i)
			if err := r.Register(prefix); err != nil {
				errs <- fmt.Errorf("Register(%q): %v", prefix, err)
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}

	// Now parse concurrently
	data := registryTestData{Val: 42}
	raw, _ := encodeGob(data)
	encoded := encodeBytes(raw)

	var wg2 sync.WaitGroup
	errs2 := make(chan error, 200)

	for i := range 100 {
		wg2.Add(1)
		go func(i int) {
			defer wg2.Done()
			prefix := fmt.Sprintf("prefix%d", i)
			s := prefix + "." + encoded
			gotPrefix, gotRaw, err := r.ParseAny(s)
			if err != nil {
				errs2 <- fmt.Errorf("ParseAny(%q): %v", s, err)
				return
			}
			if gotPrefix != prefix {
				errs2 <- fmt.Errorf("ParseAny(%q) prefix = %q, want %q", s, gotPrefix, prefix)
			}
			if !bytes.Equal(gotRaw, raw) {
				errs2 <- fmt.Errorf("ParseAny(%q) raw bytes mismatch", s)
			}
		}(i)
	}

	wg2.Wait()
	close(errs2)

	for err := range errs2 {
		t.Error(err)
	}
}

func TestDefaultRegistryDelegation(t *testing.T) {
	origRegistry := DefaultRegistry
	DefaultRegistry = MustNewRegistry()
	defer func() { DefaultRegistry = origRegistry }()

	if err := DefaultRegistry.Register("testdel"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	data := registryTestData{Val: 42}
	raw, _ := encodeGob(data)
	encoded := encodeBytes(raw)

	prefix, rawBytes, err := ParseAny("testdel." + encoded)
	if err != nil {
		t.Fatalf("ParseAny: %v", err)
	}
	if prefix != "testdel" {
		t.Errorf("prefix = %q, want %q", prefix, "testdel")
	}
	if !bytes.Equal(rawBytes, raw) {
		t.Error("raw bytes mismatch")
	}
}
