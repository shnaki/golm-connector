package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose    bool
	reportPath string
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "golm-connector",
	Short: "BFS crawler + HTML→Markdown converter + Markdown combiner",
	Long: `golm-connector crawls websites, converts HTML pages to Markdown,
and combines them into a single document suitable for NotebookLM.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level := slog.LevelWarn
		if verbose {
			level = slog.LevelDebug
		}
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
		return nil
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose (debug) logging")
	rootCmd.PersistentFlags().StringVar(&reportPath, "report", "", "write JSON report to this path")
}
