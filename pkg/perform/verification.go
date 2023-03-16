package perform

import (
	"context"
	"fmt"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
)

type Verification struct {
	Name    string              `json:"name"`
	Runtime string              `json:"runtime"`
	Code    *CodeVerification   `json:"code,omitempty"`
	Custom  *CustomVerification `json:"custom,omitempty"`
}

type CodeVerification struct {
	Init   *Action    `json:"init"`
	Verify string     `json:"verify"`
	Files  []File     `json:"files"`
	Cases  []TestCase `json:"cases"`
}

type CustomVerification struct {
	Action
}

type Action struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Files   []File `json:"files"`
}

func (a *Action) ToStep() *pipeline.Step {
	fileRefs := make([]pipeline.FileRef, len(a.Files))
	for i := range a.Files {
		fileRefs[i].Path = a.Files[i].Path
		fileRefs[i].DataRef.ExternalRef = &pipeline.ExternalRef{
			FileName: fmt.Sprintf("%s-%s", a.Name, a.Files[i].Path),
		}
	}
	return &pipeline.Step{
		Name: a.Name,
		InlineTemplate: &pipeline.Template{
			Name: a.Name,
			Cmd:  "/bin/sh",
			Args: []string{"-c", a.Command},
		},
		InputRef:       nil,
		FileRefs:       fileRefs,
		ContinueOnFail: false,
		LogMate:        false,
		Limit:          nil,
	}
}

func (a *Action) GetFiles() ([]pipeline.File, error) {
	res := make([]pipeline.File, len(a.Files))
	for i := range a.Files {
		data, err := a.Files[i].Read()
		if err != nil {
			return nil, err
		}
		res[i].Content = data
		res[i].Name = fmt.Sprintf("%s-%s", a.Name, a.Files[i].Path)
	}

	return res, nil
}

type File struct {
	Path string `json:"path,omitempty"`
	//IsZipped bool   `json:"isZipped"`
	OssPath string `json:"ossPath,omitempty"`
}

func (f *File) Read() ([]byte, error) {
	return ossClient.Get(context.Background(), f.OssPath)
}

type Report struct {
	Pass    bool         `json:"pass"`
	Message string       `json:"message,omitempty"`
	Cases   []CaseResult `json:"cases,omitempty"`
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
