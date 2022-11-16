package pp

import (
	"bytes"
	"fmt"
	"github.com/vincent-vinf/code-validator/pkg/sandbox"
	"log"
)

type Executor struct {
	box sandbox.Sandbox
}

func NewExecutor(id int) (*Executor, error) {
	box, err := sandbox.New(id)
	if err != nil {
		return nil, err
	}

	return &Executor{
		box: box,
	}, nil
}

func (e *Executor) Exec(pipeline Pipeline) error {
	err := e.box.Init()
	if err != nil {
		return err
	}
	templates := make(map[string]*Template)
	for i := range pipeline.Templates {
		name := pipeline.Templates[i].Name
		if _, ok := templates[name]; ok {
			return fmt.Errorf("duplicate template name: %s", name)
		}
		templates[name] = &pipeline.Templates[i]
	}
	for _, step := range pipeline.Steps {
		log.Printf("run step: %s", step.Name)
		var temp *Template
		if t, ok := templates[step.Template]; ok {
			temp = t
		} else {
			return fmt.Errorf("template %s does not exist", step.Template)
		}

		var combinedOutBuf bytes.Buffer
		err = e.box.Run(temp.Cmd, temp.Args,
			sandbox.Network(true),
			//sandbox.Stdin(inReader),
			sandbox.Stdout(&combinedOutBuf),
			sandbox.Stderr(&combinedOutBuf),
			sandbox.Env(map[string]string{
				"HOME": "/tmp",
				"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			}),
		)
	}

	return nil
}

func (e *Executor) Clean() error {
	if err := e.box.Clean(); err != nil {
		return fmt.Errorf("clean sandbox err: %s", err)
	}

	return nil
}
