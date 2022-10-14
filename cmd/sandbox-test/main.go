package main

import (
	"bytes"
	"code-validator/pkg/sandbox"
	"flag"
	"log"
	"math/rand"
	"time"
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
	err = s.WriteFile("./t.js", []byte(`
var http = require('http');
var fs = require('fs');

var download = function(url, dest, cb) {
    var file = fs.createWriteStream(dest);
    var request = http.get(url, function(response) {
        response.pipe(file);
        file.on('finish', function() {
            file.close(cb);  // close() is async, call cb after close completes.
        });
    }).on('error', function(err) { // Handle errors
        fs.unlink(dest); // Delete the file async. (But we don't check the result)
        if (cb) cb(err.message);
    });
};

download("http://example.com","./example.html",null)
`))
	if err != nil {
		panic(err)
	}
	var out, eBuf bytes.Buffer

	err = s.Run("/bin/sh", []string{"-c", "npm init -y && npm install ping && node ./t.js"},
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
