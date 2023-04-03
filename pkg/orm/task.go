package orm

import "time"

type Verification struct {
	ID      int
	BatchID int
	Name    string
	Runtime string
	Data    string
}

type Batch struct {
	ID            int             `json:"id,omitempty"`
	UserID        int             `json:"userID,omitempty"`
	Name          string          `json:"name,omitempty"`
	Describe      string          `json:"describe,omitempty"`
	Runtime       string          `json:"runtime,omitempty"`
	CreatedAt     time.Time       `json:"createdAt,omitempty"`
	Verifications []*Verification `json:"verifications,omitempty"`
}

type Task struct {
	ID        int
	UserID    int
	BatchID   int
	Code      string
	CreatedAt time.Time
	SubTasks  []*SubTask
}

type SubTask struct {
	ID             int
	TaskID         int
	VerificationID int
	Status         string
	Result         string
	Message        string
}
