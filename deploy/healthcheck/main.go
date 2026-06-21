package main

import (
    "net/http"
    "os"
    "time"
)

func main() {
    url := os.Getenv("HEALTHCHECK_URL")
    if url == "" {
        url = "http://localhost:8382/health"
    }

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        os.Exit(1)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        os.Exit(1)
    }
}
