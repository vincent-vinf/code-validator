//go:build golang

package perform

import (
	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/types"
)

const (
	Runtime = types.GolangRuntime
)

func GetCodeTemplates() []pipeline.Template {
	return []pipeline.Template{
		{
			Name: RunStepName,
			Cmd:  "sh",
			Args: []string{
				"-c",
				"/usr/local/go/bin/go mod init code.vinf.top/user/code && go run ./main.go",
			},
		},
	}
}

func GetCodeSteps() []pipeline.Step {
	return []pipeline.Step{
		{
			Name:     RunStepName,
			Template: RunStepName,
			LogMate:  true,
			InputRef: &pipeline.DataRef{
				ExternalRef: &pipeline.ExternalRef{FileName: "input"},
			},
			FileRefs: []pipeline.FileRef{
				{
					DataRef: pipeline.DataRef{
						ExternalRef: &pipeline.ExternalRef{FileName: "code"},
					},
					Path: "./main.go",
				},
			},
		},
	}
}
