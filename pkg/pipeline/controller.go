package pipeline

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

type Controller struct {
	box sandbox.Sandbox
}

func NewController(box sandbox.Sandbox) (*Controller, error) {
	if box == nil {
		return nil, fmt.Errorf("sandbox must be specified")
	}

	return &Controller{
		box: box,
	}, nil
}

func (e Controller) Exec(pipeline types.Pipeline) error {
	err := e.box.Init()
	if err != nil {
		return err
	}
	// todo clean
	//defer func(box sandbox.Sandbox) {
	//	if err := box.Clean(); err != nil {
	//		logger.Errorf("clean sandbox err: %s", err)
	//	}
	//}(e.box)

	for _, file := range pipeline.InputFile {
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

	logger.Infof("====run pipeline: %s====", pipeline.Name)
	for _, step := range pipeline.Steps {
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
			fmt.Sprintf("./std-%s/stdout", step.Name): outBuf.Bytes(),
			fmt.Sprintf("./std-%s/stderr", step.Name): errBuf.Bytes(),
		} {
			if err = e.box.WriteFile(p, data); err != nil {
				return err
			}
		}
	}

	return nil
}
