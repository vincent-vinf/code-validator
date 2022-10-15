package main

import (
	"bytes"
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/vincent-vinf/code-validator/pkg/sandbox"
)

func init() {
	flag.Parse()
	rand.Seed(time.Now().Unix())
}

func main() {
	s, err := sandbox.NewIsolate(rand.Int() % sandbox.MaxID)
	if err != nil {
		panic(err)
	}
	//defer s.Clean()

	if err = s.Init(); err != nil {
		panic(err)
	}

	err = s.WriteFile("./in", []byte("123"))
	if err != nil {
		panic(err)
	}

	var out, eBuf bytes.Buffer

	err = s.Run("/bin/sh", []string{"-c", "cat < ./in"},
		sandbox.Network(true),
		sandbox.Stdout(&out),
		sandbox.Stderr(&eBuf),
		sandbox.Env(map[string]string{
			"HOME": "/tmp",
			"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		}),
	)
	log.Println("out:", out.String())
	log.Println("err:", eBuf.String())
	if err != nil {
		panic(err)
	}

}
