package validator

import (
	"fmt"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/sandbox"
	"github.com/vincent-vinf/code-validator/pkg/types"
)

const (
	SupportRuntime = types.JavaScriptRuntime
)

type Validator struct {
	task       *types.Task
	controller *pipeline.Controller
}

func (e *Validator) Exec() (*types.Report, error) {
	if e.task == nil || e.controller == nil {
		return nil, fmt.Errorf("task and controller cannot be nil")
	}
	verify, err := e.task.Verify.ToStep()
	if err != nil {
		return nil, err
	}
	steps := []types.Step{
		{
			Name: "npm-init",
			Cmd:  "/usr/local/bin/npm",
			Args: []string{
				"init",
				"-y",
			},
		},
	}
	if e.task.Init != nil {
		steps = append(steps, *e.task.Init)
	}
	steps = append(steps,
		types.Step{
			Name: types.RunStepName,
			Cmd:  "/usr/local/bin/node",
			Args: []string{
				"./index.js",
			},
		},
		*verify,
	)
	p := types.Pipeline{
		Name:  e.task.Name,
		Steps: steps,
		Files: append(
			[]types.File{
				{
					Name:   "index.js",
					Path:   "./index.js",
					Source: e.task.Run.Source,
				},
			},
			e.task.Files...,
		),
	}

	if err = e.controller.Exec(p); err != nil {
		return nil, fmt.Errorf("exec fail, err: %w", err)
	}

	file, err := e.controller.ReadFile("./report/result")
	if err != nil {
		return nil, err
	}

	return &types.Report{
		Result: string(file),
	}, nil
}

func New(id int, task *types.Task) (*Validator, error) {
	if task.Runtime != SupportRuntime {
		return nil, fmt.Errorf("the runtime(%s) is not supported", task.Runtime)
	}
	box, err := sandbox.New(id)
	if err != nil {
		return nil, err
	}
	controller, err := pipeline.NewController(box)
	if err != nil {
		return nil, err
	}

	return &Validator{
		task:       task,
		controller: controller,
	}, nil
}
