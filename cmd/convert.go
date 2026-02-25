package cmd

import (
	"fmt"
	"log/slog"
	"strings"

	"golm-connector/internal/converter"
	"golm-connector/internal/report"

	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert <input-dir>",
	Short: "Convert HTML files to Markdown",
	Args:  cobra.ExactArgs(1),
	RunE:  runConvert,
}

var (
	convertOutput       string
	convertZip          string
	convertWorkers      int
	convertStripTags    string
	convertStripClasses string
	convertRetryReport  string
)

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringVarP(&convertOutput, "output", "o", "md_output", "directory for Markdown output")
	convertCmd.Flags().StringVar(&convertZip, "zip", "", "ZIP archive containing HTML files")
	convertCmd.Flags().IntVar(&convertWorkers, "max-workers", 4, "number of parallel conversion workers")
	convertCmd.Flags().StringVar(&convertStripTags, "strip-tags", "", "comma-separated HTML tags to remove (e.g. nav,footer)")
	convertCmd.Flags().StringVar(&convertStripClasses, "strip-classes", "", "comma-separated CSS classes to remove")
	convertCmd.Flags().StringVar(&convertRetryReport, "retry-from-report", "", "retry failed files from a previous report JSON")
}

func runConvert(cmd *cobra.Command, args []string) error {
	inputDir := args[0]

	cfg := converter.ConvertConfig{
		InputDir:  inputDir,
		OutputDir: convertOutput,
		ZipPath:   convertZip,
		Workers:   convertWorkers,
	}

	if convertStripTags != "" {
		cfg.StripTags = splitAndTrim(convertStripTags)
	}
	if convertStripClasses != "" {
		cfg.StripClasses = splitAndTrim(convertStripClasses)
	}

	if convertRetryReport != "" {
		r, err := report.Load(convertRetryReport)
		if err != nil {
			return fmt.Errorf("load retry report: %w", err)
		}
		_ = report.FailedURLs(r, "convert") // for future filtering
	}

	var rep *report.Report
	var step *report.StepResult
	if reportPath != "" {
		rep = report.NewReport()
		step = report.AddStep(rep, "convert")
	}

	result, err := converter.Run(cfg)

	if step != nil {
		for _, saved := range result.Saved {
			step.URLs = append(step.URLs, report.URLStatus{
				URL:    saved,
				Status: report.StatusOK,
			})
		}
		for f, e := range result.Errors {
			step.URLs = append(step.URLs, report.URLStatus{
				URL:    f,
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

	fmt.Printf("Convert complete: %d saved, %d errors\n", len(result.Saved), len(result.Errors))
	return nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
