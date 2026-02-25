package combiner

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const defaultMaxWords = 500_000

// Run collects all .md files from cfg.InputDir (alphabetically), concatenates
// them, and writes the result to cfg.OutputPath.  When the accumulated word
// count exceeds cfg.MaxWords, a new output file is started automatically.
func Run(cfg CombineConfig) (*CombineResult, error) {
	maxWords := cfg.MaxWords
	if maxWords <= 0 {
		maxWords = defaultMaxWords
	}

	// Collect .md files in sorted order (WalkDir is lexicographic).
	var mdFiles []string
	err := filepath.WalkDir(cfg.InputDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			mdFiles = append(mdFiles, p)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk input dir: %w", err)
	}

	if len(mdFiles) == 0 {
		return &CombineResult{}, nil
	}

	// Ensure output directory exists.
	outDir := filepath.Dir(cfg.OutputPath)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir output: %w", err)
	}

	res := &CombineResult{FileCount: len(mdFiles)}

	partIndex := 1
	wordCount := 0
	var buf strings.Builder

	flush := func() error {
		if buf.Len() == 0 {
			return nil
		}
		path := partPath(cfg.OutputPath, partIndex, len(res.OutputFiles) == 0 && wordCount <= maxWords)
		// If we have more than one part, always use numbered names.
		// Recalculate: use numbered only when we have split (partIndex > 1 or will split).
		if err := os.WriteFile(path, []byte(buf.String()), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		slog.Info("wrote combined file", "path", path, "words", wordCount)
		res.OutputFiles = append(res.OutputFiles, path)
		partIndex++
		buf.Reset()
		wordCount = 0
		return nil
	}

	// First pass: read all files and determine if we need splitting.
	type entry struct {
		path    string
		content string
		words   int
	}

	entries := make([]entry, 0, len(mdFiles))
	totalWords := 0
	for _, f := range mdFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			slog.Warn("read error, skipping", "file", f, "err", err)
			continue
		}
		content := string(data)
		w := len(strings.Fields(content))
		entries = append(entries, entry{path: f, content: content, words: w})
		totalWords += w
	}
	res.TotalWords = totalWords

	needSplit := totalWords > maxWords

	for _, e := range entries {
		// If adding this file would exceed the limit, flush current buffer.
		if needSplit && wordCount+e.words > maxWords && buf.Len() > 0 {
			if err := flush(); err != nil {
				return nil, err
			}
		}

		if buf.Len() > 0 {
			buf.WriteString("\n\n---\n\n")
		}
		buf.WriteString(e.content)
		wordCount += e.words
	}

	// Flush remaining content.
	if buf.Len() > 0 {
		var path string
		if needSplit {
			path = numberedPath(cfg.OutputPath, partIndex)
		} else {
			path = cfg.OutputPath
		}
		if err := os.WriteFile(path, []byte(buf.String()), 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", path, err)
		}
		slog.Info("wrote combined file", "path", path)
		res.OutputFiles = append(res.OutputFiles, path)
	}

	// Rename first parts if we split (replace the flush-written paths).
	// Since flush() uses numbered paths when needSplit, nothing to rename.

	return res, nil
}

// numberedPath inserts a zero-padded part number before the file extension.
// e.g. combined.md + 2 → combined-002.md
func numberedPath(base string, n int) string {
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	return fmt.Sprintf("%s-%03d%s", stem, n, ext)
}

// partPath returns either the plain output path (when no split is needed) or
// a numbered path.  The flush helper uses this to decide naming on the fly.
func partPath(base string, n int, noSplit bool) string {
	if noSplit {
		return base
	}
	return numberedPath(base, n)
}
