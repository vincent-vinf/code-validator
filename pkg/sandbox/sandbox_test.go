package sandbox

import (
	"testing"
)

func init() {
}

func TestIsolate(t *testing.T) {
	m := &Meta{}
	err := m.ReadFile("/Users/vincent/Documents/repo/code-validator/meta")
	if err != nil {
		t.Fatal(err)
	}
}
