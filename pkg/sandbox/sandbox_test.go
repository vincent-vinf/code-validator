package sandbox

import (
	"encoding/json"
	"log"
	"testing"
)

type A struct {
	D []byte
}

func TestIsolate(t *testing.T) {
	d := []byte("123")
	a := &A{
		D: d,
	}
	bytes, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	log.Println(string(bytes))
}
