Certainly! Here's the improved description of the configuration parameters in the `config.json` file, along with an example of how the folder structure looks after running the program. The entire content is in English.

---

# S3 Operations Benchmark System

This project implements a system to generate, replicate, and upload files to an S3 bucket, as well as perform benchmarking operations like GET, HEAD (STAT), and DELETE, similar to MinIO's Warp tool. The goal is to assess the performance of an S3 server under load.

## Table of Contents

- [Description](#description)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
  - [Detailed Configuration Parameters](#detailed-configuration-parameters)
  - [Example Folder Structure](#example-folder-structure)
- [Usage](#usage)
- [Sample Output](#sample-output)
- [License](#license)
- [Contributing](#contributing)

## Description

The program performs the following steps:

1. **Base File Generation**: Creates a set of base files with random content if they don't already exist.
2. **File Replication**: Replicates the base files using `reflink` to save space and time.
3. **Upload to S3**: Uploads the replicated files to an S3 bucket in parallel.
4. **Benchmarking**: Performs GET, HEAD (STAT), and DELETE operations on the stored S3 files for a defined period.
5. **Final Report**: Generates a detailed report with performance metrics of the operations performed.

## Features

- **High Concurrency**: Configure concurrency levels for uploads and benchmarking operations.
- **Parallel Operations**: Utilizes goroutines and channels to perform operations in parallel.
- **Performance Metrics**: Collects and reports minimum, maximum, and average times for each operation type.
- **Flexible Configuration**: Customize program behavior via a `config.json` file.
- **Compatibility**: Works with any server compatible with the S3 API.

## Prerequisites

- **Go**: Version 1.16 or higher installed.
- **AWS SDK for Go**: Can be installed via `go get`.
- **AWS Credentials**: Access key and secret with permissions for S3 operations.
- **Filesystem with Reflink Support**: Such as Btrfs or XFS (optional but recommended).

## Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/ghcoelhopsa/scale_s3_benchmark
   cd scale_s3_benchmark
   ```

2. **Install Dependencies**

   ```bash
   go get github.com/aws/aws-sdk-go
   ```

3. **Build the Program**

   ```bash
   go build
   ```

## Configuration

Create a `config.json` file in the project's root directory with the following format:

```json
{
    "bucketName": "your-bucket-name",
    "s3Folder": "your-s3-base-folder",
    "accessKey": "your-access-key",
    "secretKey": "your-secret-key",
    "baseDirectory": "./base_files",
    "minSize": 1024,
    "maxSize": 2048,
    "maxFilesPerFolder": 1000,
    "baseFileCount": 10,
    "totalFiles": 100,
    "maxConcurrentUploads": 10,
    "maxIdleConns": 100,
    "maxIdleConnsPerHost": 100,
    "httpTimeout": 30,
    "maxRetries": 3,
    "endpointURLs": ["https://your-s3-url"],
    "maxConcurrentReplicas": 5,
    "pauseDurationSeconds": 60,
    "maxLocalFiles": 100,
    "maxBenchmarkThreads": 10,
    "benchmarkDurationSeconds": 60
}
```

### Detailed Configuration Parameters

- **`bucketName`**: *(string)*

  The name of the S3 bucket where the files will be stored. Ensure the bucket already exists and your AWS credentials have the necessary permissions to access it.

- **`s3Folder`**: *(string)*

  The base folder (prefix) within the S3 bucket where files will be uploaded. This allows you to organize your files within the bucket.

- **`accessKey`** and **`secretKey`**: *(string)*

  Your AWS access key ID and secret access key. These credentials should have permissions to perform S3 operations (PUT, GET, HEAD, DELETE) on the specified bucket.

- **`baseDirectory`**: *(string)*

  The local directory where base files will be generated and stored. If the directory doesn't exist, it will be created. This directory will also contain replicated files organized into subfolders.

- **`minSize`** and **`maxSize`**: *(int)*

  The minimum and maximum size (in bytes) for the randomly generated base files. Each base file will have a random size within this range.

- **`maxFilesPerFolder`**: *(int)*

  The maximum number of files to store in each subfolder within the `baseDirectory`. This helps in organizing files and avoiding too many files in a single directory.

- **`baseFileCount`**: *(int)*

  The number of base files to generate. These files will be used as the source for replication.

- **`totalFiles`**: *(int)*

  The total number of files to replicate and upload to S3. The program will replicate the base files to reach this total count.

- **`maxConcurrentUploads`**: *(int)*

  The maximum number of concurrent uploads to S3. Increasing this number can improve upload throughput but may increase resource usage.

- **`maxIdleConns`** and **`maxIdleConnsPerHost`**: *(int)*

  These settings configure the HTTP client's connection pooling. `maxIdleConns` sets the maximum total idle (keep-alive) connections across all hosts, and `maxIdleConnsPerHost` sets the maximum idle connections per host.

- **`httpTimeout`**: *(int)*

  The timeout (in seconds) for HTTP requests made by the S3 client. Adjust this if you experience timeouts due to network latency.

- **`maxRetries`**: *(int)*

  The maximum number of retry attempts for failed S3 upload operations. Retries use exponential backoff.

- **`endpointURLs`**: *(array of strings)*

  A list of S3 endpoint URLs to use. This allows you to specify custom endpoints, such as for S3-compatible storage services. The program will round-robin requests across these endpoints.

- **`maxConcurrentReplicas`**: *(int)*

  The maximum number of concurrent file replication operations. Adjusting this can affect the speed of local file replication.

- **`pauseDurationSeconds`**: *(int)*

  The duration (in seconds) to pause between uploads of subfolders. This can be useful to throttle the upload rate or to introduce delays between batches.

- **`maxLocalFiles`**: *(int)*

  The maximum number of local replicated files to create and reuse during uploads. This helps limit disk space usage by reusing the same set of files for multiple uploads.

- **`maxBenchmarkThreads`**: *(int)*

  The maximum number of concurrent threads to use during benchmarking operations. This controls the level of concurrency for GET, HEAD, and DELETE operations.

- **`benchmarkDurationSeconds`**: *(int)*

  The duration (in seconds) for which each benchmarking phase will run. This defines how long the program will perform GET, HEAD, and DELETE operations.

### Example Folder Structure

After running the program, the local `baseDirectory` and the S3 bucket will have the following folder structures:

#### Local `baseDirectory` Structure:

```
./base_files/
├── file_base_0.txt
├── file_base_1.txt
├── file_base_2.txt
├── ...
├── folder_1/
│   ├── file_0.txt
│   ├── file_1.txt
│   ├── ...
├── folder_2/
│   ├── file_1000.txt
│   ├── file_1001.txt
│   ├── ...
├── ...
```

- **`file_base_X.txt`**: Base files with random content.
- **`folder_X/`**: Subfolders containing replicated files based on the base files.

#### S3 Bucket Structure:

Assuming `s3Folder` is set to `"your-s3-base-folder"` and `bucketName` is `"your-bucket-name"`, the S3 structure will be:

```
s3://your-bucket-name/your-s3-base-folder/
└── your-bucket-name_DDMMYYYYHHMMSS_100/
    ├── folder_1/
    │   ├── file_0.txt
    │   ├── file_1.txt
    │   ├── ...
    ├── folder_2/
    │   ├── file_1000.txt
    │   ├── file_1001.txt
    │   ├── ...
    ├── ...
```

- **`your-bucket-name_DDMMYYYYHHMMSS_100/`**: A subfolder named with a combination of the bucket name, the current date and time (in `DDMMYYYYHHMMSS` format), and the number of local files (`maxLocalFiles`).
- **`folder_X/`**: Subfolders containing the uploaded files, mirroring the local folder structure.

This naming convention helps in organizing uploads and keeping track of when and how many files were uploaded.

## Usage

1. **Prepare the Configuration File**

   Edit the `config.json` file with your desired parameters.

2. **Run the Program**

   ```bash
   ./s3-benchmark
   ```

3. **Monitor the Output**

   The program will display information about the progress of operations, including file generation, upload, and benchmarking.

## Sample Output

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

This project is licensed under the [MIT License](LICENSE).

## Contributing

Contributions are welcome! Feel free to open issues or pull requests in the repository.

---

**Note**: Make sure to update the `LICENSE` file with the appropriate license for your project.

---

If you have any further questions or need additional assistance, please let me know!
