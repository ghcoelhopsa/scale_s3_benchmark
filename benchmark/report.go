// benchmark/report.go
package benchmark

import (
    "fmt"
    "time" // Added import for time
)

// GenerateFinalReport generates a summary report of the benchmarking operations.
func GenerateFinalReport(result BenchmarkResult) {
    fmt.Println("\nBenchmarking Report:")
    fmt.Println("====================")

    totalOperations := int64(0)
    totalErrors := int64(0)

    for opType, metrics := range result.Metrics {
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
    fmt.Printf("Benchmarking Duration: %v\n", result.Duration)
    fmt.Println("====================")
}

