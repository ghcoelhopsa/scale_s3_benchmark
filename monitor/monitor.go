// monitor/monitor.go
package monitor

import (
    "encoding/json"
    "sync"
    "time"
)

// Stats defines the structure to store statistics.
type Stats struct {
    TotalUploads int64     `json:"TotalUploads"`
    Successes    int64     `json:"Successes"`
    Failures     int64     `json:"Failures"`
    StartTime    time.Time `json:"StartTime"`
}

var (
    stats     Stats
    statsLock sync.Mutex
)

// InitializeStats initializes the statistics.
func InitializeStats() {
    stats = Stats{
        StartTime: time.Now(),
    }
}

// UpdateStats updates the statistics after an upload.
func UpdateStats(success bool) {
    statsLock.Lock()
    defer statsLock.Unlock()

    stats.TotalUploads++
    if success {
        stats.Successes++
    } else {
        stats.Failures++
    }
}

// GetStats returns a copy of the current statistics.
func GetStats() Stats {
    statsLock.Lock()
    defer statsLock.Unlock()
    return stats
}

// ToJSON returns the statistics in JSON format.
func ToJSON() ([]byte, error) {
    statsLock.Lock()
    defer statsLock.Unlock()
    return json.Marshal(stats)
}

// ResetStats resets the statistics (optional).
func ResetStats() {
    statsLock.Lock()
    defer statsLock.Unlock()
    stats = Stats{
        StartTime: time.Now(),
    }
}

