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

func (e *Controller) ReadFile(filepath string) ([]byte, error) {
	return e.box.ReadFile(filepath)
}

func (e *Controller) Exec(pipeline *types.Pipeline) error {
	if pipeline == nil {
		return fmt.Errorf("pipeline cannot be empty")
	}
	err := e.box.Init()
	if err != nil {
		return err
	}

	// mount global files
	for i := range pipeline.Files {
		if pipeline.Files[i] == nil || pipeline.Files[i].Type != types.GlobalFileType {
			continue
		}
		if err := e.mountFile(pipeline.Files[i]); err != nil {
			return err
		}
	}

	logger.Infof("====pipeline: %s====", pipeline.Name)
	for _, step := range pipeline.Steps {
		if step == nil {
			logger.Warnf("nil step in pipeline(%s)", pipeline.Name)
			continue
		}
		logger.Infof("====step: %s====", step.Name)

		// mount file
		files := mountInStepFiles(append(step.MountFiles, step.StdinFile), pipeline.Files)
		for i := range files {
			if err := e.mountFile(files[i]); err != nil {
				return err
			}
		}

		var inReader io.Reader
		if step.StdinFile != "" {
			data, err := e.box.ReadFile(step.StdinFile)
			if err != nil {
				return err
			}
			inReader = bytes.NewReader(data)
		}
		var combinedOutBuf bytes.Buffer

		err = e.box.Run(step.Cmd, step.Args,
			sandbox.Network(true),
			sandbox.Stdin(inReader),
			sandbox.Stdout(&combinedOutBuf),
			sandbox.Stderr(&combinedOutBuf),
			sandbox.Env(map[string]string{
				"HOME": "/tmp",
				"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			}),
		)

		// write logs
		if err := e.box.WriteFile(fmt.Sprintf("./%s.logs", step.Name), combinedOutBuf.Bytes()); err != nil {
			logger.Error(err)
		}
		// remove files
		paths := make([]string, len(files))
		for i := range files {
			paths[i] = files[i].Path
		}

		if err := e.box.RemoveFile(false, paths...); err != nil {
			logger.Error(err)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
func (e *Controller) Logs(steps ...string) []byte {
	if len(steps) == 0 {
		// get all
	} else {
		// get some
	}

	return []byte{}
}
func (e *Controller) Clean() error {
	if err := e.box.Clean(); err != nil {
		return fmt.Errorf("clean sandbox err: %s", err)
	}

	return nil
}

func (e *Controller) mountFile(file *types.File) error {
	if file == nil {
		return nil
	}
	data, err := file.Read()
	if err != nil {
		return err
	}

	return e.box.WriteFile(file.Path, data)
}

func mountInStepFiles(fileNames []string, files []*types.File) (res []*types.File) {
	mountedFilesSet := make(map[string]struct{}, len(fileNames))
	for i := range fileNames {
		mountedFilesSet[fileNames[i]] = struct{}{}
	}
	for i := range files {
		if files[i] == nil {
			continue
		}
		switch files[i].Type {
		case types.GlobalFileType:
		default:
			if _, ok := mountedFilesSet[files[i].Name]; ok {
				res = append(res, files[i])
			}
		}
	}

	return
}
