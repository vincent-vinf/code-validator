package executor

import (
	"bytes"
	"fmt"
	"io"

	"github.com/vincent-vinf/code-validator/pkg/sandbox"
	"github.com/vincent-vinf/code-validator/pkg/types"
	"github.com/vincent-vinf/code-validator/pkg/util/log"
)

var (
	logger = log.GetLogger()
)

type Executor struct {
	box sandbox.Sandbox
}

func New(box sandbox.Sandbox) (*Executor, error) {
	if box == nil {
		return nil, fmt.Errorf("sandbox must be specified")
	}

	return &Executor{
		box: box,
	}, nil
}

func (e Executor) Exec(task types.Task) error {
	err := e.box.Init()
	if err != nil {
		return err
	}
	//defer func(box sandbox.Sandbox) {
	//	if err := box.Clean(); err != nil {
	//		logger.Errorf("clean sandbox err: %s", err)
	//	}
	//}(e.box)

	for _, file := range task.InputFile {
		if file.Source.Text != nil {
			if err = e.box.WriteFile(file.Path, []byte(file.Source.Text.Content)); err != nil {
				return err
			}
		} else if file.Source.OSS != nil {
			continue
		} else if file.Source.URL != nil {
			continue
		} else {
			logger.Errorf("unknown file(%s) source, skip it", file.Name)

			continue
		}
	}

	logger.Infof("====run task: %s====", task.Name)
	for _, step := range task.Steps {
		logger.Infof("====step: %s====", step.Name)
		var inReader io.Reader
		if step.StdinFile != "" {
			data, err := e.box.ReadFile(step.StdinFile)
			if err != nil {
				return err
			}
			inReader = bytes.NewReader(data)
		}
		var outBuf, errBuf bytes.Buffer

		if err = e.box.Run(step.Cmd, step.Args,
			sandbox.Network(true),
			sandbox.Stdin(inReader),
			sandbox.Stdout(&outBuf),
			sandbox.Stderr(&errBuf),
			sandbox.Env(map[string]string{
				"HOME": "/tmp",
				"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			}),
		); err != nil {
			return err
		}

		for p, data := range map[string][]byte{
			fmt.Sprintf("./std-descriptor-%s/stdout", step.Name): outBuf.Bytes(),
			fmt.Sprintf("./std-descriptor-%s/stderr", step.Name): errBuf.Bytes(),
		} {
			if err = e.box.WriteFile(p, data); err != nil {
				return err
			}
		}
	}

	return nil
}
