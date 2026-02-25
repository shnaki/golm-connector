package converter

// ConvertConfig holds all parameters for a convert run.
type ConvertConfig struct {
	// InputDir is the directory containing HTML files to convert.
	InputDir string
	// OutputDir is the directory where Markdown files are written.
	OutputDir string
	// ZipPath is an optional ZIP archive to read HTML from (overrides InputDir).
	ZipPath string
	// Workers is the number of parallel conversion goroutines.
	Workers int
	// StripTags is a list of HTML tag names whose elements should be removed.
	StripTags []string
	// StripClasses is a list of CSS class names whose elements should be removed.
	StripClasses []string
}

// ConvertResult summarises the outcome of a convert run.
type ConvertResult struct {
	// Saved lists the output .md file paths that were written.
	Saved []string
	// Errors maps input file path → error message.
	Errors map[string]string
}
