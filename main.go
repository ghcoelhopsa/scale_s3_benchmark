// Written by Gabriel H. Coelho 11/2024 v1.0
// Package main implements a system for managing files and uploading them to AWS S3 in parallel, and performs benchmarking operations.
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "math"
    "math/rand"
    "net/http"
    _ "net/http/pprof" // Enables pprof for profiling the performance of the application.
    "os"
    "os/exec"
    "path/filepath"
    "sync"
    "sync/atomic"
    "syscall"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

// Config defines the structure for configuration details loaded from a JSON file.
type Config struct {
    BucketName              string   `json:"bucketName"`              // Name of the S3 bucket.
    S3Folder                string   `json:"s3Folder"`                // S3 base folder where files will be uploaded.
    AccessKey               string   `json:"accessKey"`               // AWS access key.
    SecretKey               string   `json:"secretKey"`               // AWS secret key.
    BaseDirectory           string   `json:"baseDirectory"`           // Local directory for base files.
    MinSize                 int      `json:"minSize"`                 // Minimum file size for generated files.
    MaxSize                 int      `json:"maxSize"`                 // Maximum file size for generated files.
    MaxFilesPerFolder       int      `json:"maxFilesPerFolder"`       // Maximum number of files per folder.
    BaseFileCount           int      `json:"baseFileCount"`           // Number of base files to generate.
    TotalFiles              int      `json:"totalFiles"`              // Total number of files to replicate.
    MaxConcurrentUploads    int      `json:"maxConcurrentUploads"`    // Maximum concurrent uploads to S3.
    MaxIdleConns            int      `json:"maxIdleConns"`            // Maximum number of idle HTTP connections.
    MaxIdleConnsPerHost     int      `json:"maxIdleConnsPerHost"`     // Maximum number of idle connections per host.
    HttpTimeout             int      `json:"httpTimeout"`             // HTTP client timeout in seconds.
    MaxRetries              int      `json:"maxRetries"`              // Maximum retry attempts for S3 uploads.
    EndpointURLs            []string `json:"endpointURLs"`            // List of S3 endpoint URLs.
    MaxConcurrentReplicas   int      `json:"maxConcurrentReplicas"`   // Maximum concurrent file replications.
    PauseDurationSeconds    int      `json:"pauseDurationSeconds"`    // Pause duration between subfolder uploads.
    MaxLocalFiles           int      `json:"maxLocalFiles"`           // Maximum number of local files to create and reuse.
    MaxBenchmarkThreads     int      `json:"maxBenchmarkThreads"`     // Maximum concurrent threads for benchmarking.
    BenchmarkDurationSeconds int     `json:"benchmarkDurationSeconds"`// Duration for benchmarking operations in seconds.
}

// Global variables used throughout the application.
var (
    config                  Config
    successCount            int64              // Counter for successful uploads.
    s3ClientIndex           uint64             // Index used to round-robin between S3 clients.
    uploadedS3Files         []string           // Tracks successfully uploaded S3 files.
    uploadWG                sync.WaitGroup     // WaitGroup to manage upload tasks.
    startTime               time.Time          // Global start time for rate calculations.
    benchmarkMetrics        map[OperationType]*PerformanceMetrics // Metrics for benchmarking.
    actualBenchmarkDuration time.Duration      // Actual duration of benchmarking operations.
)

// PerformanceMetrics holds the metrics for benchmarking operations.
type PerformanceMetrics struct {
    TotalOperations int64
    TotalTime       time.Duration
    MinTime         time.Duration
    MaxTime         time.Duration
    ErrorCount      int64
}

// OperationType defines the type of S3 operation.
type OperationType string

const (
    OperationGet    OperationType = "GET"
    OperationDelete OperationType = "DELETE"
    OperationStat   OperationType = "STAT"
)

// loadConfig loads configuration data from a JSON file.
func loadConfig(configPath string) error {
    configFile, err := os.Open(configPath)
    if err != nil {
        return fmt.Errorf("error opening config file: %w", err)
    }
    defer configFile.Close()

    byteValue, err := ioutil.ReadAll(configFile)
    if err != nil {
        return fmt.Errorf("error reading config file: %w", err)
    }

    if err := json.Unmarshal(byteValue, &config); err != nil {
        return fmt.Errorf("error decoding config file: %w", err)
    }

    if config.MaxConcurrentReplicas <= 0 {
        return fmt.Errorf("maxConcurrentReplicas must be a positive number, current: %d", config.MaxConcurrentReplicas)
    }

    return nil
}

