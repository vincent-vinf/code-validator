package sandbox

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

func init() {
}

func TestIsolate(t *testing.T) {
	//m := &Meta{}
	//err := m.ReadFile("/Users/vincent/Documents/repo/code-validator/meta")
	//if err != nil {
	//	t.Fatal(err)
	//}
	m := map[string]interface{}{
		"1": map[string]string{
			"1": "2",
		},
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}
