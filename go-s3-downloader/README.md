# Go S3 Downloader

## Overview
Go S3 Downloader is a command-line application that allows users to download files from an Amazon S3 bucket. The application supports two environments: production (`prod`) and development (`dev`). Users can specify the S3 bucket and prefix from which to retrieve files, and the files will be stored in the user's home directory under the `s3_files` directory, maintaining the same folder structure as the prefix.

## Project Structure
```
go-s3-downloader
├── main.go                  # Entry point of the application
├── internal
│   ├── config
│   │   └── config.go        # Configuration settings and loading
│   ├── s3client
│   │   └── client.go        # S3 client wrapper for AWS SDK
│   └── handler
│       └── download.go      # Logic for downloading files from S3
├── go.mod                   # Module dependencies
├── go.sum                   # Module checksums
├── Makefile                 # Build instructions and commands
└── README.md                # Project documentation
```

## Installation
1. Clone the repository:
   ```
   git clone <repository-url>
   cd go-s3-downloader
   ```

2. Build the application:
   ```
   make build
   ```

## Usage
To run the application, use the following command:
```
./bin/go-s3-downloader -env=<environment> -bucket=<bucket/prefix>
```
- `-env`: Specify either `prod` or `dev` (default is `dev`).
- `-bucket`: Specify the S3 bucket and prefix from which to download files.

You can also use Make:
```
make run ENV=<environment> BUCKET=<bucket/prefix>
```

## Example
To download files from the `my-bucket` bucket under the `data/` prefix in production:
```
./bin/go-s3-downloader -env=prod -bucket=my-bucket/data
```

Or using Make:
```
make run ENV=prod BUCKET=my-bucket/data
```

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License
This project is licensed under the MIT License. See the LICENSE file for details.