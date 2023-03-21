package db

import (
	"log"
	"testing"

	"github.com/vincent-vinf/code-validator/pkg/util/config"
)

func TestDB(t *testing.T) {
	cfg, err := config.Load("/Users/vincent/Documents/repo/code-validator/configs/config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	Init(cfg.Mysql)
	defer Close()

	batch, err := GetBatchByIDWithVerifications(46)
	if err != nil {
		panic(err)
	}
	log.Printf("%+v", batch)
	for _, verification := range batch.Verifications {
		log.Printf("%+v", verification)
	}

	task, err := GetTaskByID(3)
	if err != nil {
		panic(err)
	}
	log.Printf("%+v", task)
	for _, s := range task.SubTasks {
		log.Printf("%+v", s)
	}

	//task := &orm.Task{
	//	UserID:    1,
	//	BatchID:   1,
	//	Code:      "xxx",
	//	CreatedAt: time.Now(),
	//}
	//err = AddTask(task)
	//if err != nil {
	//	panic(err)
	//}
	//log.Println(task.ID)
}
