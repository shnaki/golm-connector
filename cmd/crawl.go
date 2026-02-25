package cmd

import (
	"fmt"
	"log/slog"
	"time"

	"golm-connector/internal/crawler"
	"golm-connector/internal/report"

	"github.com/spf13/cobra"
)

var crawlCmd = &cobra.Command{
	Use:   "crawl <url>",
	Short: "BFS-crawl a website and save HTML pages",
	Args:  cobra.ExactArgs(1),
	RunE:  runCrawl,
}

var (
	crawlOutput      string
	crawlMaxPages    int
	crawlDelay       time.Duration
	crawlConcurrency int
	crawlCacheDir    string
	crawlRetryReport string
)

func init() {
	rootCmd.AddCommand(crawlCmd)

	crawlCmd.Flags().StringVarP(&crawlOutput, "output", "o", "html_output", "directory for saved HTML files")
	crawlCmd.Flags().IntVar(&crawlMaxPages, "max-pages", 0, "maximum number of pages to crawl (0 = unlimited)")
	crawlCmd.Flags().DurationVar(&crawlDelay, "delay", time.Second, "delay between requests (e.g. 1s, 500ms)")
	crawlCmd.Flags().IntVar(&crawlConcurrency, "max-concurrency", 5, "number of parallel HTTP workers")
	crawlCmd.Flags().StringVar(&crawlCacheDir, "cache-dir", "", "disk cache directory for HTTP responses")
	crawlCmd.Flags().StringVar(&crawlRetryReport, "retry-from-report", "", "retry failed URLs from a previous report JSON")
}

func runCrawl(cmd *cobra.Command, args []string) error {
	startURL := args[0]

	cfg := crawler.CrawlConfig{
		StartURL:       startURL,
		OutputDir:      crawlOutput,
		MaxPages:       crawlMaxPages,
		Delay:          crawlDelay,
		MaxConcurrency: crawlConcurrency,
		CacheDir:       crawlCacheDir,
	}

	if crawlRetryReport != "" {
		r, err := report.Load(crawlRetryReport)
		if err != nil {
			return fmt.Errorf("load retry report: %w", err)
		}
		cfg.RetryURLs = report.FailedURLs(r, "crawl")
		slog.Info("retrying failed URLs", "count", len(cfg.RetryURLs))
	}

	var rep *report.Report
	var step *report.StepResult
	if reportPath != "" {
		rep = report.NewReport()
		step = report.AddStep(rep, "crawl")
	}

	result, err := crawler.Run(cmd.Context(), cfg)

	if step != nil {
		for _, saved := range result.Saved {
			step.URLs = append(step.URLs, report.URLStatus{
				URL:    saved,
				Status: report.StatusOK,
			})
		}
		for u, e := range result.Errors {
			step.URLs = append(step.URLs, report.URLStatus{
				URL:    u,
				Status: report.StatusError,
				Error:  e,
			})
		}
		report.Finish(step)
		if writeErr := report.Write(rep, reportPath); writeErr != nil {
			slog.Warn("failed to write report", "err", writeErr)
		}
	}

	if err != nil {
		return err
	}

	fmt.Printf("Crawl complete: %d saved, %d errors\n", len(result.Saved), len(result.Errors))
	return nil
}
