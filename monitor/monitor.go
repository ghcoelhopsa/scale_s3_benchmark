// monitor/monitor.go
package monitor

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
    "os"
    "sync"
    "time"
)

// Stats define a estrutura para armazenar estatísticas.
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

// InitializeStats inicializa as estatísticas.
func InitializeStats() {
    stats = Stats{
        StartTime: time.Now(),
    }
}

// UpdateStats atualiza as estatísticas após um upload.
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

// GetStats retorna uma cópia das estatísticas atuais.
func GetStats() Stats {
    statsLock.Lock()
    defer statsLock.Unlock()
    return stats
}

// ToJSON retorna as estatísticas em formato JSON.
func ToJSON() ([]byte, error) {
    statsLock.Lock()
    defer statsLock.Unlock()
    return json.Marshal(stats)
}

// ResetStats reseta as estatísticas (opcional).
func ResetStats() {
    statsLock.Lock()
    defer statsLock.Unlock()
    stats = Stats{
        StartTime: time.Now(),
    }
}

// StartPeriodicReporting inicia uma goroutine que grava estatísticas em um arquivo CSV a cada intervalo definido.
func StartPeriodicReporting(filePath string, interval time.Duration) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()

        // Abrir o arquivo em modo de acréscimo (append), criar se não existir
        file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
        if err != nil {
            fmt.Printf("Erro ao abrir o arquivo de relatório: %v\n", err)
            return
        }
        defer file.Close()

        writer := csv.NewWriter(file)
        defer writer.Flush()

        // Escrever cabeçalho CSV se o arquivo estiver vazio
        fileInfo, err := file.Stat()
        if err != nil {
            fmt.Printf("Erro ao obter informações do arquivo: %v\n", err)
            return
        }
        if fileInfo.Size() == 0 {
            header := []string{"Timestamp", "TotalUploads", "Successes", "Failures"}
            if err := writer.Write(header); err != nil {
                fmt.Printf("Erro ao escrever cabeçalho no arquivo de relatório: %v\n", err)
                return
            }
            writer.Flush()
            if err := writer.Error(); err != nil {
                fmt.Printf("Erro ao flush do cabeçalho no arquivo de relatório: %v\n", err)
                return
            }
        }

        for {
            select {
            case <-ticker.C:
                statsLock.Lock()
                currentStats := stats
                statsLock.Unlock()

                record := []string{
                    time.Now().Format(time.RFC3339), // Timestamp atual
                    fmt.Sprintf("%d", currentStats.TotalUploads),
                    fmt.Sprintf("%d", currentStats.Successes),
                    fmt.Sprintf("%d", currentStats.Failures),
                }

                if err := writer.Write(record); err != nil {
                    fmt.Printf("Erro ao escrever no arquivo de relatório: %v\n", err)
                    return
                }
                writer.Flush()
                if err := writer.Error(); err != nil {
                    fmt.Printf("Erro ao flush do arquivo de relatório: %v\n", err)
                    return
                }
            }
        }
    }()
}