// main is the entry point of the application.
func main() {
    // Load configuration.
    if err := loadConfig("config.json"); err != nil {
        fmt.Printf("Error loading config: %v\n", err)
        os.Exit(1)
    }

    // Initialize startTime
    startTime = time.Now()

    // Seed the random number generator.
    rand.Seed(time.Now().UnixNano())

    // Start a pprof server for profiling.
    go func() {
        fmt.Println("Starting pprof server on port 6060")
        if err := http.ListenAndServe("0.0.0.0:6060", nil); err != nil {
            fmt.Printf("Error starting pprof server: %v\n", err)
            os.Exit(1)
        }
    }()

    // Increase file descriptor limit for handling many files.
    if err := increaseFileDescriptorLimit(); err != nil {
        fmt.Printf("Error adjusting file descriptor limits: %v\n", err)
        return
    }

    // Prepare the base directory for generating files.
    if err := prepareBaseDirectory(); err != nil {
        fmt.Printf("Error preparing base directory: %v\n", err)
        return
    }

    // Generate the base set of files (if they don't exist).
    generateAllBaseFiles()

    // Replicate files locally up to the file limit only once.
    localFiles, err := replicateFilesWithReflinkInParallel(config.MaxLocalFiles)
    if err != nil {
        fmt.Printf("Error replicating files with reflink: %v\n", err)
        return
    }

    fmt.Printf("Replication of %d files completed.\n", len(localFiles))

    // Calculate the number of S3 subfolders needed.
    numSubfolders := int(math.Ceil(float64(config.TotalFiles) / float64(config.MaxLocalFiles)))

    totalFilesUploaded := 0

    for subfolderIndex := 0; subfolderIndex < numSubfolders; subfolderIndex++ {
        filesToProcess := config.MaxLocalFiles
        if config.TotalFiles-totalFilesUploaded < config.MaxLocalFiles {
            filesToProcess = config.TotalFiles - totalFilesUploaded
        }

        // Create a context for managing the upload process.
        ctx, cancel := context.WithCancel(context.Background())

        var wg sync.WaitGroup
        fileChan := make(chan string, 1000)

        // Start uploading files to S3 in parallel.
        uploadToS3InParallel(ctx, &wg, fileChan, subfolderIndex)

        // Send replicated files to the upload channel.
        wg.Add(1)
        go func() {
            defer wg.Done()
            for i := 0; i < filesToProcess; i++ {
                filePath := localFiles[i%len(localFiles)]
                fileChan <- filePath
            }
            close(fileChan)
        }()

        wg.Wait()
        cancel()

        fmt.Printf("\nUpload completed for subfolder index %d.\n", subfolderIndex)

        // Pause between subfolder uploads as per configuration.
        fmt.Printf("Pausing for %d seconds before next upload...\n", config.PauseDurationSeconds)
        time.Sleep(time.Duration(config.PauseDurationSeconds) * time.Second)

        totalFilesUploaded += filesToProcess

        if totalFilesUploaded >= config.TotalFiles {
            break
        }
    }

    fmt.Println("\nAll uploads completed.")

    // Clean up local files to free up space.
    cleanupLocalFiles(localFiles)

    // Perform benchmarking operations
    performBenchmarkOperations()

    // Generate final report
    generateFinalReport()
}

// increaseFileDescriptorLimit increases the file descriptor limit to handle more open files.
func increaseFileDescriptorLimit() error {
    var rLimit syscall.Rlimit
    if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err == nil {
        rLimit.Cur = rLimit.Max
        return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
    } else {
        return err
    }
}

// prepareBaseDirectory prepares the base directory by creating it if it doesn't exist.
func prepareBaseDirectory() error {
    return os.MkdirAll(config.BaseDirectory, os.ModePerm)
}

// generateAllBaseFiles generates a specified number of base files with random content.
func generateAllBaseFiles() {
    startTime := time.Now()
    for i := 0; i < config.BaseFileCount; i++ {
        filename := filepath.Join(config.BaseDirectory, fmt.Sprintf("file_base_%d.txt", i))

        // Check if the file already exists.
        if _, err := os.Stat(filename); os.IsNotExist(err) {
            if err := generateTextFile(filename); err != nil {
                fmt.Printf("Error generating base file %s: %v\n", filename, err)
            }
        }

        elapsedTime := time.Since(startTime).Seconds()
        rate := float64(i+1) / elapsedTime
        progress := float64(i+1) / float64(config.BaseFileCount) * 100
        fmt.Printf("Generating base files: %d/%d (%.2f%%) | Rate: %.2f files/sec\r", i+1, config.BaseFileCount, progress, rate)
    }
    fmt.Printf("100%% completed - %d base files generated.\n", config.BaseFileCount)
}

