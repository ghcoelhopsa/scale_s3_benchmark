// config/config.go
package config

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
)

// Config defines the structure for configuration details loaded from a JSON file.
type Config struct {
    BucketName               string   `json:"bucketName"`              // Name of the S3 bucket.
    S3Folder                 string   `json:"s3Folder"`                // S3 base folder where files will be uploaded.
    AccessKey                string   `json:"accessKey"`               // AWS access key.
    SecretKey                string   `json:"secretKey"`               // AWS secret key.
    BaseDirectory            string   `json:"baseDirectory"`           // Local directory for base files.
    MinSize                  int      `json:"minSize"`                 // Minimum file size for generated files.
    MaxSize                  int      `json:"maxSize"`                 // Maximum file size for generated files.
    MaxFilesPerFolder        int      `json:"maxFilesPerFolder"`       // Maximum number of files per folder.
    BaseFileCount            int      `json:"baseFileCount"`           // Number of base files to generate.
    TotalFiles               int      `json:"totalFiles"`              // Total number of files to upload.
    MaxConcurrentUploads     int      `json:"maxConcurrentUploads"`    // Maximum concurrent uploads to S3.
    MaxIdleConns             int      `json:"maxIdleConns"`            // Maximum number of idle HTTP connections.
    MaxIdleConnsPerHost      int      `json:"maxIdleConnsPerHost"`     // Maximum number of idle connections per host.
    HttpTimeout              int      `json:"httpTimeout"`             // HTTP client timeout in seconds.
    MaxRetries               int      `json:"maxRetries"`              // Maximum retry attempts for S3 uploads.
    EndpointURLs             []string `json:"endpointURLs"`            // List of S3 endpoint URLs.
    MaxConcurrentReplicas    int      `json:"maxConcurrentReplicas"`   // Maximum concurrent file replications.
    PauseDurationSeconds     int      `json:"pauseDurationSeconds"`    // Pause duration between folder uploads.
    MaxLocalFiles            int      `json:"maxLocalFiles"`           // Maximum number of local files to create and reuse.
    MaxBenchmarkThreads      int      `json:"maxBenchmarkThreads"`     // Maximum concurrent threads for benchmarking.
    BenchmarkDurationSeconds int      `json:"benchmarkDurationSeconds"`// Duration for benchmarking operations in seconds.
    MaxConcurrentSubfolders  int      `json:"maxConcurrentSubfolders"` // Maximum number of subfolders to process simultaneously.
}

// LoadConfig loads configuration data from a JSON file.
func LoadConfig(configPath string) (*Config, error) {
    configFile, err := os.Open(configPath)
    if err != nil {
        return nil, fmt.Errorf("error opening config file: %w", err)
    }
    defer configFile.Close()

    byteValue, err := ioutil.ReadAll(configFile)
    if err != nil {
        return nil, fmt.Errorf("error reading config file: %w", err)
    }

    var cfg Config
    if err := json.Unmarshal(byteValue, &cfg); err != nil {
        return nil, fmt.Errorf("error decoding config file: %w", err)
    }

    if cfg.MaxConcurrentReplicas <= 0 {
        return nil, fmt.Errorf("maxConcurrentReplicas must be a positive number, current: %d", cfg.MaxConcurrentReplicas)
    }

    return &cfg, nil
}

