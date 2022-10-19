package types

import "fmt"

const (
	InitStepName   = "init"
	RunStepName    = "run"
	VerifyStepName = "verify"
)

type Task struct {
	// step
	Init   *Step
	Run    Code
	Verify Validator

	Files []*File

	Name    string
	Runtime string
}

type Code struct {
	Source FileSource
}

type Validator struct {
	Custom     *Step
	ExactMatch *ExactMatchValidator
}

func (v Validator) ToStep() (*Step, error) {
	switch {
	case v.Custom != nil:
		return nil, fmt.Errorf("not implemented")
	case v.ExactMatch != nil:
		return &Step{
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
	Cases  []Case
}

type Case struct {
	Result string
	Time   string
}
