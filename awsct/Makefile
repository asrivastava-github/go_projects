.PHONY: build clean test tidy install

BINARY_NAME=awsct
BUILD_DIR=bin

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

install:
	go install .

tidy:
	go mod tidy

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

lint:
	golangci-lint run ./...