// generateTextFile creates a text file with random alphabetical content of a specified size.
func generateTextFile(filename string) error {
    size := rand.Intn(config.MaxSize-config.MinSize+1) + config.MinSize
    content := make([]byte, size)
    for i := range content {
        content[i] = byte('a' + rand.Intn(26))
    }
    return os.WriteFile(filename, content, 0644)
}

// replicateFilesWithReflinkInParallel replicates files using reflink in parallel.
func replicateFilesWithReflinkInParallel(totalFiles int) ([]string, error) {
    fmt.Println("Starting file replication with reflink in parallel.")
    startTime := time.Now()

    var replicatedFiles []string
    var mu sync.Mutex
    var replicationWG sync.WaitGroup

    jobs := make(chan int, totalFiles)

    replicationWG.Add(config.MaxConcurrentReplicas)
    for w := 0; w < config.MaxConcurrentReplicas; w++ {
        go func() {
            defer replicationWG.Done()
            for currentCount := range jobs {
                folderNumber := (currentCount / config.MaxFilesPerFolder) + 1
                folderPath := filepath.Join(config.BaseDirectory, fmt.Sprintf("folder_%d", folderNumber))
                os.MkdirAll(folderPath, os.ModePerm)

                baseFileIndex := currentCount % config.BaseFileCount
                src := filepath.Join(config.BaseDirectory, fmt.Sprintf("file_base_%d.txt", baseFileIndex))
                dst := filepath.Join(folderPath, fmt.Sprintf("file_%d.txt", currentCount))

                if err := copyFileReflink(src, dst); err != nil {
                    fmt.Printf("\nError replicating file %s to %s: %v\n", src, dst, err)
                    continue
                }

                mu.Lock()
                replicatedFiles = append(replicatedFiles, dst)
                mu.Unlock()

                elapsedTime := time.Since(startTime).Seconds()
                rate := float64(len(replicatedFiles)) / elapsedTime
                progress := float64(len(replicatedFiles)) / float64(totalFiles) * 100
                fmt.Printf("Replicating files: %d/%d (%.2f%%) | Rate: %.2f files/sec\r", len(replicatedFiles), totalFiles, progress, rate)
            }
        }()
    }

    for currentCount := 0; currentCount < totalFiles; currentCount++ {
        jobs <- currentCount
    }
    close(jobs)

    replicationWG.Wait()
    fmt.Println("\nFile replication completed.")
    return replicatedFiles, nil
}

// copyFileReflink copies a file using reflink, falling back to a regular copy if reflink fails.
func copyFileReflink(src, dst string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctx, "cp", "--reflink=always", src, dst)

    _, err := cmd.CombinedOutput()
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("cp command timed out")
    }

    if err != nil {
        return fallbackCopy(src, dst) // Fallback to a regular copy if reflink fails.
    }
    return nil
}

// fallbackCopy performs a traditional file copy if reflink is not supported.
func fallbackCopy(src, dst string) error {
    srcFile, err := os.Open(src)
    if err != nil {
        return fmt.Errorf("error opening source file %s: %w", src, err)
    }
    defer srcFile.Close()

    dstFile, err := os.Create(dst)
    if err != nil {
        return fmt.Errorf("error creating destination file %s: %w", dst, err)
    }
    defer dstFile.Close()

    if _, err := io.Copy(dstFile, srcFile); err != nil {
        return fmt.Errorf("error copying file from %s to %s: %w", src, dst, err)
    }
    return nil
}

// uploadToS3InParallel uploads files to S3 in parallel.
func uploadToS3InParallel(ctx context.Context, wg *sync.WaitGroup, fileChan <-chan string, subfolderIndex int) {
    var s3Clients []*s3.S3
    for _, endpoint := range config.EndpointURLs {
        sess, err := session.NewSession(&aws.Config{
            Region:           aws.String("us-east-1"),
            Endpoint:         aws.String(endpoint),
            Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
            S3ForcePathStyle: aws.Bool(true),
            HTTPClient: &http.Client{
                Transport: &http.Transport{
                    MaxIdleConns:        config.MaxIdleConns,
                    MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
                },
                Timeout: time.Duration(config.HttpTimeout) * time.Second,
            },
        })

        if err != nil {
            fmt.Printf("Error creating S3 session for endpoint %s: %v\n", endpoint, err)
            continue
        }

        s3Clients = append(s3Clients, s3.New(sess))
    }

    if len(s3Clients) == 0 {
        fmt.Println("No S3 clients were created. Check endpoints and credentials.")
        return
    }

    semaphore := make(chan struct{}, config.MaxConcurrentUploads)

    uploadWG.Add(config.MaxConcurrentUploads)
    for i := 0; i < config.MaxConcurrentUploads; i++ {
        go func() {
            defer uploadWG.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                case filePath, ok := <-fileChan:
                    if !ok {
                        return
                    }
                    semaphore <- struct{}{}
                    uploadFileWithRetry(s3Clients, filePath, subfolderIndex)
                    <-semaphore
                }
            }
        }()
    }

    wg.Add(1)
    go func() {
        defer wg.Done()
        uploadWG.Wait()
    }()
}

