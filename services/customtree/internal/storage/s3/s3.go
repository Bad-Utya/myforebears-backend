package s3

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
)

type Storage struct {
	client *awss3.Client
	bucket string
}

func New(ctx context.Context, c config.S3Config) (*Storage, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(c.Region), awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")))
	if err != nil {
		return nil, err
	}
	client := awss3.NewFromConfig(cfg, func(o *awss3.Options) {
		o.UsePathStyle = c.ForcePathStyle
		if c.Endpoint != "" {
			o.BaseEndpoint = aws.String(c.Endpoint)
		}
	})
	return &Storage{client, c.Bucket}, nil
}
func (s *Storage) Put(ctx context.Context, key string, data []byte, mime string) error {
	_, err := s.client.PutObject(ctx, &awss3.PutObjectInput{Bucket: &s.bucket, Key: &key, Body: bytes.NewReader(data), ContentType: &mime})
	return err
}
func (s *Storage) Get(ctx context.Context, key string) ([]byte, error) {
	out, err := s.client.GetObject(ctx, &awss3.GetObjectInput{Bucket: &s.bucket, Key: &key})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	b, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}
	return b, nil
}
func (s *Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{Bucket: &s.bucket, Key: &key})
	return err
}
