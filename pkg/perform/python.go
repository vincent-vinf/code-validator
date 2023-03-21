package perform

import (
	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/types"
)

const (
	Runtime = types.PythonRuntime
)

func GetCodeTemplates() []pipeline.Template {
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
					Path: "./main.py",
				},
			},
		},
	}
}
