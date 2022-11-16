package pipeline

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/vincent-vinf/code-validator/pkg/sandbox"
)

const (
	StepOutDir = "step-out"
)

type Executor struct {
	box sandbox.Sandbox

	workdir    string
	stepOutDir string
}

func NewExecutor(id int) (*Executor, error) {
	box, err := sandbox.New(id)
	if err != nil {
		return nil, err
	}
	if err = box.Init(); err != nil {
		return nil, err
	}
	d := box.Workdir()
	e := &Executor{
		box:        box,
		workdir:    d,
		stepOutDir: path.Join(d, StepOutDir),
	}

	if err = os.MkdirAll(e.stepOutDir, 0770); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Executor) Exec(pipeline Pipeline) (*Result, error) {
	// init
	templates := make(map[string]*Template)
	for i := range pipeline.Templates {
		name := pipeline.Templates[i].Name
		if _, ok := templates[name]; ok {
			return nil, fmt.Errorf("duplicate template name: %s", name)
		}
		templates[name] = &pipeline.Templates[i]
	}
	stepNameSet := make(map[string]struct{}, len(pipeline.Steps))
	for i := range pipeline.Steps {
		if _, ok := stepNameSet[pipeline.Steps[i].Name]; ok {
			return nil, fmt.Errorf("duplicate step name: %s", pipeline.Steps[i].Name)
		}
	}
	files := make(map[string]*File)
	for i := range pipeline.Files {
		name := pipeline.Files[i].Name
		if _, ok := files[name]; ok {
			return nil, fmt.Errorf("duplicate file name: %s", name)
		}
		files[name] = &pipeline.Files[i]
	}
	res := &Result{
		Metas: map[string]*sandbox.Meta{},
		Errs:  map[string]error{},
	}
	// run
	for _, step := range pipeline.Steps {
		log.Printf("run step: %s", step.Name)
		var temp *Template
		if t, ok := templates[step.Template]; ok {
			temp = t
		} else {
			return res, fmt.Errorf("template %s does not exist", step.Template)
		}
		var autoRemoveFilePaths []string
		for _, f := range step.RefFiles {
			file, ok := files[f.FileName]
			if !ok {
				return res, fmt.Errorf("file not exist: %s", f.FileName)
			}
			data, err := file.Source.Read()
			if err != nil {
				return res, fmt.Errorf("get file %s, err: %w", f.FileName, err)
			}
			if err = e.box.WriteFile(f.Path, data); err != nil {
				return res, fmt.Errorf("copy file %s, err: %w", f.FileName, err)
			}
			if f.AutoRemove {
				autoRemoveFilePaths = append(autoRemoveFilePaths, f.Path)
			}
		}

		var meta *sandbox.Meta
		if step.NeedMate {
			meta = sandbox.NewMeta()
		}
		inReader, err := e.getInputReader(step.InputFile)
		if err != nil {
			return res, fmt.Errorf("get stdin of step %s, err: %w", step.Name, err)
		}
		var combinedOutBuf bytes.Buffer
		cmdErr := e.box.Run(temp.Cmd, temp.Args,
			sandbox.Network(true),
			sandbox.Stdin(inReader),
			sandbox.Stdout(&combinedOutBuf),
			sandbox.Stderr(&combinedOutBuf),
			sandbox.Metadata(meta),
			sandbox.Env(map[string]string{
				"HOME": "/tmp",
				"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			}),
		)
		if cmdErr != nil {
			res.Errs[step.Name] = cmdErr
		}

		if err = e.writeStepOut(step.Name, combinedOutBuf.Bytes()); err != nil {
			return res, fmt.Errorf("write step out file, err: %w", err)
		}

		if err = e.box.RemoveFile(autoRemoveFilePaths...); err != nil {
			return res, fmt.Errorf("auto remove files err: %w", err)
		}
	}

	return res, nil
}

func (e *Executor) Clean() error {
	if err := e.box.Clean(); err != nil {
		return fmt.Errorf("clean sandbox err: %s", err)
	}

	return nil
}

func (e *Executor) getInputReader(inputFile *InputFile) (io.Reader, error) {
	if inputFile == nil {
		return nil, nil
	}
	var input []byte
	var err error
	switch {
	case inputFile.Source != nil:
		input, err = inputFile.Source.Read()
		if err != nil {
			return nil, err
		}
	case inputFile.StepOut != nil:
		input, err = e.readStepOut(inputFile.StepOut.StepName)
		if err != nil {
			return nil, err
		}
	}

	return bytes.NewReader(input), nil
}

func (e *Executor) writeStepOut(stepName string, data []byte) error {
	return os.WriteFile(path.Join(e.stepOutDir, stepName), data, 0660)
}
func (e *Executor) readStepOut(stepName string) ([]byte, error) {
	return os.ReadFile(path.Join(e.stepOutDir, stepName))
}
