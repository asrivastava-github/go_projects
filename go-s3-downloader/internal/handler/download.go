package handler

import (
	"fmt"
	"os"
	"path/filepath"

	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func DownloadFiles(s3Client *s3.S3, bucket, prefix string) error {
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
	err = s3Client.ListObjectsV2Pages(input, func(result *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, item := range result.Contents {
			if item == nil || item.Key == nil {
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
			err := downloadFileWithPath(s3Client, bucket, key, localPath)
			if err != nil {
				fmt.Printf("Warning: Failed to download %s: %v\n", key, err)
				// Continue downloading other files even if one fails
				continue
			}
		}

		return true // Continue processing pages
	})

	if err != nil {
		return fmt.Errorf("failed to list objects in bucket %s: %w", bucket, err)
	}

	return nil
}

// downloadFileWithPath downloads an S3 object to a specific local file path
func downloadFileWithPath(s3Client *s3.S3, bucket, key, localFilePath string) error {
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

	result, err := s3Client.GetObject(input)
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
