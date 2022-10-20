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
	Init   *pipeline.Step
	Run    Code
	Verify Validator

	Cases []TestCase

	Name    string
	Runtime string
}

type Code struct {
	Source pipeline.FileSource
}

type Validator struct {
	Custom     *pipeline.Step
	ExactMatch *ExactMatchValidator
}

type TestCase struct {
	Name   string
	Input  *pipeline.File
	Output *pipeline.File
}

func (v Validator) ToStep() (*pipeline.Step, error) {
	switch {
	case v.Custom != nil:
		return nil, fmt.Errorf("not implemented")
	case v.ExactMatch != nil:
		return &pipeline.Step{
			Name: VerifyStepName,
			Cmd:  "/bin/sh",
			Args: []string{
				"-c",
				fmt.Sprintf(
					`
code-validator match %s %s; ec=$?
mkdir -p ./report;
case $ec in
    0) echo pass > ./report/result;;
    1) exit 1;;
    *) echo fail > ./report/result;;
esac
`, v.ExactMatch.File1, v.ExactMatch.File2),
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
