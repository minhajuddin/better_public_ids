package bpid

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	DefaultRegistry = MustNewRegistry(
		WithType[testUserID](),
		WithType[testPostID](),
		WithType[testAutoRegID](),
		WithType[fuzzUserID](),
		WithType[benchUserID](),
		WithType[testEncID](),
		WithType[testRegID](),
	)
	os.Exit(m.Run())
}
