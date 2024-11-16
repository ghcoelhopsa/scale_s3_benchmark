// main.go
package main

import (
    "fmt"
    "math/rand"
    "os"
    "sync"
    "syscall"
    "time"

    "scale_s3_benchmark/benchmark"
    "scale_s3_benchmark/config"
    "scale_s3_benchmark/filegen"
    "scale_s3_benchmark/monitor"
    "scale_s3_benchmark/s3upload"
)

func main() {
    // Load configuration from config.json.
    cfg, err := config.LoadConfig("config.json")
    if err != nil {
        fmt.Printf("Error loading configuration: %v\n", err)
        os.Exit(1)
    }

    // Initialize statistics.
    monitor.InitializeStats()

    // Start the web server for the dashboard.
    //startWebServer()

    // Seed the random number generator.
    rand.Seed(time.Now().UnixNano())

    // Increase the file descriptor limit to handle many files.
    if err := increaseFileDescriptorLimit(); err != nil {
        fmt.Printf("Error adjusting file descriptor limits: %v\n", err)
        return
    }

    // Prepare the base directory for generating files.
    if err := filegen.PrepareBaseDirectory(cfg.BaseDirectory); err != nil {
        fmt.Printf("Error preparing base directory: %v\n", err)
        return
    }

    // Generate the base set of files (if they don't already exist).
    filegen.GenerateAllBaseFiles(cfg)

    // Replicate files locally up to the maximum local files limit only once.
    localFiles, err := filegen.ReplicateFilesWithReflinkInParallel(cfg)
    if err != nil {
        fmt.Printf("Error replicating files with reflink: %v\n", err)
        return
    }

    fmt.Printf("Replication of %d files completed.\n", len(localFiles))

    // Initialize S3 clients.
    s3Clients, err := s3upload.InitializeS3Clients(cfg)
    if err != nil {
        fmt.Printf("Error initializing S3 clients: %v\n", err)
        return
    }

    // Create an uploader instance.
    uploader := s3upload.NewUploader(cfg, s3Clients, time.Now())

    totalFilesUploaded := int64(0)

    // Channel to control the number of subfolders being processed concurrently.
    subfolderSemaphore := make(chan struct{}, cfg.MaxConcurrentSubfolders)
    var wg sync.WaitGroup

    for folderIndex := 0; totalFilesUploaded < int64(cfg.TotalFiles); folderIndex++ {
        filesToProcess := int64(cfg.MaxFilesPerFolder)
        if int64(cfg.TotalFiles)-totalFilesUploaded < filesToProcess {
            filesToProcess = int64(cfg.TotalFiles) - totalFilesUploaded
        }

        wg.Add(1)

        go func(folderIdx int, filesCount int64) {
            defer wg.Done()

            // Acquire a slot in the semaphore.
            subfolderSemaphore <- struct{}{}

            // Process the subfolder.
            processSubfolder(folderIdx, filesCount, localFiles, uploader, cfg)

            // Release the slot in the semaphore.
            <-subfolderSemaphore

            // Pause between folder uploads as per configuration.
            fmt.Printf("Pausing for %d seconds before the next upload...\n", cfg.PauseDurationSeconds)
            time.Sleep(time.Duration(cfg.PauseDurationSeconds) * time.Second)

        }(folderIndex, filesToProcess)

        totalFilesUploaded += filesToProcess
    }

    wg.Wait()

    fmt.Println("\nAll uploads completed.")

    // Clean up local files to free up space.
    cleanupLocalFiles(localFiles)

    // Perform benchmarking operations.
    benchmarkResult := benchmark.PerformBenchmarkOperations(cfg, s3Clients[0], uploader.UploadedS3Files, monitor.GetStats().StartTime)

    // Generate the final report.
    benchmark.GenerateFinalReport(benchmarkResult)
}

// processSubfolder handles the creation and upload of files to a single subfolder.
func processSubfolder(folderIndex int, filesToProcess int64, localFiles []string, uploader *s3upload.Uploader, cfg *config.Config) {
    fmt.Printf("\nProcessing subfolder %d...\n", folderIndex)

    // Include the folderIndex in the subfolderName to ensure uniqueness.
    dateTimeStr := time.Now().Format("02012006150405") // DDMMYYYYHHMMSS
    folderFilesCount := fmt.Sprintf("%d", filesToProcess)
    subfolderName := fmt.Sprintf("FOLDER_%s_%s_%d", dateTimeStr, folderFilesCount, folderIndex)

    // Prepare the list of files to upload.
    filePaths := make([]string, filesToProcess)
    for i := int64(0); i < filesToProcess; i++ {
        filePaths[i] = localFiles[i%int64(len(localFiles))]
    }

    // Start uploading files to S3 in parallel.
    uploader.UploadFiles(subfolderName, filePaths)

    fmt.Printf("\nUpload completed for subfolder index %d.\n", folderIndex)
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

// cleanupLocalFiles removes the replicated local files to free up space.
func cleanupLocalFiles(files []string) {
    for _, filePath := range files {
        os.Remove(filePath)
    }
}

