//go:build javascript

package perform_back

import (
	"fmt"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/types"
)

const (
	npmInstallStepName = "npm-init"
)

var (
	SupportRuntime = types.Runtime{
		Lang:    types.JavaScriptRuntime,
		Version: "latest",
	}
)

type Lang struct {
	runtime types.Runtime
}

func NewLang(runtime types.Runtime) (*Lang, error) {
	if runtime != SupportRuntime {
		return nil, fmt.Errorf("the runtime(%s/%s) is not supported", runtime.Lang, runtime.Version)
	}

	return &Lang{runtime: runtime}, nil
}

func (l *Lang) GetTemplates() []pipeline.Template {
	return []pipeline.Template{
		{
			Name: npmInstallStepName,
			Cmd:  "/usr/local/bin/npm",
			Args: []string{
				"init",
				"-y",
			},
		},
		{
			Name: RunStepName,
			Cmd:  "/usr/local/bin/node",
			Args: []string{
				"./index.js",
			},
		},
	}
}

func (l *Lang) GetRunSteps() []pipeline.Step {
	return []pipeline.Step{
		{
			Name:     npmInstallStepName,
			Template: npmInstallStepName,
			FileRefs: []pipeline.FileRef{
				{
					DataRef: &pipeline.DataRef{
						ExternalRef: &pipeline.ExternalRef{FileName: "code"},
					},
					Path: "./index.js",
				},
			},
		},
		{
			Name:     RunStepName,
			Template: RunStepName,
			LogMate:  true,
			InputRef: &pipeline.DataRef{
				ExternalRef: &pipeline.ExternalRef{FileName: "input"},
			},
		},
	}
}
