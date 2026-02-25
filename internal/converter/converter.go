package converter

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Run converts HTML files in cfg.InputDir (or cfg.ZipPath) to Markdown files
// in cfg.OutputDir, using a bounded worker pool.
func Run(cfg ConvertConfig) (*ConvertResult, error) {
	inputDir := cfg.InputDir

	// If a ZIP was provided, extract it to a temp dir first.
	if cfg.ZipPath != "" {
		tmp, err := os.MkdirTemp("", "golm-zip-*")
		if err != nil {
			return nil, fmt.Errorf("temp dir for zip: %w", err)
		}
		defer os.RemoveAll(tmp)

		if _, err := ExtractZip(cfg.ZipPath, tmp); err != nil {
			return nil, fmt.Errorf("extract zip: %w", err)
		}
		inputDir = tmp
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir output: %w", err)
	}

	// Collect all .html files.
	var htmlFiles []string
	err := filepath.WalkDir(inputDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".html") {
			htmlFiles = append(htmlFiles, p)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk input dir: %w", err)
	}

	workers := cfg.Workers
	if workers <= 0 {
		workers = 4
	}

	type job struct{ path string }
	type result struct {
		path string
		out  string
		err  error
	}

	jobs := make(chan job, len(htmlFiles))
	results := make(chan result, len(htmlFiles))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				outPath, err := convertFile(j.path, inputDir, cfg.OutputDir, &cfg)
				results <- result{path: j.path, out: outPath, err: err}
			}
		}()
	}

	slog.Info("convert: started", "files", len(htmlFiles), "workers", workers)

	for _, f := range htmlFiles {
		jobs <- job{path: f}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	res := &ConvertResult{Errors: make(map[string]string)}
	done := 0
	for r := range results {
		if r.err != nil {
			slog.Warn("convert error", "file", r.path, "err", r.err)
			res.Errors[r.path] = r.err.Error()
		} else {
			done++
			res.Saved = append(res.Saved, r.out)
			slog.Info("convert: done", "n", done, "total", len(htmlFiles), "file", filepath.Base(r.out))
			slog.Debug("convert: done path", "file", r.path, "out", r.out)
		}
	}

	return res, nil
}

// convertFile reads an HTML file, converts it to Markdown, and writes the output.
func convertFile(htmlPath, inputDir, outputDir string, cfg *ConvertConfig) (string, error) {
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", htmlPath, err)
	}

	node, err := ExtractMainNode(data)
	if err != nil {
		return "", fmt.Errorf("extract %s: %w", htmlPath, err)
	}

	md, err := Transform(node, cfg)
	if err != nil {
		return "", fmt.Errorf("transform %s: %w", htmlPath, err)
	}

	// Derive output path: replace inputDir prefix, change extension to .md.
	rel, err := filepath.Rel(inputDir, htmlPath)
	if err != nil {
		rel = filepath.Base(htmlPath)
	}
	rel = strings.TrimSuffix(rel, filepath.Ext(rel)) + ".md"
	outPath := filepath.Join(outputDir, rel)

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", filepath.Dir(outPath), err)
	}

	if err := os.WriteFile(outPath, []byte(md), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", outPath, err)
	}

	return outPath, nil
}
