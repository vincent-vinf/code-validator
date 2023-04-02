package util

import (
	"encoding/json"
	"fmt"
)

func LogStruct(obj any) {
	val, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return
	}
	fmt.Println(string(val))
}
