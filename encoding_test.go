package bpid

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type testEncID struct {
	OrgID   int64
	UserSeq int64
}

func TestEncodeGobRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data testEncID
	}{
		{name: "zero value", data: testEncID{}},
		{name: "positive values", data: testEncID{OrgID: 42, UserSeq: 1001}},
		{name: "negative values", data: testEncID{OrgID: -1, UserSeq: -999}},
		{name: "large values", data: testEncID{OrgID: 1<<62 - 1, UserSeq: 1<<62 - 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := encodeGob(tt.data)
			if err != nil {
				t.Fatalf("encodeGob: %v", err)
			}
			if len(raw) == 0 {
				t.Fatal("encodeGob returned empty bytes")
			}

			decoded, err := decodeGob[testEncID](raw)
			if err != nil {
				t.Fatalf("decodeGob: %v", err)
			}
			if decoded != tt.data {
				t.Errorf("round-trip: got %+v, want %+v", decoded, tt.data)
			}
		})
	}
}

func TestEncodeGobDeterminism(t *testing.T) {
	data := testEncID{OrgID: 42, UserSeq: 1001}

	b1, err := encodeGob(data)
	if err != nil {
		t.Fatalf("first encodeGob: %v", err)
	}
	b2, err := encodeGob(data)
	if err != nil {
		t.Fatalf("second encodeGob: %v", err)
	}
	if !bytes.Equal(b1, b2) {
		t.Errorf("encodeGob not deterministic:\n  first:  %x\n  second: %x", b1, b2)
	}

	// Different data should produce different bytes
	b3, err := encodeGob(testEncID{OrgID: 99, UserSeq: 2002})
	if err != nil {
		t.Fatalf("third encodeGob: %v", err)
	}
	if bytes.Equal(b1, b3) {
		t.Error("different data should produce different bytes")
	}
}

func TestDecodeGobInvalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "garbage", data: []byte{0xFF, 0xFE, 0xFD}},
		{name: "single byte", data: []byte{0x0A}},
		{name: "empty-like", data: []byte{0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeGob[testEncID](tt.data)
			if err == nil {
				t.Fatal("expected error for invalid gob bytes")
			}
			if !errors.Is(err, ErrDecodingFailed) {
				t.Errorf("error = %v, want ErrDecodingFailed", err)
			}
		})
	}
}

func TestEncodeBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{name: "empty", input: []byte{}},
		{name: "single byte", input: []byte{0x42}},
		{name: "16 bytes", input: bytes.Repeat([]byte{0xAB}, 16)},
		{name: "100 bytes", input: bytes.Repeat([]byte{0xCD}, 100)},
	}

	const base64urlChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := encodeBytes(tt.input)
			// Output should only contain base64url characters
			for _, c := range encoded {
				if !strings.ContainsRune(base64urlChars, c) {
					t.Errorf("encodeBytes output contains non-base64url char %q in %q", string(c), encoded)
				}
			}
			// Should not contain padding
			if strings.Contains(encoded, "=") {
				t.Errorf("encodeBytes output contains padding: %q", encoded)
			}
		})
	}
}

func TestDecodeBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "valid", input: "AAAAAAAAAAAAAAAAAAAAAA"},
		{name: "valid short", input: "Qg"},
		{name: "empty", input: ""},
		{name: "invalid chars", input: "!!invalid!!", wantErr: ErrInvalidEncoding},
		{name: "with padding", input: "AAAA==", wantErr: ErrInvalidEncoding},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeBytes(tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("decodeBytes(%q) = nil error, want %v", tt.input, tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("decodeBytes(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil && tt.input != "" {
				t.Fatalf("decodeBytes(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestDecodeBytesInvalid(t *testing.T) {
	_, err := decodeBytes("!!invalid!!")
	if err == nil {
		t.Fatal("expected error for invalid base64url")
	}
	if !errors.Is(err, ErrInvalidEncoding) {
		t.Errorf("error = %v, want ErrInvalidEncoding", err)
	}
}

func TestEncodeDecodeBytesRoundTrip(t *testing.T) {
	inputs := [][]byte{
		{},
		{0x01, 0x02, 0x03},
		bytes.Repeat([]byte{0xFF}, 16),
		bytes.Repeat([]byte{0xAB, 0xCD}, 50),
	}

	for _, input := range inputs {
		encoded := encodeBytes(input)
		decoded, err := decodeBytes(encoded)
		if err != nil {
			t.Fatalf("round-trip failed for %x: encode=%q, error=%v", input, encoded, err)
		}
		if !bytes.Equal(decoded, input) {
			t.Errorf("round-trip mismatch: input=%x, encoded=%q, decoded=%x", input, encoded, decoded)
		}
	}
}
