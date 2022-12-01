package pipeline

import (
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
	Name           string
	Template       string
	InlineTemplate *Template
	InputRef       *DataRef
	FileRefs       []FileRef

	ContinueOnFail bool
	LogMate        bool

	Limit *Limit
}

type Limit struct {
	EnableNetWork bool
	Memory        int
	Time          float64
}

type DataRef struct {
	ExternalRef *ExternalRef `json:"externalRef,omitempty"`
	StepOutRef  *StepOutRef  `json:"stepOutRef,omitempty"`
}
type FileRef struct {
	*DataRef

	Path       string `json:"path"`
	AutoRemove bool   `json:"autoRemove"`
}
type ExternalRef struct {
	FileName string `json:"fileName"`
}
type StepOutRef struct {
	StepName string `json:"stepName"`
}

type File struct {
	Name    string
	Content []byte `json:"content"`
}

type Result struct {
	Metas map[string]*sandbox.Meta
	Errs  map[string]error
}
