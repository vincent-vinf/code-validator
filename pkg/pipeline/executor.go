package pipeline

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

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
		if name == "" {
			return nil, errors.New("template name cannot be empty")
		}
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
		if step.InlineTemplate != nil {
			temp = step.InlineTemplate
		} else {
			if t, ok := templates[step.Template]; ok {
				temp = t
			} else {
				return res, fmt.Errorf("template %s does not exist", step.Template)
			}
		}

		var autoRemoveFilePaths []string
		for _, f := range step.FileRefs {
			data, err := e.readDataRef(f.DataRef, files)
			if err != nil {
				return res, fmt.Errorf("get file data err: %w", err)
			}

			if err = e.box.WriteFile(f.Path, data); err != nil {
				return res, fmt.Errorf("copy file %s, err: %w", f.Path, err)
			}
			if f.AutoRemove {
				autoRemoveFilePaths = append(autoRemoveFilePaths, f.Path)
			}
		}

		var meta *sandbox.Meta
		if step.LogMate {
			meta = sandbox.NewMeta()
			res.Metas[step.Name] = meta
		}

		var input []byte
		if step.InputRef != nil {
			var err error
			input, err = e.readDataRef(*step.InputRef, files)
			if err != nil {
				return res, fmt.Errorf("get stdin of step %s, err: %w", step.Name, err)
			}
		}

		network := true
		var timeout time.Duration
		if step.Limit != nil {
			network = step.Limit.EnableNetWork
			timeout = step.Limit.Time
		}
		var combinedOutBuf bytes.Buffer
		cmdErr := e.box.Run(temp.Cmd, temp.Args,
			sandbox.Network(network),
			sandbox.Stdin(bytes.NewReader(input)),
			sandbox.Stdout(&combinedOutBuf),
			sandbox.Time(timeout),
			sandbox.Stderr(&combinedOutBuf),
			sandbox.Metadata(meta),
			sandbox.Env(map[string]string{
				"HOME": "/tmp",
				"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			}),
		)
		if err := e.writeStepOut(step.Name, combinedOutBuf.Bytes()); err != nil {
			return res, fmt.Errorf("write step out file, err: %w", err)
		}
		if cmdErr != nil {
			res.Errs[step.Name] = cmdErr

			if !step.ContinueOnFail {
				return res, cmdErr
			}
		}

		if err := e.box.RemoveFile(autoRemoveFilePaths...); err != nil {
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

func (e *Executor) StepOutDir() string {
	return e.stepOutDir
}

func (e *Executor) ReadFile(path string) ([]byte, error) {
	return e.box.ReadFile(path)
}

func (e *Executor) readDataRef(ref DataRef, files map[string]*File) ([]byte, error) {
	switch {
	case ref.ExternalRef != nil:
		f, ok := files[ref.ExternalRef.FileName]
		if !ok {
			return nil, fmt.Errorf("file not exist: %s", ref.ExternalRef.FileName)
		}
		return f.Content, nil
	case ref.StepOutRef != nil:
		return e.readStepOut(ref.StepOutRef.StepName)
	default:
		return nil, errors.New("data ref not specified")
	}
}

func (e *Executor) writeStepOut(stepName string, data []byte) error {
	return os.WriteFile(path.Join(e.stepOutDir, stepName), data, 0660)
}
func (e *Executor) readStepOut(stepName string) ([]byte, error) {
	return os.ReadFile(path.Join(e.stepOutDir, stepName))
}
