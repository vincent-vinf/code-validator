package performer

import (
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

//func (p *Performer) Run(task *Task) (*Report, error) {
//	if task == nil {
//		return nil, fmt.Errorf("task cannot be nil")
//	}
//	if task.Runtime != SupportRuntime {
//		return nil, fmt.Errorf("the runtime(%s) is not supported", task.Runtime)
//	}
//
//	executor, err := pipeline.NewExecutor(p.id)
//	if err != nil {
//		return nil, err
//	}
//	verifyTemp, err := task.Verify.ToTemplate()
//	if err != nil {
//		return nil, err
//	}
//
//	steps := []pipeline.Step{
//		{
//			Name:     "npm-init",
//			Template: "npm-init",
//		},
//		{
//			Name:     RunStepName,
//			Template: RunStepName,
//		},
//	}
//	templates := []pipeline.Template{
//		{
//			Name: "npm-init",
//			Cmd:  "/usr/local/bin/npm",
//			Args: []string{
//				"init",
//				"-y",
//			},
//		},
//		{
//			Name: RunStepName,
//			Cmd:  "/usr/local/bin/node",
//			Args: []string{
//				"./index.js",
//			},
//		},
//	}
//	// todo add init step
//
//	pl := &pipeline.Pipeline{
//		Name:  task.Name,
//		Steps: steps,
//		Files: []*pipeline.File{
//			{
//				Name:   "index.js",
//				Path:   "./index.js",
//				Source: task.Run.Source,
//			},
//			{
//				Name: "run-out",
//				Path: fmt.Sprintf("./%s.out", RunStepName),
//				Source: pipeline.FileSource{
//					Host: &pipeline.Host{Path: executor.GetStepLogPath(RunStepName)},
//				},
//			},
//		},
//	}
//
//	if err = executor.Exec(pl); err != nil {
//		return nil, fmt.Errorf("exec fail, err: %w", err)
//	}
//
//	file, err := executor.ReadFile("./report/result")
//	if err != nil {
//		return nil, err
//	}
//
//	return &Report{
//		Result: string(file),
//	}, nil
//}
