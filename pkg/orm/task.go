package orm

import "time"

type Verification struct {
	ID      int
	BatchID int
	Name    string
	Data    []byte
}

type Task struct {
	ID        int
	UserID    int
	BatchID   int
	Code      string
	CreatedAt time.Time
}

type Batch struct {
	ID            int             `json:"id,omitempty"`
	UserID        int             `json:"userID,omitempty"`
	Name          string          `json:"name,omitempty"`
	CreatedAt     time.Time       `json:"createdAt,omitempty"`
	Verifications []*Verification `json:"verifications,omitempty"`
}
