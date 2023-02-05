package orm

import "time"

type Verification struct {
	ID      int
	BatchID int
	Name    string
	Data    []byte
}

type Task struct {
	ID       int
	UserID   int
	BatchID  int
	Code     string
	CreateAt time.Time
}

type Batch struct {
	ID            int
	UserID        int
	Name          string
	CreateAt      time.Time
	Verifications []*Verification
}
