package crawler

import "time"

// CrawlConfig holds all parameters for a crawl run.
type CrawlConfig struct {
	// StartURL is the seed URL (scope is derived from its host+path prefix).
	StartURL string
	// OutputDir is the directory where downloaded HTML files are saved.
	OutputDir string
	// MaxPages is the maximum number of pages to crawl (0 = unlimited).
	MaxPages int
	// Delay is the minimum time between requests (per-domain rate limiter).
	Delay time.Duration
	// MaxConcurrency is the number of parallel HTTP workers.
	MaxConcurrency int
	// CacheDir is an optional disk-cache directory for HTTP responses.
	CacheDir string
	// RetryURLs is an optional list of URLs to retry (from a previous report).
	RetryURLs []string
}

// CrawlResult summarises the outcome of a crawl run.
type CrawlResult struct {
	// Saved lists the output file paths that were written.
	Saved []string
	// Errors maps URL → error message for failed fetches.
	Errors map[string]string
}
