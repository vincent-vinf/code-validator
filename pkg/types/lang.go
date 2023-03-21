package types

const (
	JavaScriptRuntime = "Javascript"
	PythonRuntime     = "Python"
	GolangRuntime     = "Golang"
	CPPRuntime        = "CPP"
	CRuntime          = "C"
)

type Runtime struct {
	Lang    string `json:"lang"`
	Version string `json:"version"`
}
