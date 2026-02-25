package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Write serializes r as pretty-printed JSON to path.
func Write(r *Report, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("report: create %s: %w", path, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("report: encode: %w", err)
	}
	return nil
}

// Load reads a previously written report from path.
func Load(path string) (*Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("report: read %s: %w", path, err)
	}
	var r Report
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("report: unmarshal: %w", err)
	}
	return &r, nil
}

// FailedURLs returns all URLs whose status is StatusError for the given step name.
func FailedURLs(r *Report, step string) []string {
	var out []string
	for _, s := range r.Steps {
		if s.Step != step {
			continue
		}
		for _, u := range s.URLs {
			if u.Status == StatusError {
				out = append(out, u.URL)
			}
		}
	}
	return out
}

// AddStep appends a StepResult to the report and returns a pointer to it for in-place updates.
func AddStep(r *Report, name string) *StepResult {
	r.Steps = append(r.Steps, StepResult{
		Step:      name,
		StartTime: time.Now().UTC(),
	})
	return &r.Steps[len(r.Steps)-1]
}

// Finish records the step end time.
func Finish(s *StepResult) {
	s.EndTime = time.Now().UTC()
}
