package util

import (
	"encoding/json"
	"fmt"
	"path"
)

const (
	DefaultAPIPrefix = "/api"
)

func LogStruct(obj any) {
	val, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return
	}
	fmt.Println(string(val))
}

func WithGlobalAPIPrefix(p string) string {
	return path.Join(DefaultAPIPrefix, p)
}
