package pp

type Pipeline struct {
	Name      string
	Steps     []*Step
	Templates []*Template
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
type RefFile struct {
}
