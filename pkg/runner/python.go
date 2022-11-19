//go:build python

package runner

import (
	"bytes"

	"github.com/vincent-vinf/code-validator/pkg/sandbox"
	"github.com/vincent-vinf/code-validator/pkg/types"
)

const (
	SupportRuntime = types.PythonRuntime
)

func Run(id int, input, code []byte) ([]byte, error) {
	box, err := sandbox.New(id)
	if err != nil {
		return nil, err
	}
	if err = box.Init(); err != nil {
		return nil, err
	}
	if err = box.WriteFile("./main.py", code); err != nil {
		return nil, err
	}

	var combinedOutBuf bytes.Buffer
	cmdErr := box.Run("/usr/local/bin/python", []string{"./main.py"},
		sandbox.Network(true),
		sandbox.Stdin(bytes.NewReader(input)),
		sandbox.Stdout(&combinedOutBuf),
		sandbox.Stderr(&combinedOutBuf),
		sandbox.Env(map[string]string{
			"HOME": "/tmp",
			"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		}),
	)

	return combinedOutBuf.Bytes(), cmdErr
}