// uploadFileWithRetry attempts to upload a file to S3, retrying on failure.
func uploadFileWithRetry(s3Clients []*s3.S3, filePath string, subfolderIndex int) error {
    // Build the s3Key with the specified naming convention.
    dateTimeStr := time.Now().Format("02012006150405") // DDMMYYYYHHMMSS
    quantityOfFilesStr := fmt.Sprintf("%d", config.MaxLocalFiles)
    subfolderName := fmt.Sprintf("%s_%s_%s", config.BucketName, dateTimeStr, quantityOfFilesStr)

    // S3 key structure: s3Folder/subfolderName/folderName/fileName
    folderName := filepath.Base(filepath.Dir(filePath))
    fileName := filepath.Base(filePath)
    s3Key := filepath.Join(config.S3Folder, subfolderName, folderName, fileName)

    for attempt := 1; attempt <= config.MaxRetries; attempt++ {
        if err := uploadFile(s3Clients, filePath, s3Key); err == nil {
            atomic.AddInt64(&successCount, 1)

            elapsedTime := time.Since(startTime).Seconds()
            rate := float64(successCount) / elapsedTime
            progress := float64(successCount) / float64(config.TotalFiles) * 100
            fmt.Printf("Uploading to S3: %d/%d (%.2f%%) | Rate: %.2f files/sec\r", successCount, config.TotalFiles, progress, rate)

            // Store uploaded S3 key
            uploadedS3Files = append(uploadedS3Files, s3Key)

            return nil
        } else if attempt < config.MaxRetries {
            backoffDuration := time.Duration(math.Pow(2, float64(attempt))) * time.Second
            time.Sleep(backoffDuration) // Exponential backoff before retrying.
        } else {
            fmt.Printf("\nFailed to upload %s after %d attempts\n", filePath, config.MaxRetries)
            return fmt.Errorf("failed to upload %s after %d attempts", filePath, config.MaxRetries)
        }
    }

    return fmt.Errorf("failed to upload %s after %d attempts", filePath, config.MaxRetries)
}

// uploadFile uploads a single file to S3 using a selected S3 client.
func uploadFile(s3Clients []*s3.S3, filePath, s3Key string) error {
    clientIndex := atomic.AddUint64(&s3ClientIndex, 1)
    s3Client := s3Clients[clientIndex%uint64(len(s3Clients))]

    fileData, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("error opening file %s: %w", filePath, err)
    }
    defer fileData.Close()

    _, err = s3Client.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(config.BucketName),
        Key:    aws.String(s3Key),
        Body:   fileData,
    })
    return err
}

// cleanupLocalFiles removes the replicated local files to free up space.
func cleanupLocalFiles(files []string) {
    for _, filePath := range files {
        os.Remove(filePath)
    }
}

