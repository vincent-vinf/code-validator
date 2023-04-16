package vo

import (
	"github.com/vincent-vinf/code-validator/pkg/orm"
)

type Batch struct {
	orm.Batch
	Username string `json:"username"`
}

type Task struct {
	Username  string `json:"username"`
	BatchName string `json:"batchName"`
	Runtime   string `json:"runtime"`

	orm.Task
	SubTasks []*SubTask `json:"subTasks,omitempty"`
}

type SubTask struct {
	orm.SubTask
	VerificationName string `json:"verificationName"`
}
