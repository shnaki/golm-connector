package report

import "time"

// Status represents the result of processing a single URL.
type Status string

const (
	StatusOK      Status = "ok"
	StatusSkipped Status = "skipped"
	StatusError   Status = "error"
)

// URLStatus records the outcome of crawling or converting a single URL/file.
type URLStatus struct {
	URL    string    `json:"url"`
	Status Status    `json:"status"`
	Error  string    `json:"error,omitempty"`
	Time   time.Time `json:"time"`
}

// StepResult holds the aggregate result of one pipeline step.
type StepResult struct {
	Step      string      `json:"step"`
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
	URLs      []URLStatus `json:"urls"`
}

// Report is the top-level structure serialized to JSON.
type Report struct {
	Version   string       `json:"version"`
	CreatedAt time.Time    `json:"created_at"`
	Steps     []StepResult `json:"steps"`
}

// NewReport creates a Report with the current timestamp.
func NewReport() *Report {
	return &Report{
		Version:   "1",
		CreatedAt: time.Now().UTC(),
	}
}
