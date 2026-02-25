package cmd

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"golm-connector/internal/combiner"
	"golm-connector/internal/converter"
	"golm-connector/internal/crawler"
	"golm-connector/internal/report"

	"github.com/spf13/cobra"
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline <url>",
	Short: "Run crawl → convert → combine in sequence",
	Args:  cobra.ExactArgs(1),
	RunE:  runPipeline,
}

var (
	pipelineOutput      string
	pipelineMaxPages    int
	pipelineDelay       time.Duration
	pipelineConcurrency int
	pipelineWorkers     int
	pipelineStripTags   string
	pipelineStripCls    string
	pipelineMaxWords    int
)

func init() {
	rootCmd.AddCommand(pipelineCmd)

	pipelineCmd.Flags().StringVarP(&pipelineOutput, "output", "o", "pipeline_output", "base output directory")
	pipelineCmd.Flags().IntVar(&pipelineMaxPages, "max-pages", 0, "maximum pages to crawl")
	pipelineCmd.Flags().DurationVar(&pipelineDelay, "delay", time.Second, "delay between crawl requests")
	pipelineCmd.Flags().IntVar(&pipelineConcurrency, "max-concurrency", 5, "parallel crawl workers")
	pipelineCmd.Flags().IntVar(&pipelineWorkers, "max-workers", 4, "parallel convert workers")
	pipelineCmd.Flags().StringVar(&pipelineStripTags, "strip-tags", "", "HTML tags to strip during convert")
	pipelineCmd.Flags().StringVar(&pipelineStripCls, "strip-classes", "", "CSS classes to strip during convert")
	pipelineCmd.Flags().IntVar(&pipelineMaxWords, "max-words", 500_000, "max words per combined output file")
}

func runPipeline(cmd *cobra.Command, args []string) error {
	startURL := args[0]
	htmlDir := filepath.Join(pipelineOutput, "html")
	mdDir := filepath.Join(pipelineOutput, "md")
	combinedPath := filepath.Join(pipelineOutput, "combined.md")

	var rep *report.Report
	if reportPath != "" {
		rep = report.NewReport()
	}

	// --- Crawl ---
	slog.Info("pipeline: starting crawl", "url", startURL)
	crawlCfg := crawler.CrawlConfig{
		StartURL:       startURL,
		OutputDir:      htmlDir,
		MaxPages:       pipelineMaxPages,
		Delay:          pipelineDelay,
		MaxConcurrency: pipelineConcurrency,
	}
	var crawlStep *report.StepResult
	if rep != nil {
		crawlStep = report.AddStep(rep, "crawl")
	}
	crawlRes, err := crawler.Run(cmd.Context(), crawlCfg)
	if crawlStep != nil {
		for _, s := range crawlRes.Saved {
			crawlStep.URLs = append(crawlStep.URLs, report.URLStatus{URL: s, Status: report.StatusOK})
		}
		for u, e := range crawlRes.Errors {
			crawlStep.URLs = append(crawlStep.URLs, report.URLStatus{URL: u, Status: report.StatusError, Error: e})
		}
		report.Finish(crawlStep)
	}
	if err != nil {
		return fmt.Errorf("crawl: %w", err)
	}
	slog.Info("pipeline: crawl done", "saved", len(crawlRes.Saved), "errors", len(crawlRes.Errors))

	// --- Convert ---
	slog.Info("pipeline: starting convert")
	convCfg := converter.ConvertConfig{
		InputDir:     htmlDir,
		OutputDir:    mdDir,
		Workers:      pipelineWorkers,
		StripTags:    splitAndTrim(pipelineStripTags),
		StripClasses: splitAndTrim(pipelineStripCls),
	}
	var convStep *report.StepResult
	if rep != nil {
		convStep = report.AddStep(rep, "convert")
	}
	convRes, err := converter.Run(convCfg)
	if convStep != nil {
		for _, s := range convRes.Saved {
			convStep.URLs = append(convStep.URLs, report.URLStatus{URL: s, Status: report.StatusOK})
		}
		for f, e := range convRes.Errors {
			convStep.URLs = append(convStep.URLs, report.URLStatus{URL: f, Status: report.StatusError, Error: e})
		}
		report.Finish(convStep)
	}
	if err != nil {
		return fmt.Errorf("convert: %w", err)
	}
	slog.Info("pipeline: convert done", "saved", len(convRes.Saved), "errors", len(convRes.Errors))

	// --- Combine ---
	slog.Info("pipeline: starting combine")
	combRes, err := combiner.Run(combiner.CombineConfig{
		InputDir:   mdDir,
		OutputPath: combinedPath,
		MaxWords:   pipelineMaxWords,
	})
	if err != nil {
		return fmt.Errorf("combine: %w", err)
	}
	slog.Info("pipeline: combine done", "outputs", len(combRes.OutputFiles))

	if rep != nil && reportPath != "" {
		if writeErr := report.Write(rep, reportPath); writeErr != nil {
			slog.Warn("failed to write report", "err", writeErr)
		}
	}

	fmt.Printf("Pipeline complete:\n")
	fmt.Printf("  crawl:   %d pages saved\n", len(crawlRes.Saved))
	fmt.Printf("  convert: %d files converted\n", len(convRes.Saved))
	fmt.Printf("  combine: %d output file(s), %d words total\n", len(combRes.OutputFiles), combRes.TotalWords)
	for _, f := range combRes.OutputFiles {
		fmt.Printf("    %s\n", f)
	}
	return nil
}
