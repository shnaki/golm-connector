package cmd

import (
	"fmt"

	"golm-connector/internal/combiner"

	"github.com/spf13/cobra"
)

var combineCmd = &cobra.Command{
	Use:   "combine <input-dir>",
	Short: "Combine Markdown files into a single document",
	Args:  cobra.ExactArgs(1),
	RunE:  runCombine,
}

var (
	combineOutput   string
	combineMaxWords int
)

func init() {
	rootCmd.AddCommand(combineCmd)

	combineCmd.Flags().StringVarP(&combineOutput, "output", "o", "combined.md", "output file path")
	combineCmd.Flags().IntVar(&combineMaxWords, "max-words", 500_000, "max words per output file (0 = unlimited)")
}

func runCombine(cmd *cobra.Command, args []string) error {
	inputDir := args[0]

	cfg := combiner.CombineConfig{
		InputDir:   inputDir,
		OutputPath: combineOutput,
		MaxWords:   combineMaxWords,
	}

	result, err := combiner.Run(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("Combine complete: %d files → %d output(s), %d total words\n",
		result.FileCount, len(result.OutputFiles), result.TotalWords)
	for _, f := range result.OutputFiles {
		fmt.Printf("  %s\n", f)
	}
	return nil
}
