package validator

import (
	"fmt"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/types"
)

const (
	SupportRuntime = types.JavaScriptRuntime
)

type Validator struct {
	id int
}

func New(id int) *Validator {
	return &Validator{
		id: id,
	}
}

func (e *Validator) Exec(task *types.Task) (*types.Report, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}
	if task.Runtime != SupportRuntime {
		return nil, fmt.Errorf("the runtime(%s) is not supported", task.Runtime)
	}

	controller, err := pipeline.NewController(e.id)
	if err != nil {
		return nil, err
	}
	verify, err := task.Verify.ToStep()
	if err != nil {
		return nil, err
	}
	verify.RefFiles = []string{
		"run-out",
		"expected-output",
	}
	steps := []*types.Step{
		{
			Name: "npm-init",
			Cmd:  "/usr/local/bin/npm",
			Args: []string{
				"init",
				"-y",
			},
		},
	}
	if task.Init != nil {
		steps = append(steps, task.Init)
	}
	steps = append(steps,
		&types.Step{
			Name: types.RunStepName,
			Cmd:  "/usr/local/bin/node",
			Args: []string{
				"./index.js",
			},
		},
		verify,
	)
	p := &types.Pipeline{
		Name:  task.Name,
		Steps: steps,
		Files: append(
			[]*types.File{
				{
					Name:   "index.js",
					Path:   "./index.js",
					Source: task.Run.Source,
					Type:   types.GlobalFileType,
				},
				{
					Name: "run-out",
					Path: fmt.Sprintf("./%s.out", types.RunStepName),
					Source: types.FileSource{
						Host: &types.Host{Path: controller.GetStepLogPath(types.RunStepName)},
					},
				},
			},
			task.Files...,
		),
	}

	if err = controller.Exec(p); err != nil {
		return nil, fmt.Errorf("exec fail, err: %w", err)
	}

	file, err := controller.ReadFile("./report/result")
	if err != nil {
		return nil, err
	}

	return &types.Report{
		Result: string(file),
	}, nil
}
