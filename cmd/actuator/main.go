package main

import (
	"encoding/json"
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/vincent-vinf/code-validator/pkg/orm"
	"github.com/vincent-vinf/code-validator/pkg/perform"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
	"github.com/vincent-vinf/code-validator/pkg/util/db"
	"github.com/vincent-vinf/code-validator/pkg/util/mq"
	"github.com/vincent-vinf/code-validator/pkg/util/oss"
)

var (
	configPath = flag.String("config-path", "configs/config.yaml", "")
	log        = logrus.New()
)

// 缓存Verification
func main() {
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	db.Init(cfg)
	defer db.Close()
	ossClient, err := oss.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	perform.SetOssClient(ossClient)

	if err = dealQueue(cfg); err != nil {
		log.Fatal(err)
	}
}

func dealQueue(cfg config.Config) error {
	mqClient, err := mq.NewClient("python", cfg.RabbitMQ)
	if err != nil {
		return err
	}
	defer mqClient.Close()

	if err = mqClient.Consume(
		func(data []byte) {
			req := &SubTaskRequest{}
			if err := json.Unmarshal(data, req); err != nil {
				// Ignore abnormal json data
				log.Warn(err)

				return
			}
			if err := subTaskHandle(req); err != nil {
				log.Warn(err)

				return
			}
		},
	); err != nil {
		return err
	}

	return nil
}

type SubTaskRequest struct {
	TaskID         int
	VerificationID int
}

func getVerificationByID(id int) (*orm.Verification, error) {
	return db.GetVerificationByID(id)
}

func getTaskByID(id int) (*orm.Task, error) {
	return db.GetTaskByID(id)
}

func subTaskHandle(req *SubTaskRequest) error {
	vf, err := getVerificationByID(req.VerificationID)
	if err != nil {
		return err
	}
	task, err := getTaskByID(req.TaskID)
	if err != nil {
		return err
	}

	v := &perform.Verification{}
	if err = json.Unmarshal(vf.Data, v); err != nil {
		return err
	}
	report, err := perform.Perform(v, task.Code)
	if err != nil {
		return err
	}

	// todo insert report to db
	log.Info(report)

	return nil
}
