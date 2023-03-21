package types

const (
	TaskStatusRunning = "running"
	TaskStatusSuccess = "success"
	TaskStatusFailed  = "failed"
)

type SubTaskRequest struct {
	TaskID         int
	VerificationID int
}
