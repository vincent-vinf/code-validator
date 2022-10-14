package main

import (
	"flag"

	"code-validator/pkg/util/config"
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
