package db

import (
	"github.com/vincent-vinf/code-validator/pkg/orm"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
	"testing"
	"time"
)

func TestJson(t *testing.T) {
	cfg, err := config.Load("C:\\Users\\85761\\repo\\code-validator\\configs\\config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	Init(cfg.Mysql)
	defer Close()

	err = AddBatch(&orm.Batch{
		UserID:   1,
		Name:     "test",
		CreateAt: time.Now(),
		Verifications: []*orm.Verification{
			{
				Name: "123",
				Data: []byte("123"),
			},
			{
				Name: "566",
				Data: []byte("456"),
			},
		},
	})
	if err != nil {
		panic(err)
	}
}
