package oss

import (
	"path"
	"strconv"
)

const (
	DefaultBatchDir     = "batch"
	DefaultTaskDir      = "task"
	DefaultCodeFileName = "code"
	DefaultTmpDir       = "tmp"
)

func GetBatchDir(batchID int) string {
	return path.Join(DefaultBatchDir, strconv.Itoa(batchID))
}

func GetTaskDir(taskID int) string {
	return path.Join(DefaultTaskDir, strconv.Itoa(taskID))
}

func GetCodePath(taskID int) string {
	return path.Join(GetTaskDir(taskID), DefaultCodeFileName)
}

func GetUserTempDir(uid int) string {
	return path.Join(DefaultTmpDir, strconv.Itoa(uid))
}
