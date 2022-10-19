package main

import (
	"github.com/vincent-vinf/code-validator/pkg/types"
	"github.com/vincent-vinf/code-validator/pkg/util/log"
	"github.com/vincent-vinf/code-validator/pkg/validator"
)

var (
	logger = log.GetLogger()
)

func main() {
	t := &types.Task{
		Init: nil,
		Run: types.Code{
			Source: types.FileSource{
				Raw: &types.Raw{Content: []byte("console.log('Hello World');\n")},
			},
		},
		Verify: types.Validator{
			ExactMatch: &types.ExactMatchValidator{
				File1: "./expected-output",
				File2: "",
			},
		},
		Runtime: types.JavaScriptRuntime,
		Files: []*types.File{
			{
				Name: "expected-output",
				Path: "./expected-output",
				Source: types.FileSource{
					Raw: &types.Raw{Content: []byte("Hello World")},
				},
			},
		},
	}
	report, err := validator.New(1).Exec(t)
	if err != nil {
		panic(err)
	}

	logger.Info(report.Result)
}
