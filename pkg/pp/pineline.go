package pp

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
	Name     string
	Template string

	RefFiles []RefFile
}

type File struct {
	Name string
}

type RefFile struct {
	Name       string
	Path       string
	AutoRemove bool
}
