package types

const (
	TaskStatusRunning = "running"
	TaskStatusFinish  = "finish"
	TaskStatusSuccess = "success"
	TaskStatusFailed  = "failed"
)

type SubTaskRequest struct {
	TaskID         int
	VerificationID int
}
