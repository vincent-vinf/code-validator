package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/vincent-vinf/code-validator/pkg/executor"
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
	e, err := executor.New(box)
	if err != nil {
		panic(err)
	}

	scripte := `
const readline = require('readline').createInterface({
    input: process.stdin,
    output: process.stdout
});

readline.question('Who are you?', name => {` +
		"    console.log(`Hey there ${name}!`);" + `
    readline.close();
});
`

	t := types.Task{
		Name: "test-task",
		Steps: []types.Step{
			{
				Name: "init",
				Cmd:  "/bin/sh",
				Args: []string{
					"-c",
					"npm init -y",
				},
			},
			{
				Name: "run",
				Cmd:  "/usr/local/bin/node",
				Args: []string{
					"./script",
				},
				StdinFile: "./test-file",
			},
		},
		InputFile: []types.File{
			{
				Name: "test-file",
				Path: "./test-file",
				Source: types.FileSource{
					Text: &types.Text{
						Content: "Vincent\n",
					},
				},
			},
			{
				Name: "script",
				Path: "./script",
				Source: types.FileSource{
					Text: &types.Text{
						Content: scripte,
					},
				},
			},
		},
	}
	if err = e.Exec(t); err != nil {
		return
	}
}
