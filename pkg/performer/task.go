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
	Run    Code
	Verify Validator

	Cases []TestCase

	Name    string
	Runtime string
}
type Init struct {
}
type Code struct {
	Source pipeline.FileSource
}

type Validator struct {
	Custom     *pipeline.Template
	ExactMatch *ExactMatchValidator
}

type TestCase struct {
	Name   string
	Input  *pipeline.FileSource
	Output *pipeline.FileSource
}

func (v Validator) ToTemplate() (*pipeline.Template, error) {
	switch {
	case v.Custom != nil:
		return nil, fmt.Errorf("not implemented")
	case v.ExactMatch != nil:
		return &pipeline.Template{
			Name: VerifyStepName,
			Cmd:  "/usr/local/bin/code-validator",
			Args: []string{
				"match",
				v.ExactMatch.File1,
				v.ExactMatch.File2,
			},
		}, nil
	default:
		return nil, fmt.Errorf("no validator specified")
	}
}

type ExactMatchValidator struct {
	File1 string
	File2 string
}

type Report struct {
	Result string
	Cases  []CaseResult
}

type CaseResult struct {
	Result string
	Time   string
}
