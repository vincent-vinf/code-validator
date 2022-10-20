package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"

	"github.com/vincent-vinf/code-validator/pkg/sandbox"
)

func init() {
	flag.Parse()
	rand.Seed(time.Now().Unix())
}

func main() {
	e, err := pipeline.NewController(rand.Int() % sandbox.MaxID)
	if err != nil {
		panic(err)
	}

	t := &pipeline.Pipeline{
		Name: "test-task",
		Steps: []*pipeline.Step{
			{
				Name: "init",
				Cmd:  "/bin/ls",
				Args: []string{
					"./",
				},
				RefFiles: []string{
					"default",
				},
			},
			{
				Name: "run",
				Cmd:  "/bin/ls",
				Args: []string{
					"./",
				},
				RefFiles: []string{},
			},
		},
		Files: []*pipeline.File{
			{
				Name: "default",
				Path: "./default",
				Source: pipeline.FileSource{
					Raw: &pipeline.Raw{
						Content: []byte("Vincent\n"),
					},
				},
			},
			{
				Name: "global",
				Path: "./global",
				Source: pipeline.FileSource{
					Raw: &pipeline.Raw{
						Content: []byte("123"),
					},
				},
				Type: pipeline.GlobalFileType,
			},
		},
	}
	if err = e.Exec(t); err != nil {
		return
	}
}
