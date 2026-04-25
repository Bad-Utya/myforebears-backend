package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"

	cfg "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/config"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage struct {
	client *s3.Client
	bucket string
}

func New(ctx context.Context, c cfg.S3Config) (*Storage, error) {
	const op = "storage.s3.New"

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(c.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = &c.Endpoint
		o.UsePathStyle = c.ForcePathStyle
	})

	return &Storage{
		client: client,
		bucket: c.Bucket,
	}, nil
}

func (s *Storage) PutObject(ctx context.Context, key string, content []byte, mimeType string) error {
	const op = "storage.s3.PutObject"

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		Body:        bytes.NewReader(content),
		ContentType: &mimeType,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetObject(ctx context.Context, key string) ([]byte, error) {
	const op = "storage.s3.GetObject"

	obj, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer obj.Body.Close()

	data, err := io.ReadAll(obj.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return data, nil
}

func (s *Storage) DeleteObject(ctx context.Context, key string) error {
	const op = "storage.s3.DeleteObject"

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
