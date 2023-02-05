package main

import (
	"log"

	"github.com/vincent-vinf/code-validator/pkg/perform-back"
)

func main() {
	t := &perform_back.Task{
		Init: nil,
		Code: perform_back.Code{
			Data: []byte("const readline = require('readline').createInterface({\n    input: process.stdin,\n    output: process.stdout\n});\n\nreadline.question('', response => {\n    console.log(response)\n    readline.close();\n});\n"),
		},
		Verify: perform_back.Validator{
			Default: &perform_back.DefaultValidator{},
		},
		Runtime: perform_back.SupportRuntime,
		Cases: []perform_back.TestCase{
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
	report, err := perform_back.New(5).Run(t)
	if err != nil {
		panic(err)
	}

	log.Println(report.Result)
	for _, c := range report.Cases {
		log.Printf("%+v", c)
	}

}
