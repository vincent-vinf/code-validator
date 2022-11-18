package performer

import (
	"fmt"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
)

const (
	InitStepName   = "init"
	RunStepName    = "run"
	VerifyStepName = "verify"
)

type Task struct {
	Init   *Init
	Run    Run
	Verify Validator

	Cases []TestCase

	Name    string
	Runtime string
}
type Init struct {
}
type Run struct {
	SourceCode []byte
}

type Validator struct {
	Custom  *pipeline.Template
	Default *DefaultValidator
}

type TestCase struct {
	Name   string
	Input  []byte
	Output []byte
}

func (v Validator) ToTemplate() (*pipeline.Template, error) {
	switch {
	case v.Custom != nil:
		return nil, fmt.Errorf("not implemented")
	case v.Default != nil:
		return &pipeline.Template{
			Name: VerifyStepName,
			Cmd:  "/usr/local/bin/code-performer",
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
