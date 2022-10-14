package sandbox

import (
	"log"
	"path"
	"testing"
)

func init() {
}

func TestIsolate(t *testing.T) {
	log.Println(path.Join("box", "./a"))
}
