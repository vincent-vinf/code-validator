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
			Default: &types.DefaultValidator{
				Name: types.ExactMatchValidator,
			},
		},
		Runtime: types.JavaScriptRuntime,
		Files: []types.File{
			{
				Name: "expected-output",
				Path: "./expected-output",
				Source: types.FileSource{
					Raw: &types.Raw{Content: []byte("Hello World")},
				},
			},
		},
	}
	v, err := validator.New(1, t)
	if err != nil {
		panic(err)
	}
	report, err := v.Exec()
	if err != nil {
		panic(err)
	}

	logger.Info(report.Result)
}
