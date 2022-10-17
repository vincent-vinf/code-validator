package types

import "fmt"

type Pipeline struct {
	Name  string
	Steps []Step
	Files []File
}

type Step struct {
	Name      string
	Cmd       string
	Args      []string
	Type      string
	StdinFile string
}

type File struct {
	Name string
	// path in sandbox
	Path   string
	Source FileSource
}

func (f *File) Read() ([]byte, error) {
	if f.Source.Raw != nil {
		return f.Source.Raw.Content, nil
	} else if f.Source.OSS != nil {
		return []byte{}, nil
	} else if f.Source.URL != nil {
		return []byte{}, nil
	} else {
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
