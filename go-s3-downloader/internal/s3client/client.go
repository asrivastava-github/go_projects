package s3client

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"go-s3-downloader/internal/config"
)

type S3Client struct {
	s3 *s3.S3
}

// GetS3Client returns the underlying AWS S3 client
func (c *S3Client) GetS3Client() *s3.S3 {
	return c.s3
}

func NewS3Client(cfg *config.Config) (*S3Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &S3Client{
		s3: s3.New(sess),
	}, nil
}

func (c *S3Client) DownloadFile(bucket, key string) error {
	outputPath := filepath.Join(os.Getenv("HOME"), "s3_files", key)
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := c.s3.GetObjectWithContext(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer result.Body.Close()

	_, err = io.Copy(file, result.Body)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}
