package vo

import "github.com/vincent-vinf/code-validator/pkg/orm"

type Batch struct {
	orm.Batch
	Username string `json:"username"`
}

type Task struct {
	orm.Task
	Username string `json:"username"`
}
