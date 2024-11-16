// filegen/replication.go
package filegen

import (
    "context"
    "fmt"
    "io" // Added import for io
    "os"
    "os/exec"
    "path/filepath"
    "sync"
    "time"

    "scale_s3_benchmark/config"
)

// ReplicateFilesWithReflinkInParallel replicates files using reflink in parallel.
// It returns the list of replicated file paths and any error encountered.
func ReplicateFilesWithReflinkInParallel(cfg *config.Config) ([]string, error) {
    fmt.Println("Starting file replication with reflink in parallel.")
    startTime := time.Now()

    var replicatedFiles []string
    var mu sync.Mutex
    var replicationWG sync.WaitGroup

    jobs := make(chan int, 1000) // Adjusted buffer size

    folderPath := cfg.BaseDirectory
    os.MkdirAll(folderPath, os.ModePerm)

    // Error channel to collect errors
    errorChan := make(chan error, 1000)

    replicationWG.Add(cfg.MaxConcurrentReplicas)
    for w := 0; w < cfg.MaxConcurrentReplicas; w++ {
        go func() {
            defer replicationWG.Done()
            for currentCount := range jobs {
                baseFileIndex := currentCount % cfg.BaseFileCount
                src := filepath.Join(cfg.BaseDirectory, fmt.Sprintf("file_base_%d.txt", baseFileIndex))
                dst := filepath.Join(folderPath, fmt.Sprintf("file_%d.txt", currentCount))

                if err := CopyFileReflink(src, dst); err != nil {
                    fmt.Printf("\nError replicating file %s to %s: %v\n", src, dst, err)
                    errorChan <- err
                    continue
                }

                mu.Lock()
                replicatedFiles = append(replicatedFiles, dst)
                mu.Unlock()

                elapsedTime := time.Since(startTime).Seconds()
                rate := float64(len(replicatedFiles)) / elapsedTime
                progress := float64(len(replicatedFiles)) / float64(cfg.MaxLocalFiles) * 100
                fmt.Printf("Replicating files: %d/%d (%.2f%%) | Rate: %.2f files/sec\r", len(replicatedFiles), cfg.MaxLocalFiles, progress, rate)
            }
        }()
    }

    // Send jobs to workers
    go func() {
        defer close(jobs)
        for currentCount := 0; currentCount < cfg.MaxLocalFiles; currentCount++ {
            jobs <- currentCount
        }
    }()

    replicationWG.Wait()
    close(errorChan)

    // Check for replication errors
    errorCount := len(errorChan)
    if errorCount > 0 {
        fmt.Printf("\n%d errors occurred during file replication.\n", errorCount)
    } else {
        fmt.Println("\nFile replication completed successfully.")
    }

    return replicatedFiles, nil
}

// CopyFileReflink copies a file using reflink, falling back to a regular copy if reflink fails.
func CopyFileReflink(src, dst string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctx, "cp", "--reflink=always", src, dst)

    _, err := cmd.CombinedOutput()
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("cp command timed out")
    }

    if err != nil {
        return FallbackCopy(src, dst) // Fallback to a regular copy if reflink fails.
    }
    return nil
}

// FallbackCopy performs a traditional file copy if reflink is not supported.
func FallbackCopy(src, dst string) error {
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

