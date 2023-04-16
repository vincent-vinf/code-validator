package perform

import (
	"context"
	"fmt"
	"path"

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
			FileName: GetFileName(a.Name, a.Files[i].Path),
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

func (a *Action) GetFiles(srcDir string) ([]pipeline.File, error) {
	return ToPipelineFile(srcDir, a.Name, a.Files)
}

func ToPipelineFile(srcDir, stepName string, files []File) (res []pipeline.File, err error) {
	var data []byte
	for i := range files {
		data, err = ReadOSSFile(path.Join(srcDir, files[i].OssPath))
		if err != nil {
			return
		}
		res = append(res, pipeline.File{
			Name:    GetFileName(stepName, files[i].Path),
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

func ReadOSSFile(path string) ([]byte, error) {
	data, err := ossClient.Get(context.Background(), path)
	if err != nil {
		return nil, fmt.Errorf("path %s, err: %w", path, err)
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

func GetFileName(stepName, path string) string {
	return fmt.Sprintf("%s-%s", stepName, path)
}
