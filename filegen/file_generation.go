// filegen/file_generation.go
package filegen

import (
    "fmt"
    "math/rand"
    "os"
    "path/filepath"
    "time"

    "scale_s3_benchmark/config"
)

// PrepareBaseDirectory prepares the base directory by creating it if it doesn't exist.
func PrepareBaseDirectory(baseDir string) error {
    return os.MkdirAll(baseDir, os.ModePerm)
}

// GenerateAllBaseFiles generates a specified number of base files with random content.
// It skips generating files that already exist.
func GenerateAllBaseFiles(cfg *config.Config) {
    startTime := time.Now()
    for i := 0; i < cfg.BaseFileCount; i++ {
        filename := filepath.Join(cfg.BaseDirectory, fmt.Sprintf("file_base_%d.txt", i))

        // Check if the file already exists.
        if _, err := os.Stat(filename); os.IsNotExist(err) {
            if err := GenerateTextFile(filename, cfg.MinSize, cfg.MaxSize); err != nil {
                fmt.Printf("Error generating base file %s: %v\n", filename, err)
            }
        }

        elapsedTime := time.Since(startTime).Seconds()
        rate := float64(i+1) / elapsedTime
        progress := float64(i+1) / float64(cfg.BaseFileCount) * 100
        fmt.Printf("Generating base files: %d/%d (%.2f%%) | Rate: %.2f files/sec\r", i+1, cfg.BaseFileCount, progress, rate)
    }
    fmt.Printf("100%% completed - %d base files generated.\n", cfg.BaseFileCount)
}

// GenerateTextFile creates a text file with random alphabetical content of a specified size.
func GenerateTextFile(filename string, minSize, maxSize int) error {
    size := rand.Intn(maxSize-minSize+1) + minSize
    content := make([]byte, size)
    for i := range content {
        content[i] = byte('a' + rand.Intn(26))
    }
    return os.WriteFile(filename, content, 0644)
}

