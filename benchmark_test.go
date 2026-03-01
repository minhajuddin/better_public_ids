package bpid

import (
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

func BenchmarkMarshalJSON(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	b.ResetTimer()
	for b.Loop() {
		_, _ = id.MarshalJSON()
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	data, _ := id.MarshalJSON()
	b.ResetTimer()
	for b.Loop() {
		var parsed ID[benchUserDef]
		_ = parsed.UnmarshalJSON(data)
	}
}

func BenchmarkMarshalText(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	b.ResetTimer()
	for b.Loop() {
		_, _ = id.MarshalText()
	}
}

func BenchmarkUnmarshalText(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	data, _ := id.MarshalText()
	b.ResetTimer()
	for b.Loop() {
		var parsed ID[benchUserDef]
		_ = parsed.UnmarshalText(data)
	}
}

func BenchmarkSQLValue(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	b.ResetTimer()
	for b.Loop() {
		_, _ = id.Value()
	}
}

func BenchmarkSQLScan(b *testing.B) {
	id := MustNew(benchUserDef{OrgID: 42, UserSeq: 1001})
	s := id.String()
	b.ResetTimer()
	for b.Loop() {
		var parsed ID[benchUserDef]
		_ = parsed.Scan(s)
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
