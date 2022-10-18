package types

import "fmt"

const (
	ExactMatchValidator = "exact-match"

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
	Custom  *Step
	Default *DefaultValidator
}

func (v Validator) ToStep() (*Step, error) {
	if v.Custom != nil {
		return nil, fmt.Errorf("not implemented")
	} else if v.Default != nil {
		switch v.Default.Name {
		case ExactMatchValidator:
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
`, "./std-run/stdout", "./expected-output"),
				},
			}, nil
		default:
			return nil, fmt.Errorf("unknown validator(%s)", v.Default.Name)
		}
	} else {
		return nil, fmt.Errorf("no validator specified")
	}
}

type DefaultValidator struct {
	Name string
}

type Report struct {
	Result string
	Cases  []Case
}

type Case struct {
	Result string
	Time   string
}
