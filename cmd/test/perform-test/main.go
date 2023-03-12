package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/vincent-vinf/code-validator/pkg/perform"
	"github.com/vincent-vinf/code-validator/pkg/types"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
	"github.com/vincent-vinf/code-validator/pkg/util/oss"
)

var (
	configPath = flag.String("config-path", "configs/config.yaml", "")
)

func init() {
	flag.Parse()
}

func main() {
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	ossClient, err := oss.NewClient(cfg.Minio)
	if err != nil {
		log.Fatal(err)
	}
	perform.SetOssClient(ossClient)
	vf := &perform.Verification{
		Name:    "123",
		Runtime: types.PythonRuntime,
		Code: &perform.CodeVerification{
			Init:   nil,
			Verify: "code-match match ./output ./answer",
			Cases: []perform.TestCase{
				{
					Name: "1",
					In: perform.File{
						OssPath: "t/1.txt",
					},
					Out: perform.File{
						OssPath: "t/1.txt",
					},
				},
				//{
				//	Name: "2",
				//	In: perform.File{
				//		OssPath: "t/1.txt",
				//	},
				//	Out: perform.File{
				//		OssPath: "t/1.txt",
				//	},
				//},
			},
		},
		//Custom: &perform.CustomVerification{
		//	Action: perform.Action{
		//		Name:    "test",
		//		Command: "pwd",
		//		Files: []perform.File{
		//			{
		//				Path:    "t.in",
		//				OssPath: "t/1.txt",
		//			},
		//		},
		//	},
		//},
	}
	data, err := json.Marshal(vf)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	rep, err := perform.Perform(
		vf, "t/in2out.py")
	if err != nil {
		panic(err)
	}
	log.Println(rep)
}
