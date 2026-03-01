package bpid

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	RegisterType[testUserID]()
	RegisterType[testPostID]()
	RegisterType[testAutoRegID]()
	RegisterType[fuzzUserID]()
	RegisterType[benchUserID]()
	RegisterType[testEncID]()
	RegisterType[testRegID]()
	os.Exit(m.Run())
}
