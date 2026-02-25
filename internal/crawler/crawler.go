package crawler

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// Run executes a BFS crawl according to cfg.
// It returns a CrawlResult summarising saved files and errors.
func Run(ctx context.Context, cfg CrawlConfig) (*CrawlResult, error) {
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir output: %w", err)
	}

	fetcher := NewFetcher(cfg.Delay, cfg.CacheDir)
	result := &CrawlResult{Errors: make(map[string]string)}

	seed := Normalize(cfg.StartURL)
	if seed == "" {
		return nil, fmt.Errorf("invalid start URL: %s", cfg.StartURL)
	}

	// Seed the initial queue.
	initialURLs := []string{seed}
	if len(cfg.RetryURLs) > 0 {
		initialURLs = cfg.RetryURLs
	}

	type fetchJob struct {
		url string
	}
	type fetchRes struct {
		url      string
		finalURL string
		data     []byte
		err      error
	}

	concurrency := cfg.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 5
	}

	jobs := make(chan fetchJob, concurrency*2)
	results := make(chan fetchRes, concurrency*2)

	// Start worker pool.
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				data, finalURL, err := fetcher.Do(ctx, job.url)
				results <- fetchRes{url: job.url, finalURL: finalURL, data: data, err: err}
			}
		}()
	}

	// Close results when all workers are done.
	go func() {
		wg.Wait()
		close(results)
	}()

	// BFS dispatcher (runs in main goroutine).
	visited := make(map[string]bool)
	queue := make([]string, 0, len(initialURLs))
	for _, u := range initialURLs {
		n := Normalize(u)
		if n != "" && !visited[n] {
			visited[n] = true
			queue = append(queue, n)
		}
	}

	pending := 0

	dispatch := func() {
		for len(queue) > 0 && pending < concurrency*2 {
			if cfg.MaxPages > 0 && len(result.Saved)+len(result.Errors)+pending >= cfg.MaxPages {
				break
			}
			url := queue[0]
			queue = queue[1:]
			jobs <- fetchJob{url: url}
			pending++
		}
	}

	dispatch()

	for pending > 0 {
		res := <-results
		pending--

		if res.err != nil {
			slog.Warn("fetch error", "url", res.url, "err", res.err)
			result.Errors[res.url] = res.err.Error()
		} else {
			// Parse HTML and save.
			outPath, err := saveHTML(cfg.OutputDir, res.url, res.data)
			if err != nil {
				slog.Warn("save error", "url", res.url, "err", err)
				result.Errors[res.url] = err.Error()
			} else {
				slog.Info("saved", "url", res.url, "path", outPath)
				result.Saved = append(result.Saved, outPath)
			}

			// Extract links and enqueue new ones (skip if retry mode).
			if len(cfg.RetryURLs) == 0 {
				doc, err := html.Parse(bytes.NewReader(res.data))
				if err == nil {
					base := res.finalURL
					if base == "" {
						base = res.url
					}
					for _, link := range ExtractLinks(base, doc) {
						if !visited[link] && InScope(seed, link) && isHTTPURL(link) {
							visited[link] = true
							queue = append(queue, link)
						}
					}
				}
			}
		}

		dispatch()
	}

	close(jobs)

	return result, nil
}

// saveHTML writes data to OutputDir using a path derived from the URL.
func saveHTML(outputDir, rawURL string, data []byte) (string, error) {
	rel := URLToFilename(rawURL)
	if rel == "" {
		return "", fmt.Errorf("cannot derive filename from URL: %s", rawURL)
	}
	outPath := filepath.Join(outputDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return "", err
	}
	return outPath, nil
}

func isHTTPURL(u string) bool {
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
}
