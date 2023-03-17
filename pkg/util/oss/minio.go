package oss

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/vincent-vinf/code-validator/pkg/util/config"
)

const (
	MIMEPlain = "text/plain"
)

type Client struct {
	client *minio.Client
	bucket string
}

func NewClient(cfg config.Minio) (*Client, error) {
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

func (c *Client) Put(ctx context.Context, path string, data io.Reader, len int64, contentType string) error {
	// gin.MIMEPlain
	_, err := c.client.PutObject(ctx, c.bucket, path, data, len, minio.PutObjectOptions{ContentType: contentType})
	logrus.Infof("put file in path(%s) with contentType(%s)", path, contentType)

	return err
}

func (c *Client) Remove(ctx context.Context, path string) error {
	return c.client.RemoveObject(ctx, c.bucket, path, minio.RemoveObjectOptions{})
}

func (c *Client) Copy(ctx context.Context, src, dst string) error {
	_, err := c.client.CopyObject(ctx, minio.CopyDestOptions{
		Bucket: c.bucket,
		Object: dst,
	}, minio.CopySrcOptions{
		Bucket: c.bucket,
		Object: src,
	})

	return err
}

func (c *Client) Move(ctx context.Context, src, dst string) error {
	if err := c.Copy(ctx, src, dst); err != nil {
		return err
	}

	if err := c.Remove(ctx, src); err != nil {
		return fmt.Errorf("failed to remove source file, err: %s", err)
	}

	return nil
}

func (c *Client) PutLocalTextFile(ctx context.Context, path, ossPath string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fstat, err := file.Stat()
	if err != nil {
		return err
	}
	return c.Put(ctx, ossPath, file, fstat.Size(), MIMEPlain)
}

func (c *Client) PutTextFile(ctx context.Context, data []byte, ossPath string) error {
	return c.Put(ctx, ossPath, bytes.NewReader(data), int64(len(data)), MIMEPlain)
}
