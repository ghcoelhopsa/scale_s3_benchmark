// s3upload/upload.go
package s3upload

import (
    "fmt"
    "math"
    "os"
    "path/filepath"
    "sync"
    "sync/atomic"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/s3"

    "scale_s3_benchmark/config"
    "scale_s3_benchmark/monitor"
)

// Uploader handles uploading files to S3 with retry logic.
type Uploader struct {
    Config          *config.Config
    S3Clients       []*s3.S3
    SuccessCount    int64
    ClientIndex     uint64
    UploadedS3Files []string
    Mutex           sync.Mutex
    StartTime       time.Time
}

// NewUploader creates a new Uploader instance.
func NewUploader(cfg *config.Config, s3Clients []*s3.S3, startTime time.Time) *Uploader {
    return &Uploader{
        Config:          cfg,
        S3Clients:       s3Clients,
        UploadedS3Files: make([]string, 0),
        StartTime:       startTime,
    }
}

// UploadFiles concurrently uploads a list of files to S3 with a specified concurrency.
func (u *Uploader) UploadFiles(subfolderName string, filePaths []string) {
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, u.Config.MaxConcurrentUploads)

    for _, filePath := range filePaths {
        wg.Add(1)
        go func(fp string) {
            defer wg.Done()
            semaphore <- struct{}{}
            err := u.UploadFileWithRetry(fp, subfolderName)
            if err != nil {
                fmt.Printf("Error uploading file %s: %v\n", fp, err)
            }
            <-semaphore
        }(filePath)
    }

    wg.Wait()
}

// UploadFileWithRetry attempts to upload a file to S3, retrying on failure.
func (u *Uploader) UploadFileWithRetry(filePath string, subfolderName string) error {
    // S3 key structure: s3Folder/subfolderName/fileName
    fileName := filepath.Base(filePath)
    s3Key := filepath.Join(u.Config.S3Folder, subfolderName, fileName)

    for attempt := 1; attempt <= u.Config.MaxRetries; attempt++ {
        if err := u.uploadFile(filePath, s3Key); err == nil {
            atomic.AddInt64(&u.SuccessCount, 1)

            // Update global statistics
            monitor.UpdateStats(true)

            elapsedTime := time.Since(u.StartTime).Seconds()
            rate := float64(u.SuccessCount) / elapsedTime
            progress := float64(u.SuccessCount) / float64(u.Config.TotalFiles) * 100
            fmt.Printf("Uploading to S3: %d/%d (%.2f%%) | Rate: %.2f files/sec\r", u.SuccessCount, u.Config.TotalFiles, progress, rate)

            // Store uploaded S3 key
            u.Mutex.Lock()
            u.UploadedS3Files = append(u.UploadedS3Files, s3Key)
            u.Mutex.Unlock()

            return nil
        } else if attempt < u.Config.MaxRetries {
            backoffDuration := time.Duration(math.Pow(2, float64(attempt))) * time.Second
            time.Sleep(backoffDuration) // Exponential backoff before retrying.
        } else {
            fmt.Printf("\nFailed to upload %s after %d attempts\n", filePath, u.Config.MaxRetries)
            // Update global statistics
            monitor.UpdateStats(false)
            return fmt.Errorf("failed to upload %s after %d attempts", filePath, u.Config.MaxRetries)
        }
    }

    return fmt.Errorf("failed to upload %s after %d attempts", filePath, u.Config.MaxRetries)
}

// uploadFile uploads a single file to S3 using a selected S3 client.
func (u *Uploader) uploadFile(filePath, s3Key string) error {
    clientIndex := atomic.AddUint64(&u.ClientIndex, 1)
    s3Client := u.S3Clients[clientIndex%uint64(len(u.S3Clients))]

    fileData, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("error opening file %s: %w", filePath, err)
    }
    defer fileData.Close()

    _, err = s3Client.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(u.Config.BucketName),
        Key:    aws.String(s3Key),
        Body:   fileData,
    })
    return err
}

