package main

import (
	"github.com/vincent-vinf/code-validator/pkg/pp"
	"log"
)

func main() {
	e, err := pp.NewExecutor(5)
	if err != nil {
		log.Println(err)
		return
	}
	defer e.Clean()
	p := pp.Pipeline{
		Name: "pp",
		Steps: []pp.Step{
			{
				Name:     "a",
				Template: "main",
			},
		},
		Templates: []pp.Template{
			{
				Name: "main",
				Cmd:  "/bin/ls",
				Args: []string{
					"/",
				},
			},
		},
		Files: nil,
	}
	if err = e.Exec(p); err != nil {
		log.Println(err)
		return
	}
}
