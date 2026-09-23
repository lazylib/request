package request

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
)

// Download performs opts and writes the response body to the file path
// specified by `to`. It returns an error if the request cannot be built or sent,
// if the server responds with a non-2xx status, or if writing the file fails.
func Download(opts Options, to string) error {
    body, contentType, err := buildBody(opts.Body)
    if err != nil {
        return fmt.Errorf("request: build body: %w", err)
    }
    req, err := http.NewRequest(opts.Method, opts.Url, body)
    if err != nil {
        return fmt.Errorf("request: build request: %w", err)
    }
    if contentType != "" {
        req.Header.Set("Content-Type", contentType)
    }
    for k, v := range opts.Headers {
        req.Header.Set(k, v)
    }
    if opts.Auth != nil {
        opts.Auth.apply(req)
    }
    resp, err := (&http.Client{}).Do(req)
    if err != nil {
        return fmt.Errorf("request: send: %w", err)
    }
    defer func() { _ = resp.Body.Close() }()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("request: server responded with status %d", resp.StatusCode)
    }

    // Ensure destination directory exists.
    if dir := filepath.Dir(to); dir != "" && dir != "." {
        if err := os.MkdirAll(dir, 0o755); err != nil {
            return fmt.Errorf("request: create dirs: %w", err)
        }
    }
    f, err := os.Create(to)
    if err != nil {
        return fmt.Errorf("request: create file: %w", err)
    }
    defer func() { _ = f.Close() }()
    if _, err := io.Copy(f, resp.Body); err != nil {
        return fmt.Errorf("request: write file: %w", err)
    }
    return nil
}

// DownloadX performs the same operation as Download but panics on error.
func DownloadX(opts Options, to string) {
    if err := Download(opts, to); err != nil {
        panic(err)
    }
}
