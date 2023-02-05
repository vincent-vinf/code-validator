package perform

import (
	"context"
	"github.com/vincent-vinf/code-validator/pkg/pipeline"
)

type Verification struct {
	Name    string              `json:"name"`
	Runtime string              `json:"runtime"`
	Code    *CodeVerification   `json:"code"`
	Custom  *CustomVerification `json:"custom"`
}

type CodeVerification struct {
	Init   *Action    `json:"init"`
	Verify string     `json:"verify"`
	Cases  []TestCase `json:"cases"`
}

type CustomVerification struct {
	Actions []Action `json:"actions"`
}

type Action struct {
	Name    string
	Command string `json:"command"`
	Files   []File `json:"files"`
}

func (a *Action) ToStep() *pipeline.Step {
	return &pipeline.Step{
		Name: a.Name,
		InlineTemplate: &pipeline.Template{
			Name: a.Name,
			Cmd:  "sh",
			Args: []string{"-c", a.Command},
		},
		InputRef:       nil,
		FileRefs:       nil,
		ContinueOnFail: false,
		LogMate:        false,
		Limit:          nil,
	}
}

type File struct {
	Path string `json:"path"`
	//IsZipped bool   `json:"isZipped"`
	OssPath string `json:"ossPath"`
}

func (f *File) Read() ([]byte, error) {
	return ossClient.Get(context.Background(), f.OssPath)
}

type Report struct {
	Pass    bool
	Message string
	Cases   []CaseResult
}

type TestCase struct {
	Name string
	In   File
	Out  File
}

type CaseResult struct {
	Name    string
	Pass    bool
	Message string

	ExitCode int
	Time     float64
	Memory   int
}
