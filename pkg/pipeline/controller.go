package pipeline

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"

	"github.com/vincent-vinf/code-validator/pkg/sandbox"
	"github.com/vincent-vinf/code-validator/pkg/util/log"
)

var (
	logger = log.GetLogger()
)

type Controller struct {
	box     sandbox.Sandbox
	tempDir string
}

func NewController(id int) (*Controller, error) {
	box, err := sandbox.New(id)
	if err != nil {
		return nil, err
	}
	p := path.Join(DefaultTempDir, strconv.Itoa(id))
	if err = os.MkdirAll(p, 0755); err != nil {
		return nil, err
	}

	return &Controller{
		box:     box,
		tempDir: p,
	}, nil
}

func (e *Controller) ReadFile(filepath string) ([]byte, error) {
	return e.box.ReadFile(filepath)
}

func (e *Controller) Exec(pipeline *Pipeline) error {
	if pipeline == nil {
		return fmt.Errorf("pipeline cannot be empty")
	}
	err := e.box.Init()
	if err != nil {
		return err
	}

	// copy global files
	for i := range pipeline.Files {
		if pipeline.Files[i] == nil || pipeline.Files[i].Type != GlobalFileType {
			continue
		}
		if err := e.copyFile(pipeline.Files[i]); err != nil {
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

		// copy file
		files := filesToBeCopied(append(step.RefFiles, step.StdinFile), pipeline.Files)
		for i := range files {
			if err := e.copyFile(files[i]); err != nil {
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
		if err := e.writeLogFile(step.Name, combinedOutBuf.Bytes()); err != nil {
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
	// todo
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
	err := os.RemoveAll(e.tempDir)
	if err != nil {
		return err
	}

	return nil
}
func (e *Controller) GetStepLogPath(StepName string) string {
	return path.Join(e.tempDir, StepName, StepLogFile)
}

func (e *Controller) copyFile(file *File) error {
	if file == nil {
		return nil
	}
	data, err := file.Read()
	if err != nil {
		return err
	}

	return e.box.WriteFile(file.Path, data)
}
func (e *Controller) writeLogFile(stepName string, data []byte) error {
	err := os.MkdirAll(path.Join(e.tempDir, stepName), 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(e.GetStepLogPath(stepName), data, 0644)
}
func filesToBeCopied(fileNames []string, files []*File) (res []*File) {
	set := make(map[string]struct{}, len(fileNames))
	for i := range fileNames {
		set[fileNames[i]] = struct{}{}
	}
	for i := range files {
		if files[i] == nil {
			continue
		}
		switch files[i].Type {
		case GlobalFileType:
		default:
			if _, ok := set[files[i].Name]; ok {
				res = append(res, files[i])
			}
		}
	}

	return
}
