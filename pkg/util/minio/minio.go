package minio

import (
	"sync"

	"github.com/minio/minio-go/v6"
)

var (
	client *minio.Client
	once   sync.Once
)

func main() {
	_, err := minio.New("play.min.io", "Q3AM3UQ867SPQQA43P2F", "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG", true)
	if err != nil {
		return
	}
}
