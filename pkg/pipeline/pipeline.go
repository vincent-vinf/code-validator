package pipeline

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const (
	GlobalFileType = "global"
	// LocalFileType is only visible within the step
	// when filetype is not specified, LocalFileType is the default value
	LocalFileType = "local"

	DefaultTempDir = "/var/local/lib/code-pipeline"

	StepLogFile = "out"
)

type Pipeline struct {
	Name  string
	Steps []*Step
	Files []*File
}

type Step struct {
	Name string

	Cmd  string
	Args []string

	StdinFile string
	RefFiles  []string
}

type File struct {
	Name string
	Type string
	// Path must be a writable path in the sandbox
	Path   string
	Source FileSource
}

func (f *File) Read() ([]byte, error) {
	switch {
	case f.Source.Raw != nil:
		return f.Source.Raw.Content, nil
	case f.Source.OSS != nil:
		return []byte{}, nil
	case f.Source.URL != nil:
		return []byte{}, nil
	case f.Source.Host != nil:
		p := path.Clean(f.Source.Host.Path)
		if !strings.HasPrefix(p, DefaultTempDir) {
			return nil, fmt.Errorf("invalid path %s", f.Source.Host.Path)
		}

		return os.ReadFile(p)
	default:
		return nil, fmt.Errorf("file(%s) source not specified", f.Name)
	}
}

type FileSource struct {
	URL  *URL  `json:"url"`
	Raw  *Raw  `json:"raw"`
	OSS  *OSS  `json:"oss"`
	Host *Host `json:"host"`
}
type URL struct {
	Src string `json:"src"`
}
type Raw struct {
	Content []byte `json:"content"`
}
type OSS struct {
	Path string `json:"path"`
}
type Host struct {
	Path string `json:"path"`
}

//{
//"url":{
//"src":""
//},
//"text":{
//"content":""
//},
//"oss":{
//"path":""
//}
//}
