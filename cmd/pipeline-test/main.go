package main

import (
	"log"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
)

func main() {
	e, err := pipeline.NewExecutor(5)
	if err != nil {
		log.Println(err)
		return
	}
	//defer e.Clean()
	p := pipeline.Pipeline{
		Name: "pp",
		Steps: []pipeline.Step{
			{
				Name:     "a",
				Template: "cat",
				InputFile: &pipeline.InputFile{
					Source: &pipeline.FileSource{
						Raw: &pipeline.Raw{Content: []byte("1234")},
					},
				},
			},
			{
				Name:     "b",
				Template: "cat",
				InputFile: &pipeline.InputFile{
					StepOut: &pipeline.StepOut{
						StepName: "a",
					},
				},
			},
		},
		Templates: []pipeline.Template{
			{
				Name: "cat",
				Cmd:  "/bin/cat",
				Args: []string{
					"-",
				},
			},
		},
		//Files: []ppp.File{
		//	{
		//		Name: "test",
		//		Source: ppp.FileSource{
		//			Raw: &ppp.Raw{Content: []byte("1234")},
		//		},
		//	},
		//},
	}
	if err = e.Exec(p); err != nil {
		log.Println(err)
		return
	}
}
