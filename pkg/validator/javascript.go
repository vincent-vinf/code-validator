package validator

import (
	"fmt"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/sandbox"
	"github.com/vincent-vinf/code-validator/pkg/types"
)

const (
	SupportRuntime = types.RuntimeJavaScript
)

type Validator struct {
	task       *types.Task
	controller *pipeline.Controller
}

func (e *Validator) Exec() (*types.Report, error) {
	return nil, nil
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
