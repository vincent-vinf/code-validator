package main

import (
	"flag"

	"github.com/vincent-vinf/code-validator/pkg/util/config"
)

var (
	configPath = flag.String("config-path", "configs/config.yaml", "")
)

func init() {
	flag.Parse()
}

func main() {
	config.Init(*configPath)
}
