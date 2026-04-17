package handler

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func DownloadFiles(s3Client *s3.Client, bucket, prefix string) error {
	ctx := context.Background()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Base local directory where files will be stored
	baseDestDir := filepath.Join(homeDir, "s3_files")

	// Create the base directory
	if err := os.MkdirAll(baseDestDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create base destination directory: %w", err)
	}

	// Set up ListObjectsV2 input parameters
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String(""), // No delimiter to get all objects recursively
	}

	// Use pagination to handle large numbers of objects
	paginator := s3.NewListObjectsV2Paginator(s3Client, input)
	for paginator.HasMorePages() {
		result, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects in bucket %s: %w", bucket, err)
		}

		for _, item := range result.Contents {
			if item.Key == nil {
				continue
			}

			key := *item.Key
			// Skip empty "directory" objects (objects with trailing slashes)
			if len(key) > 0 && key[len(key)-1] == '/' {
				continue
			}

			// Use the full key as the relative path
			relPath := key

			// Determine the local file path
			localPath := filepath.Join(baseDestDir, relPath)

			// Download the file
			err := downloadFileWithPath(ctx, s3Client, bucket, key, localPath)
			if err != nil {
				fmt.Printf("Warning: Failed to download %s: %v\n", key, err)
				// Continue downloading other files even if one fails
				continue
			}
		}
	}

	return nil
}

// downloadFileWithPath downloads an S3 object to a specific local file path
func downloadFileWithPath(ctx context.Context, s3Client *s3.Client, bucket, key, localFilePath string) error {
	// Ensure directory exists (create if not)
	dirPath := filepath.Dir(localFilePath)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory structure for %s: %w", key, err)
	}

	// Create or overwrite file
	file, err := os.OpenFile(localFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file %s at %s: %w", key, localFilePath, err)
	}
	defer file.Close()

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s3Client.GetObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to download file %s: %w", key, err)
	}
	defer result.Body.Close()

	_, err = io.Copy(file, result.Body)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	fmt.Printf("Downloaded: %s → %s\n", key, localFilePath)
	return nil
}
