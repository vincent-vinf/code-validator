package perform

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
)

type Verification struct {
	Name    string              `json:"name"`
	Runtime string              `json:"runtime"`
	Code    *CodeVerification   `json:"code,omitempty"`
	Custom  *CustomVerification `json:"custom,omitempty"`
}

func (vf Verification) String() string {
	data, err := yaml.Marshal(vf)
	if err != nil {
		return fmt.Sprintf("yaml marshal err: %s", err)
	}
	return string(data)
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
			FileName: getFileName(a.Name, a.Files[i].Path),
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
	return ToPipelineFile(a.Name, a.Files)
}

func ToPipelineFile(stepName string, files []File) (res []pipeline.File, err error) {
	var data []byte
	for i := range files {
		data, err = files[i].Read()
		if err != nil {
			return
		}
		res = append(res, pipeline.File{
			Name:    getFileName(stepName, files[i].Path),
			Content: data,
		})
	}

	return
}

type File struct {
	Path string `json:"path,omitempty"`
	//IsZipped bool   `json:"isZipped"`
	OssPath string `json:"ossPath,omitempty"`
}

func (f *File) Read() ([]byte, error) {
	data, err := ossClient.Get(context.Background(), f.OssPath)
	if err != nil {
		return nil, fmt.Errorf("path %s, err: %w", f.OssPath, err)
	}

	return data, nil
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

func getFileName(stepName, path string) string {
	return fmt.Sprintf("%s-%s", stepName, path)
}
