package bpid

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	DefaultRegistry = MustNewRegistry(
		WithType[userIDDef](),
		WithType[postIDDef](),
		WithType[autoRegTestDef](),
		WithType[fuzzUserDef](),
		WithType[benchUserDef](),
		WithType[encTestData](),
		WithType[registryTestData](),
	)
	os.Exit(m.Run())
}
