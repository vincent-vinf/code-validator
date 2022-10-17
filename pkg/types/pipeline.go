package types

type Pipeline struct {
	Name      string
	Steps     []Step
	InputFile []File
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

type FileSource struct {
	URL  *URL  `json:"url"`
	Text *Text `json:"text"`
	OSS  *OSS  `json:"oss"`
}
type URL struct {
	Src string `json:"src"`
}
type Text struct {
	Content string `json:"content"`
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
