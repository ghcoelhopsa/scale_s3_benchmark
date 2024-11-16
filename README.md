# IBM Scale S3 Benchmark (nooba)
## Overview
This Go-based application is designed to generate and upload large numbers of small files to an IBM Scale S3 Benchmark (nooba) service. It includes functionality for file generation, benchmarking, replication, and upload management. The program is intended for performance testing of S3-compatible storage, particularly when handling numerous small files.

**"The web part with real-time updates is not yet functional. Disabled."**

## Features
- **File Generation**: Generates files between 4KB and 8KB in size for testing purposes. File counts can be configured to scale from a thousand to several million files.
- **S3 Upload**: Uploads generated files to a specified S3 bucket, with support for parallel uploads and configurable concurrency levels.
- **Replication**: Supports replicating files to multiple S3 endpoints.
- **Benchmarking**: Includes benchmarking capabilities to assess the performance of different operations on the files and S3 storage, such as GET, HEAD, and DELETE.
- **Final Report**: Generates a detailed report with metrics of the performed operations.

## Configuration
The application uses a configuration file, `config.json`, to define all operational parameters. Here are key properties from the configuration file:

- **S3 Settings**:
  - `bucketName`: The name of the S3 bucket where files will be uploaded.
  - `s3Folder`: Base folder in the S3 bucket for the uploads.
  - `accessKey` and `secretKey`: Credentials for accessing the S3 service.
- **File Generation Settings**:
  - `baseDirectory`: Local directory used to store generated files.
  - `minSize` and `maxSize`: File size range for generated files, in bytes (4KB to 8KB).
  - `baseFileCount` and `totalFiles`: Controls the number of files generated and uploaded.
- **Upload and Concurrency Settings**:
  - `maxConcurrentUploads`: Number of concurrent upload operations allowed.
  - `maxConcurrentReplicas`: Number of concurrent replica operations.
  - `maxConcurrentSubfolders`: Maximum number of concurrent subfolder operations.
  - `maxRetries`: Number of retries for failed operations.
- **HTTP Settings**:
  - `maxIdleConns` and `maxIdleConnsPerHost`: HTTP connection pooling parameters.
  - `httpTimeout`: Timeout for HTTP requests, in seconds.
  - `pauseDurationSeconds`: Pause duration between retries for failed uploads.
- **Benchmark Settings**:
  - `maxBenchmarkThreads`: Number of threads for benchmarking.
  - `benchmarkDurationSeconds`: Duration of benchmarking runs.

The `config.json` file plays a crucial role in defining how the application will behave. By adjusting the parameters, users can control aspects like the number of files generated, their sizes, the concurrency level for uploads, and the S3 credentials required for access. This flexibility allows for tailored performance testing based on specific requirements.

## File Structure
- **main.go**: Entry point of the application.
- **config.json**: Configuration settings for the application.
- **report.go**: Manages the generation of reports related to upload and replication statistics.
- **benchmark.go**: Handles benchmarking operations.
- **config.go**: Loads and manages application configuration.
- **replication.go**: Manages file replication to multiple S3 endpoints.
- **file_generation.go**: Handles the creation of files for testing.
- **upload.go**: Manages the upload process for generated files.
- **s3_client.go**: Contains functions for interacting with the S3-compatible API.

The structure of the project is organized to facilitate easy management and extension of functionality. Each major component of the application is separated into its own Go file:

- **`main.go`**: The central entry point where the application starts execution.
- **`config.json`**: Stores configuration parameters, enabling users to adjust settings without modifying code.
- **`config.go`**: Responsible for loading and managing configurations from `config.json`.
- **`file_generation.go`**: Focuses on generating test files with specified properties, such as size and count.
- **`upload.go`**: Implements the logic for uploading generated files to the specified S3 bucket, including managing concurrency.
- **`replication.go`**: Handles the replication of files to multiple endpoints, useful for redundancy or multi-region testing.
- **`s3_client.go`**: Provides helper functions for interacting with S3-compatible APIs, such as initiating uploads, handling retries, and more.
- **`benchmark.go`**: Manages benchmarking tasks to evaluate the performance of S3 operations like GET, HEAD, and DELETE.
- **`report.go`**: Generates detailed reports summarizing the performance metrics from uploads, replications, and benchmarking operations.

## Setup Instructions
1. **Prerequisites**: Make sure to have Go installed on your system and set up the necessary AWS/S3-compatible credentials.
2. **Clone the Repository**: Download the code to your local environment.
   ```bash
   git clone https://github.com/ghcoelhopsa/scale_s3_benchmark
   cd scale_s3_benchmark
   ```
3. **Install Dependencies**
   ```bash
   go get github.com/aws/aws-sdk-go
   ```
4. **Build the Program**
   ```bash
   go build -o s3-benchmark
   ```
5. **Configure the Application**: Edit the `config.json` file to set your S3 credentials, bucket name, file size, and concurrency settings.

## Usage
- **Prepare the Configuration File**: Modify `config.json` to adjust the parameters as needed. You can change the number of files, file sizes, bucket names, and concurrency levels.
- **Run the Program**:
  ```sh
  ./s3-benchmark
  ```
  This will start generating files, uploading them to the specified S3 bucket, and running any specified benchmarks.
- **Monitor the Output**: The program will display information about the progress of operations, including file generation, upload, replication, and benchmarking.

## Example Output
```
Starting pprof server on port 6060
100% completed - 10 base files generated.
Starting file replication with reflink in parallel.
Replicating files: 100/100 (100.00%) | Rate: 5000.00 files/sec
File replication completed.
Replication of 100 files completed.
Uploading to S3: 100/100 (100.00%) | Rate: 2000.00 files/sec
Upload completed for subfolder index 0.
Pausing for 60 seconds before next upload...
All uploads completed.
Performing benchmarking operations...
GET and STAT operations completed. Starting DELETE operations...

Benchmarking Report:
====================

Operation: GET
Total Operations: 5000
Successes: 5000
Errors: 0
Min Time: 5ms
Max Time: 50ms
Avg Time: 25ms

Operation: STAT
Total Operations: 5000
Successes: 5000
Errors: 0
Min Time: 3ms
Max Time: 40ms
Avg Time: 20ms

Operation: DELETE
Total Operations: 100
Successes: 100
Errors: 0
Min Time: 10ms
Max Time: 30ms
Avg Time: 15ms

Overall Benchmark Summary:
Total Operations: 10100
Total Errors: 0
Benchmarking Duration: 1m0s
====================
```

## License
This software is provided as-is under the MIT License. Please refer to the LICENSE file for additional information.

## Contribution
Feel free to open issues and submit pull requests for new features or bug fixes. All contributions are welcome.
