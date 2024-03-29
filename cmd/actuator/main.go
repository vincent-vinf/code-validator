package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/vincent-vinf/code-validator/pkg/orm"
	"github.com/vincent-vinf/code-validator/pkg/perform"
	"github.com/vincent-vinf/code-validator/pkg/types"
	"github.com/vincent-vinf/code-validator/pkg/util"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
	"github.com/vincent-vinf/code-validator/pkg/util/db"
	"github.com/vincent-vinf/code-validator/pkg/util/mq"
	"github.com/vincent-vinf/code-validator/pkg/util/oss"
)

var (
	configPath = flag.String("config-path", "configs/config.yaml", "")
	port       = flag.Int("port", 8001, "")
	log        = logrus.New()
)

func init() {
	flag.Parse()
}

// 缓存Verification
func main() {
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	db.Init(cfg.Mysql)
	defer db.Close()
	ossClient, err := oss.NewClient(cfg.Minio)
	if err != nil {
		log.Fatal(err)
	}
	perform.SetOssClient(ossClient)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
		if err != nil {
			log.Error("listen: ", err)
		}
	}()

	if err = dealQueue(cfg); err != nil {
		log.Fatal(err)
	}
}

func dealQueue(cfg config.Config) error {
	mqClient, err := mq.NewClient(perform.Runtime, cfg.RabbitMQ)
	if err != nil {
		return err
	}
	defer mqClient.Close()
	log.Info("start deal queue")
	if err = mqClient.Consume(
		func(data []byte) {
			req := &types.SubTaskRequest{}
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

func getVerificationByID(id int) (*orm.Verification, error) {
	return db.GetVerificationByID(id)
}

func getTaskByID(id int) (*orm.Task, error) {
	return db.GetTaskByID(id)
}

func subTaskHandle(req *types.SubTaskRequest) error {
	vf, err := getVerificationByID(req.VerificationID)
	if err != nil {
		return err
	}
	task, err := getTaskByID(req.TaskID)
	if err != nil {
		return err
	}

	v := &perform.Verification{}
	if err = json.Unmarshal([]byte(vf.Data), v); err != nil {
		return err
	}

	subtask := &orm.SubTask{
		TaskID:         task.ID,
		VerificationID: vf.ID,
		Status:         types.TaskStatusRunning,
	}
	if err = db.AddSubTask(subtask); err != nil {
		return err
	}
	defer func() {
		subtask.Status = types.TaskStatusFinish
		_ = db.UpdateSubTask(subtask)
	}()

	report, err := perform.Perform(v,
		oss.GetCodePath(req.TaskID),
		oss.GetBatchDir(task.BatchID),
		oss.GetVerificationDir(req.TaskID, req.VerificationID),
	)
	if err != nil {
		subtask.Result = types.TaskStatusFailed
		subtask.Message = err.Error()
		return nil
	}
	util.LogStruct(report)

	if report.Pass {
		subtask.Result = types.TaskStatusSuccess
	} else {
		subtask.Result = types.TaskStatusFailed
	}
	if len(report.Cases) > 0 {
		var passNum int
		util.LogStruct(report.Cases)
		for _, c := range report.Cases {
			if c.Pass {
				passNum++
			}
		}
		subtask.Message = fmt.Sprintf("%d/%d", passNum, len(report.Cases))
	} else {
		subtask.Message = report.Message
	}

	return nil
}
