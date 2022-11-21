package types

const (
	JavaScriptRuntime = "Javascript"
	PythonRuntime     = "Python"
	CPPRuntime        = "CPP"
	CRuntime          = "C"
)

type Runtime struct {
	Lang    string
	Version string
}
