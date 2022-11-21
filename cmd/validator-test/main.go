package main

import (
	"log"

	"github.com/vincent-vinf/code-validator/pkg/performer"
)

func main() {
	t := &performer.Task{
		Init: nil,
		Run: performer.Run{
			SourceCode: []byte("const readline = require('readline').createInterface({\n    input: process.stdin,\n    output: process.stdout\n});\n\nreadline.question('', response => {\n    console.log(response)\n    readline.close();\n});\n"),
		},
		Verify: performer.Validator{
			Default: &performer.DefaultValidator{},
		},
		Runtime: performer.SupportRuntime,
		Cases: []performer.TestCase{
			{
				Name:   "c1",
				Input:  []byte("Hello World\n"),
				Output: []byte("Hello World"),
			},
			{
				Name:   "c2",
				Input:  []byte("22222\n"),
				Output: []byte("22222"),
			},
		},
	}
	report, err := performer.New(5).Run(t)
	if err != nil {
		panic(err)
	}

	log.Println(report.Result)
	for _, c := range report.Cases {
		log.Printf("%+v", c)
	}

}
