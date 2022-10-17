package main

import (
	"github.com/vincent-vinf/code-validator/pkg/types"
	"github.com/vincent-vinf/code-validator/pkg/util/log"
	"github.com/vincent-vinf/code-validator/pkg/validator"
)

var (
	logger = log.GetLogger()
)

func main() {
	t := &types.Task{
		Init:    types.Step{},
		Run:     types.Code{},
		Verify:  types.Validator{},
		Runtime: types.RuntimeJavaScript,
	}
	v, err := validator.New(1, t)
	if err != nil {
		panic(err)
	}
	report, err := v.Exec()
	if err != nil {
		panic(err)
	}

	logger.Info(report)
}