// performBenchmarkOperations performs GET, STAT, and DELETE operations for benchmarking.
func performBenchmarkOperations() {
    fmt.Println("\nPerforming benchmarking operations...")

    // Prepare S3 client
    sess, err := session.NewSession(&aws.Config{
        Region:           aws.String("us-east-1"),
        Endpoint:         aws.String(config.EndpointURLs[0]),
        Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
        S3ForcePathStyle: aws.Bool(true),
        HTTPClient: &http.Client{
            Transport: &http.Transport{
                MaxIdleConns:        config.MaxIdleConns,
                MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
            },
            Timeout: time.Duration(config.HttpTimeout) * time.Second,
        },
    })

    if err != nil {
        fmt.Printf("Error creating S3 session: %v\n", err)
        return
    }

    s3Client := s3.New(sess)

    // Prepare metrics storage
    metrics := map[OperationType]*PerformanceMetrics{
        OperationGet:    &PerformanceMetrics{},
        OperationStat:   &PerformanceMetrics{},
        OperationDelete: &PerformanceMetrics{},
    }

    // Start time for benchmarking duration
    benchmarkStartTime := time.Now()
    benchmarkDuration := time.Duration(config.BenchmarkDurationSeconds) * time.Second

    // Define a context with timeout for benchmarking duration
    ctx, cancel := context.WithTimeout(context.Background(), benchmarkDuration)
    defer cancel()

    // Perform GET and STAT operations first
    var wg sync.WaitGroup
    operations := []OperationType{OperationGet, OperationStat}

    for _, opType := range operations {
        wg.Add(1)
        go func(opType OperationType) {
            defer wg.Done()
            performOperation(ctx, s3Client, opType, metrics[opType])
        }(opType)
    }

    wg.Wait()

    fmt.Println("\nGET and STAT operations completed. Starting DELETE operations...")

    // Reset the context for DELETE operations
    ctx, cancel = context.WithTimeout(context.Background(), benchmarkDuration)
    defer cancel()

    wg.Add(1)
    go func() {
        defer wg.Done()
        performOperation(ctx, s3Client, OperationDelete, metrics[OperationDelete])
    }()

    wg.Wait()

    // Calculate actual benchmarking duration
    actualBenchmarkDuration = time.Since(benchmarkStartTime)

    // Store metrics for final report
    benchmarkMetrics = metrics
}

// performOperation performs a specific S3 operation for the specified duration and collects metrics.
func performOperation(ctx context.Context, s3Client *s3.S3, opType OperationType, metrics *PerformanceMetrics) {
    var mu sync.Mutex
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, config.MaxBenchmarkThreads)
    fileCount := len(uploadedS3Files)

    for {
        select {
        case <-ctx.Done():
            wg.Wait()
            return
        default:
            wg.Add(1)
            semaphore <- struct{}{}
            go func() {
                defer wg.Done()
                s3Key := uploadedS3Files[rand.Intn(fileCount)]
                start := time.Now()
                var err error

                switch opType {
                case OperationGet:
                    _, err = s3Client.GetObject(&s3.GetObjectInput{
                        Bucket: aws.String(config.BucketName),
                        Key:    aws.String(s3Key),
                    })
                case OperationDelete:
                    _, err = s3Client.DeleteObject(&s3.DeleteObjectInput{
                        Bucket: aws.String(config.BucketName),
                        Key:    aws.String(s3Key),
                    })
                case OperationStat:
                    _, err = s3Client.HeadObject(&s3.HeadObjectInput{
                        Bucket: aws.String(config.BucketName),
                        Key:    aws.String(s3Key),
                    })
                }

                duration := time.Since(start)

                mu.Lock()
                metrics.TotalOperations++
                metrics.TotalTime += duration
                if metrics.MinTime == 0 || duration < metrics.MinTime {
                    metrics.MinTime = duration
                }
                if duration > metrics.MaxTime {
                    metrics.MaxTime = duration
                }
                if err != nil {
                    metrics.ErrorCount++
                }
                mu.Unlock()

                <-semaphore
            }()
        }
    }
}

// generateFinalReport generates a summary report of the benchmarking operations.
func generateFinalReport() {
    fmt.Println("\nBenchmarking Report:")
    fmt.Println("====================")

    totalOperations := int64(0)
    totalErrors := int64(0)

    for opType, metrics := range benchmarkMetrics {
        var avgTime time.Duration
        if metrics.TotalOperations > 0 {
            avgTime = time.Duration(int64(metrics.TotalTime) / metrics.TotalOperations)
        }
        totalOperations += metrics.TotalOperations
        totalErrors += metrics.ErrorCount

        fmt.Printf("\nOperation: %s\n", opType)
        fmt.Printf("Total Operations: %d\n", metrics.TotalOperations)
        fmt.Printf("Successes: %d\n", metrics.TotalOperations-metrics.ErrorCount)
        fmt.Printf("Errors: %d\n", metrics.ErrorCount)
        fmt.Printf("Min Time: %v\n", metrics.MinTime)
        fmt.Printf("Max Time: %v\n", metrics.MaxTime)
        fmt.Printf("Avg Time: %v\n", avgTime)
    }

    fmt.Println("\nOverall Benchmark Summary:")
    fmt.Printf("Total Operations: %d\n", totalOperations)
    fmt.Printf("Total Errors: %d\n", totalErrors)
    fmt.Printf("Benchmarking Duration: %v\n", actualBenchmarkDuration)
    fmt.Println("====================")
}
