package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"go-s3-downloader/internal/awsauth"
	"go-s3-downloader/internal/config"
	"go-s3-downloader/internal/handler"
	"go-s3-downloader/internal/s3client"
)

func main() {
	// Parse command-line arguments
	env := flag.String("env", "", "Environment (prod or dev)")
	profile := flag.String("profile", "", "AWS profile to use (defaults to value of -env if not specified)")
	bucketPrefix := flag.String("bucketPrefix", "", "S3 bucket/prefix to retrieve files from")
	flag.Parse()

	// Validate env argument
	if *env == "" {
		log.Fatalf("Error: env argument is required")
	}

	// Validate bucket/prefix argument
	if *bucketPrefix == "" {
		log.Fatalf("Error: bucket/prefix argument is required")
	}

	// Parse bucket/prefix
	parts := strings.Split(*bucketPrefix, "/")
	if len(parts) < 1 {
		log.Fatalf("Error: invalid bucket/prefix format. Expected 'bucket/prefix'")
	}

	bucket := parts[0]
	prefix := strings.Join(parts[1:], "/")

	// Load configuration
	cfg := config.NewConfig(*env, *profile)
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Check and renew AWS credentials if necessary
	if awsauth.RenewCredentials(cfg.Profile) {
		os.Exit(0) // If credentials were renewed, exit (aws-vault will restart the program)
	}

	// Initialize S3 client
	s3Client, err := s3client.NewS3Client(cfg)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}
	// Download files from S3
	if err := handler.DownloadFiles(s3Client.GetS3Client(), bucket, prefix); err != nil {
		log.Fatalf("Failed to download files: %v", err)
	}

	fmt.Println("Files downloaded successfully.")
}
