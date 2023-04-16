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
	ID        int        `json:"id,omitempty"`
	UserID    int        `json:"userID,omitempty"`
	BatchID   int        `json:"batchID,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	SubTasks  []*SubTask `json:"subTasks,omitempty"`
}

type SubTask struct {
	ID             int `json:"id,omitempty"`
	TaskID         int `json:"taskID,omitempty"`
	VerificationID int `json:"verificationID,omitempty"`
	// todo status
	Status  string `json:"status,omitempty"`
	Result  string `json:"result"`
	Message string `json:"message"`
}
