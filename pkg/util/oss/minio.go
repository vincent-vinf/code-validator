package oss

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
	"io"
)

type Client struct {
	client *minio.Client
	bucket string
}

func NewClient(config config.Config) (*Client, error) {
	cfg := config.Minio
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		client: minioClient,
		bucket: cfg.Bucket,
	}, nil
}

func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	object, err := c.client.GetObject(ctx, c.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()
	buffer, err := io.ReadAll(object)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func (c *Client) Put(ctx context.Context, path string, data []byte) error {
	_, err := c.client.PutObject(ctx, c.bucket, path, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: "text/plain"})

	return err
}
