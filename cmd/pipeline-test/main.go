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
				InputRef: &pipeline.DataRef{
					ExternalRef: &pipeline.ExternalRef{
						FileName: "input",
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
		//		SourceCode: ppp.SourceCode{
		//			Raw: &ppp.Raw{Content: []byte("1234")},
		//		},
		//	},
		//},
	}
	_, err = e.Exec(p)
	if err != nil {
		log.Println(err)

		return
	}
}
