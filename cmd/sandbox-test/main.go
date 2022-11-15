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
	in := []byte("123")
	var out, eBuf bytes.Buffer
	m := &sandbox.Meta{}
	err = s.Run("/bin/sh", []string{"-c", "cat - > t"},
		sandbox.Network(true),
		sandbox.Stdin(bytes.NewReader(in)),
		sandbox.Stdout(&out),
		sandbox.Stderr(&eBuf),
		sandbox.Env(map[string]string{
			"HOME": "/tmp",
			"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		}),
		sandbox.Metadata(m),
	)
	log.Println("out:", out.String())
	log.Println("err:", eBuf.String())
	log.Printf("%+v", m)
	if err != nil {
		panic(err)
	}

	err = s.RemoveFile("t")
	if err != nil {
		panic(err)
	}
}
