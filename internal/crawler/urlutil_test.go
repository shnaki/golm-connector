package crawler

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple URL keeps scheme and host",
			input: "https://example.com/docs",
			want:  "https://example.com/docs",
		},
		{
			name:  "uppercase scheme and host lowercased",
			input: "HTTPS://EXAMPLE.COM/docs",
			want:  "https://example.com/docs",
		},
		{
			name:  "fragment is removed",
			input: "https://example.com/docs#section",
			want:  "https://example.com/docs",
		},
		{
			name:  "trailing slash stripped on non-root",
			input: "https://example.com/docs/",
			want:  "https://example.com/docs",
		},
		{
			name:  "root path keeps slash",
			input: "https://example.com/",
			want:  "https://example.com/",
		},
		{
			name:  "root without trailing slash",
			input: "https://example.com",
			want:  "https://example.com/",
		},
		{
			name:  "relative URL returns empty",
			input: "/relative/path",
			want:  "",
		},
		{
			name:  "empty string returns empty",
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Normalize(tc.input)
			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestInScope(t *testing.T) {
	tests := []struct {
		name   string
		base   string
		target string
		want   bool
	}{
		{
			name:   "same host same path prefix",
			base:   "https://example.com/docs",
			target: "https://example.com/docs/page",
			want:   true,
		},
		{
			name:   "same host different subtree",
			base:   "https://example.com/docs",
			target: "https://example.com/blog/post",
			want:   false,
		},
		{
			name:   "different host",
			base:   "https://example.com/docs",
			target: "https://other.com/docs/page",
			want:   false,
		},
		{
			name:   "base is root",
			base:   "https://example.com/",
			target: "https://example.com/anything",
			want:   true,
		},
		{
			name:   "target equals scope root",
			base:   "https://example.com/docs",
			target: "https://example.com/docs/",
			want:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := InScope(tc.base, tc.target)
			if got != tc.want {
				t.Errorf("InScope(%q, %q) = %v, want %v", tc.base, tc.target, got, tc.want)
			}
		})
	}
}

func TestExtractLinks(t *testing.T) {
	const rawHTML = `<html><body>
<a href="/page1">Page 1</a>
<a href="https://external.com/page">External</a>
<a href="/page1">Duplicate</a>
<a href="#fragment">Fragment only</a>
</body></html>`

	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		t.Fatalf("parse HTML: %v", err)
	}

	base := "https://example.com/"
	links := ExtractLinks(base, doc)

	// Should contain resolved /page1 and the external link.
	wantContains := "https://example.com/page1"
	found := false
	for _, l := range links {
		if l == wantContains {
			found = true
		}
	}
	if !found {
		t.Errorf("ExtractLinks did not contain %q; got %v", wantContains, links)
	}

	// Should not contain duplicates.
	seen := make(map[string]int)
	for _, l := range links {
		seen[l]++
	}
	for link, count := range seen {
		if count > 1 {
			t.Errorf("duplicate link %q appeared %d times", link, count)
		}
	}
}

func TestURLToFilename(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/docs/intro", "example.com/docs/intro.html"},
		{"https://example.com/", "example.com/index.html"},
		{"https://example.com", "example.com/index.html"},
	}
	for _, tc := range tests {
		got := URLToFilename(tc.url)
		if got != tc.want {
			t.Errorf("URLToFilename(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}
