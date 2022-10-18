package types

import "fmt"

const (
	GlobalFileType = "global"
	// LocalFileType is only visible within the step
	// when filetype is not specified, LocalFileType is the default value
	LocalFileType = "local"
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

	StdinFile  string
	MountFiles []string
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
	default:
		return nil, fmt.Errorf("unknown file(%s) source", f.Name)
	}
}

type FileSource struct {
	URL *URL `json:"url"`
	Raw *Raw `json:"raw"`
	OSS *OSS `json:"oss"`
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
