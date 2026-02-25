package converter

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractZip extracts all .html files from zipPath into destDir.
// Returns the list of extracted file paths.
func ExtractZip(zipPath, destDir string) ([]string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("open zip %s: %w", zipPath, err)
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", destDir, err)
	}

	var extracted []string
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(f.Name), ".html") {
			continue
		}

		outPath := filepath.Join(destDir, filepath.FromSlash(f.Name))
		// Prevent zip-slip: ensure the output path is within destDir.
		if !strings.HasPrefix(outPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return nil, fmt.Errorf("mkdir for %s: %w", outPath, err)
		}

		if err := extractFile(f, outPath); err != nil {
			return nil, err
		}
		extracted = append(extracted, outPath)
	}
	return extracted, nil
}

func extractFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open zip entry %s: %w", f.Name, err)
	}
	defer rc.Close()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create %s: %w", dest, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, rc); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	return nil
}
