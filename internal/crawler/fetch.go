package crawler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/time/rate"
)

// Fetcher performs rate-limited HTTP GET requests with optional disk caching.
type Fetcher struct {
	client   *http.Client
	limiter  *rate.Limiter
	cacheDir string
}

// NewFetcher creates a Fetcher.
// delay is the minimum interval between requests (0 = no limit).
// cacheDir is an optional directory for caching responses ("" = disabled).
func NewFetcher(delay time.Duration, cacheDir string) *Fetcher {
	var lim *rate.Limiter
	if delay > 0 {
		lim = rate.NewLimiter(rate.Every(delay), 1)
	} else {
		lim = rate.NewLimiter(rate.Inf, 0)
	}
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter:  lim,
		cacheDir: cacheDir,
	}
}

// Do fetches rawURL and returns the response body bytes.
// It respects the rate limiter and serves from cache when available.
func (f *Fetcher) Do(ctx context.Context, rawURL string) ([]byte, error) {
	if f.cacheDir != "" {
		if data, err := f.readCache(rawURL); err == nil {
			slog.Debug("cache hit", "url", rawURL)
			return data, nil
		}
	}

	if err := f.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	slog.Debug("fetching", "url", rawURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "golm-connector/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, rawURL)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if f.cacheDir != "" {
		if err := f.writeCache(rawURL, data); err != nil {
			slog.Warn("cache write failed", "url", rawURL, "err", err)
		}
	}

	return data, nil
}

func cacheKey(rawURL string) string {
	h := sha256.Sum256([]byte(rawURL))
	return fmt.Sprintf("%x", h)
}

func (f *Fetcher) cachePath(rawURL string) string {
	return filepath.Join(f.cacheDir, cacheKey(rawURL))
}

func (f *Fetcher) readCache(rawURL string) ([]byte, error) {
	return os.ReadFile(f.cachePath(rawURL))
}

func (f *Fetcher) writeCache(rawURL string, data []byte) error {
	if err := os.MkdirAll(f.cacheDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(f.cachePath(rawURL), data, 0o644)
}
