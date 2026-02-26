package combiner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCombineBasic(t *testing.T) {
	dir := t.TempDir()

	// Create two small .md files.
	writeFile(t, filepath.Join(dir, "a.md"), "# A\n\nContent of A.")
	writeFile(t, filepath.Join(dir, "b.md"), "# B\n\nContent of B.")

	outPath := filepath.Join(t.TempDir(), "combined.md")

	res, err := Run(CombineConfig{
		InputDir:   dir,
		OutputPath: outPath,
		MaxWords:   0, // no limit
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if res.FileCount != 2 {
		t.Errorf("FileCount = %d, want 2", res.FileCount)
	}
	if len(res.OutputFiles) != 1 {
		t.Errorf("OutputFiles = %v, want exactly 1", res.OutputFiles)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "Content of A") {
		t.Errorf("output missing 'Content of A'")
	}
	if !strings.Contains(content, "Content of B") {
		t.Errorf("output missing 'Content of B'")
	}
}

func TestCombineSplit(t *testing.T) {
	dir := t.TempDir()

	// Create files with 10 words each; set limit to 15 words to force split after 1st file.
	writeFile(t, filepath.Join(dir, "a.md"), "one two three four five six seven eight nine ten")
	writeFile(t, filepath.Join(dir, "b.md"), "one two three four five six seven eight nine ten")

	outBase := filepath.Join(t.TempDir(), "combined.md")

	res, err := Run(CombineConfig{
		InputDir:   dir,
		OutputPath: outBase,
		MaxWords:   15,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(res.OutputFiles) != 2 {
		t.Fatalf("expected split into 2 files, got %d: %v", len(res.OutputFiles), res.OutputFiles)
	}

	// Verify all split files have numbered names.
	outDir := filepath.Dir(outBase)
	want1 := filepath.Join(outDir, "combined-001.md")
	want2 := filepath.Join(outDir, "combined-002.md")

	if res.OutputFiles[0] != want1 {
		t.Errorf("first output = %q, want %q", res.OutputFiles[0], want1)
	}
	if res.OutputFiles[1] != want2 {
		t.Errorf("second output = %q, want %q", res.OutputFiles[1], want2)
	}

	for _, f := range res.OutputFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("output file %q not found: %v", f, err)
		}
	}

	// Verify unnumbered file does NOT exist.
	if _, err := os.Stat(outBase); err == nil {
		t.Errorf("unnumbered %q should not exist when splitting", outBase)
	}
}

func TestCombineEmpty(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(t.TempDir(), "combined.md")

	res, err := Run(CombineConfig{
		InputDir:   dir,
		OutputPath: outPath,
	})
	if err != nil {
		t.Fatalf("Run on empty dir: %v", err)
	}
	if res.FileCount != 0 {
		t.Errorf("expected 0 files, got %d", res.FileCount)
	}
}

func TestNumberedPath(t *testing.T) {
	tests := []struct {
		base string
		n    int
		want string
	}{
		{"combined.md", 1, "combined-001.md"},
		{"combined.md", 2, "combined-002.md"},
		{"/out/combined.md", 3, "/out/combined-003.md"},
	}
	for _, tc := range tests {
		got := numberedPath(tc.base, tc.n)
		if got != tc.want {
			t.Errorf("numberedPath(%q, %d) = %q, want %q", tc.base, tc.n, got, tc.want)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
