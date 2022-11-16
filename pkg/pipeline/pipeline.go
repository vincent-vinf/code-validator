package pipeline

import (
	"fmt"

	"github.com/vincent-vinf/code-validator/pkg/sandbox"
)

type Pipeline struct {
	Name      string
	Steps     []Step
	Templates []Template
	Files     []File
}

type Template struct {
	Name string

	Cmd  string
	Args []string
}

type Step struct {
	Name      string
	Template  string
	InputFile *InputFile
	RefFiles  []RefFile

	ContinueOnFail bool
	NeedMate       bool

	Limit *Limit
}
type Limit struct {
}

type InputFile struct {
	Source  *FileSource
	StepOut *StepOut
}
type StepOut struct {
	StepName string
}

type File struct {
	Name   string
	Source FileSource
}

type RefFile struct {
	FileName   string `json:"fileName"`
	Path       string `json:"path"`
	AutoRemove bool   `json:"autoRemove"`
}

type FileSource struct {
	Raw *Raw `json:"raw,omitempty"`
	OSS *OSS `json:"oss,omitempty"`
}

func (f *FileSource) Read() ([]byte, error) {
	switch {
	case f.Raw != nil:
		return f.Raw.Content, nil
	case f.OSS != nil:
		return []byte{}, nil
	default:
		return nil, fmt.Errorf("file source not specified")
	}
}

type Raw struct {
	Content []byte `json:"content"`
}
type OSS struct {
	Path string `json:"path"`
}
type WorkdirFile struct {
	Path string `json:"path"`
}
type Result struct {
	Metas map[string]*sandbox.Meta
	Errs  map[string]error
}
