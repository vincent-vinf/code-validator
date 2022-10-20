package performer

import (
	"fmt"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/types"
)

const (
	SupportRuntime = types.JavaScriptRuntime
)

type Performer struct {
	id int
}

func New(id int) *Performer {
	return &Performer{
		id: id,
	}
}

func (p *Performer) Run(task *Task) (*Report, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}
	if task.Runtime != SupportRuntime {
		return nil, fmt.Errorf("the runtime(%s) is not supported", task.Runtime)
	}

	controller, err := pipeline.NewController(p.id)
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
	steps := []*pipeline.Step{
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
		&pipeline.Step{
			Name: RunStepName,
			Cmd:  "/usr/local/bin/node",
			Args: []string{
				"./index.js",
			},
		},
		verify,
	)
	pl := &pipeline.Pipeline{
		Name:  task.Name,
		Steps: steps,
		Files: append(
			[]*pipeline.File{
				{
					Name:   "index.js",
					Path:   "./index.js",
					Source: task.Run.Source,
					Type:   pipeline.GlobalFileType,
				},
				{
					Name: "run-out",
					Path: fmt.Sprintf("./%s.out", RunStepName),
					Source: pipeline.FileSource{
						Host: &pipeline.Host{Path: controller.GetStepLogPath(RunStepName)},
					},
				},
			},
		),
	}

	if err = controller.Exec(pl); err != nil {
		return nil, fmt.Errorf("exec fail, err: %w", err)
	}

	file, err := controller.ReadFile("./report/result")
	if err != nil {
		return nil, err
	}

	return &Report{
		Result: string(file),
	}, nil
}
