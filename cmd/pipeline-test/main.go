package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"

	"github.com/vincent-vinf/code-validator/pkg/sandbox"
	"github.com/vincent-vinf/code-validator/pkg/types"
)

func init() {
	flag.Parse()
	rand.Seed(time.Now().Unix())
}

func main() {
	box, err := sandbox.NewIsolate(rand.Int() % sandbox.MaxID)
	if err != nil {
		panic(err)
	}
	e, err := pipeline.NewController(box)
	if err != nil {
		panic(err)
	}

	t := &types.Pipeline{
		Name: "test-task",
		Steps: []*types.Step{
			{
				Name: "init",
				Cmd:  "/bin/ls",
				Args: []string{
					"./",
				},
				MountFiles: []string{
					"default",
				},
			},
			{
				Name: "run",
				Cmd:  "/bin/ls",
				Args: []string{
					"./",
				},
				MountFiles: []string{},
			},
		},
		Files: []*types.File{
			{
				Name: "default",
				Path: "./default",
				Source: types.FileSource{
					Raw: &types.Raw{
						Content: []byte("Vincent\n"),
					},
				},
			},
			{
				Name: "global",
				Path: "./global",
				Source: types.FileSource{
					Raw: &types.Raw{
						Content: []byte("123"),
					},
				},
				Type: types.GlobalFileType,
			},
		},
	}
	if err = e.Exec(t); err != nil {
		return
	}
}
