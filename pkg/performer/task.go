package performer

import (
	"fmt"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/types"
)

const (
	InitStepName   = "init"
	RunStepName    = "run"
	VerifyStepName = "verify"
)

type Task struct {
	Name    string
	Runtime types.Runtime

	Init   *Action
	Code   Code
	Verify Validator

	Cases []TestCase
}

type Code struct {
	Data []byte
}

type Action struct {
	Name  string
	Cmd   string
	Args  []string
	Files []File
}

func (a *Action) GetStep() *pipeline.Step {
	return &pipeline.Step{
		Name: a.Name,
		InlineTemplate: &pipeline.Template{
			Name: a.Name,
			Cmd:  a.Cmd,
			Args: a.Args,
		},
		InputRef:       nil,
		FileRefs:       nil,
		ContinueOnFail: false,
		LogMate:        false,
		Limit:          nil,
	}
}
func (a *Action) GetFiles() {

}

type TestCase struct {
	Name   string
	Input  []byte
	Output []byte
}
type Validator struct {
	Custom  *Action
	Default *DefaultValidator
}

func (v Validator) ToTemplate() (*pipeline.Template, error) {
	switch {
	case v.Custom != nil:
		return nil, fmt.Errorf("not implemented")
	case v.Default != nil:
		return &pipeline.Template{
			Name: VerifyStepName,
			Cmd:  "/usr/local/bin/code-validator",
			Args: []string{
				"match",
				"./output",
				"./answer",
			},
		}, nil
	default:
		return nil, fmt.Errorf("no performer specified")
	}
}

type DefaultValidator struct{}

type Report struct {
	Result string
	Cases  []CaseResult
}

type CaseResult struct {
	Name   string
	Result bool

	ExitCode int
	Time     float64
	Memory   int
}

type File struct {
	Path       string `json:"path"`
	Content    []byte `json:"content"`
	AutoRemove bool   `json:"autoRemove"`
}
