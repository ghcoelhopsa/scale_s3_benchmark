// benchmark/benchmark.go
package benchmark

import (
    "context"
    "fmt"
    "math/rand"
    "sync"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/s3"

    "scale_s3_benchmark/config"
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

// BenchmarkResult holds the results of the benchmarking.
type BenchmarkResult struct {
    Metrics  map[OperationType]*PerformanceMetrics
    Duration time.Duration
}

// PerformBenchmarkOperations performs GET, STAT, and DELETE operations for benchmarking.
func PerformBenchmarkOperations(cfg *config.Config, s3Client *s3.S3, uploadedS3Files []string, startTime time.Time) BenchmarkResult {
    fmt.Println("\nPerforming benchmarking operations...")

    // Prepare metrics storage
    metrics := map[OperationType]*PerformanceMetrics{
        OperationGet:    &PerformanceMetrics{},
        OperationStat:   &PerformanceMetrics{},
        OperationDelete: &PerformanceMetrics{},
    }

    // Start time for benchmarking duration
    benchmarkStartTime := time.Now()
    benchmarkDuration := time.Duration(cfg.BenchmarkDurationSeconds) * time.Second

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
            performOperation(ctx, cfg, s3Client, opType, metrics[opType], uploadedS3Files, cfg.MaxBenchmarkThreads)
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
        performOperation(ctx, cfg, s3Client, OperationDelete, metrics[OperationDelete], uploadedS3Files, cfg.MaxBenchmarkThreads)
    }()

    wg.Wait()

    // Calculate actual benchmarking duration
    actualBenchmarkDuration := time.Since(benchmarkStartTime)

    // Return benchmark results
    return BenchmarkResult{
        Metrics:  metrics,
        Duration: actualBenchmarkDuration,
    }
}

// performOperation performs a specific S3 operation for the specified duration and collects metrics.
func performOperation(ctx context.Context, cfg *config.Config, s3Client *s3.S3, opType OperationType, metrics *PerformanceMetrics, uploadedS3Files []string, maxBenchmarkThreads int) {
    var mu sync.Mutex
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, maxBenchmarkThreads)

    fileCount := len(uploadedS3Files)
    if fileCount == 0 {
        fmt.Println("No uploaded S3 files available for benchmarking.")
        return
    }

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
                        Bucket: aws.String(cfg.BucketName),
                        Key:    aws.String(s3Key),
                    })
                case OperationDelete:
                    _, err = s3Client.DeleteObject(&s3.DeleteObjectInput{
                        Bucket: aws.String(cfg.BucketName),
                        Key:    aws.String(s3Key),
                    })
                case OperationStat:
                    _, err = s3Client.HeadObject(&s3.HeadObjectInput{
                        Bucket: aws.String(cfg.BucketName),
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

