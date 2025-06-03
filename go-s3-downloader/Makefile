# Makefile

.PHONY: build run clean test fmt

# Build the application
build:
	go build -o /usr/local/bin/go-s3-downloader .

# Run the application
run:
	go-s3-downloader -env=$(ENV) -bucket=$(BUCKET)

# Clean build artifacts
clean:
	go clean
	rm -rf bin/

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...