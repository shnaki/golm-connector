package combiner

// CombineConfig holds parameters for combining Markdown files.
type CombineConfig struct {
	// InputDir is the directory containing .md files to combine.
	InputDir string
	// OutputPath is the path for the output file (e.g., combined.md).
	// When auto-split is triggered, numbered suffixes are added before the extension.
	OutputPath string
	// MaxWords is the word limit per output file (0 = no limit).
	// Default is 500_000 when not set (callers should set this explicitly).
	MaxWords int
}

// CombineResult summarises the output of a combine run.
type CombineResult struct {
	// OutputFiles lists the paths of the files that were written.
	OutputFiles []string
	// TotalWords is the total word count across all input files.
	TotalWords int
	// FileCount is the number of input .md files processed.
	FileCount int
}
