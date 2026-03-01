package bpid

import "testing"

type benchUserID struct {
	OrgID   int64
	UserSeq int64
}

func (benchUserID) Prefix() string { return "benchuser" }

var benchRegistry = MustNewRegistry(WithType[benchUserID]())

func BenchmarkSerialize(b *testing.B) {
	for b.Loop() {
		_, _ = Serialize(benchRegistry, benchUserID{OrgID: 42, UserSeq: 1001})
	}
}

func BenchmarkMustSerialize(b *testing.B) {
	for b.Loop() {
		_ = MustSerialize(benchRegistry, benchUserID{OrgID: 42, UserSeq: 1001})
	}
}

func BenchmarkDeserialize(b *testing.B) {
	s := MustSerialize(benchRegistry, benchUserID{OrgID: 42, UserSeq: 1001})
	b.ResetTimer()
	for b.Loop() {
		_, _ = Deserialize[benchUserID](benchRegistry, s)
	}
}

func BenchmarkRegistryPrefix(b *testing.B) {
	s := MustSerialize(benchRegistry, benchUserID{OrgID: 42, UserSeq: 1001})
	b.ResetTimer()
	for b.Loop() {
		_, _ = benchRegistry.Prefix(s)
	}
}
