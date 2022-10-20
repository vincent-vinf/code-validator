package main

import (
	"github.com/vincent-vinf/code-validator/pkg/performer"
	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/types"
	"github.com/vincent-vinf/code-validator/pkg/util/log"
)

var (
	logger = log.GetLogger()
)

func main() {
	t := &performer.Task{
		Init: nil,
		Run: performer.Code{
			Source: pipeline.FileSource{
				Raw: &pipeline.Raw{Content: []byte("console.log('Hello World');\n")},
			},
		},
		Verify: performer.Validator{
			ExactMatch: &performer.ExactMatchValidator{
				File1: "./expected-output",
				File2: "",
			},
		},
		Runtime: types.JavaScriptRuntime,
		Cases: []performer.TestCase{
			{
				Name:  "hello",
				Input: nil,
				Output: &pipeline.File{
					Name: "expected-output",
					Path: "./expected-output",
					Source: pipeline.FileSource{
						Raw: &pipeline.Raw{Content: []byte("Hello World")},
					},
				},
			},
		},
	}
	report, err := performer.New(1).Run(t)
	if err != nil {
		panic(err)
	}

	logger.Info(report.Result)
}
