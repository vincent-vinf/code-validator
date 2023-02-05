//go:build python

package perform_back

import (
	"fmt"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/types"
)

var (
	SupportRuntime = types.Runtime{
		Lang:    types.PythonRuntime,
		Version: "latest",
	}
)

type Lang struct {
	runtime types.Runtime
}

func NewLang(runtime types.Runtime) (*Lang, error) {
	if runtime != SupportRuntime {
		return nil, fmt.Errorf("the runtime(%s) is not supported", runtime)
	}

	return &Lang{runtime: runtime}, nil
}

func (l *Lang) GetTemplates() []pipeline.Template {
	return []pipeline.Template{
		{
			Name: RunStepName,
			Cmd:  "/usr/local/bin/python",
			Args: []string{
				"./main.py",
			},
		},
	}
}

func (l *Lang) GetRunSteps() []pipeline.Step {
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
					DataRef: &pipeline.DataRef{
						ExternalRef: &pipeline.ExternalRef{FileName: "code"},
					},
					Path: "./main.py",
				},
			},
		},
	}
}
