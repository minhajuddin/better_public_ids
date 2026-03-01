package bpid

import (
	"bytes"
	"encoding/gob"
	"testing"
)

type benchUserDef struct {
	OrgID   int64
	UserSeq int64
}

func (benchUserDef) Prefix() string { return "benchuser" }

func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		_, _ = New(benchUserDef{OrgID: 42, UserSeq: 1001})
	}
}

func BenchmarkString(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	b.ResetTimer()
	for b.Loop() {
		_ = id.String()
	}
}

func BenchmarkParse(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	s := id.String()
	b.ResetTimer()
	for b.Loop() {
		_, _ = Parse[benchUserDef](s)
	}
}

func BenchmarkData(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	b.ResetTimer()
	for b.Loop() {
		_, _ = id.Data()
	}
}

func BenchmarkEqual(b *testing.B) {
	id1 := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	id2 := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	b.ResetTimer()
	for b.Loop() {
		_ = id1.Equal(id2)
	}
}

func BenchmarkGobEncode(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	b.ResetTimer()
	for b.Loop() {
		var buf bytes.Buffer
		_ = gob.NewEncoder(&buf).Encode(&id)
	}
}

func BenchmarkGobDecode(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(&id)
	data := buf.Bytes()
	b.ResetTimer()
	for b.Loop() {
		var parsed ID[benchUserDef]
		_ = gob.NewDecoder(bytes.NewReader(data)).Decode(&parsed)
	}
}

func BenchmarkRegistryParseAny(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	s := id.String()
	b.ResetTimer()
	for b.Loop() {
		_, _, _ = ParseAny(s)
	}
}
