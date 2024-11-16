// web.go
package main

import (
//    "encoding/json"
    "fmt"
    "html/template"
    "net/http"
    "scale_s3_benchmark/monitor"
    "time"
)

// dashboardHandler handles the main dashboard route and renders the dashboard template.
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("templates/dashboard.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    currentStats := monitor.GetStats()

    data := struct {
        Stats monitor.Stats
    }{
        Stats: currentStats,
    }

    tmpl.Execute(w, data)
}

// statsHandler provides the statistics in JSON format.
func statsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    jsonData, err := monitor.ToJSON()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Write(jsonData)
}

// sseHandler implements Server-Sent Events for real-time updates.
func sseHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-r.Context().Done():
            return
        case <-ticker.C:
            jsonData, err := monitor.ToJSON()
            if err != nil {
                continue
            }
            fmt.Fprintf(w, "data: %s\n\n", jsonData)
            if f, ok := w.(http.Flusher); ok {
                f.Flush()
            }
        }
    }
}

// startWebServer initializes the HTTP server with the necessary routes and starts it.
func startWebServer() {
    // Route for the dashboard.
    http.HandleFunc("/", dashboardHandler)

    // Route for fetching statistics in JSON.
    http.HandleFunc("/stats", statsHandler)

    // Route for Server-Sent Events.
    http.HandleFunc("/events", sseHandler)

    // Serve static files (CSS, JS, etc.).
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // Start the server in a separate goroutine.
    go func() {
        fmt.Println("Web server started on port 8080")
        if err := http.ListenAndServe(":8080", nil); err != nil {
            panic(err)
        }
    }()
}

