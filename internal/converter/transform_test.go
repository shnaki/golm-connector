package converter

import (
	"strings"
	"testing"
)

func TestTransformBasic(t *testing.T) {
	const rawHTML = `<html><body>
<nav>Navigation</nav>
<main>
  <article>
    <h1>Hello</h1>
    <p>World</p>
  </article>
</main>
</body></html>`

	node, err := ExtractMainNode([]byte(rawHTML))
	if err != nil {
		t.Fatalf("ExtractMainNode: %v", err)
	}

	cfg := &ConvertConfig{
		StripTags: []string{"nav"},
	}

	md, err := Transform(node, cfg)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	if !strings.Contains(md, "Hello") {
		t.Errorf("expected 'Hello' in output, got:\n%s", md)
	}
	if !strings.Contains(md, "World") {
		t.Errorf("expected 'World' in output, got:\n%s", md)
	}
	// Navigation should be stripped.
	if strings.Contains(md, "Navigation") {
		t.Errorf("expected 'Navigation' to be stripped, got:\n%s", md)
	}
}

func TestTransformNewlineCollapse(t *testing.T) {
	const rawHTML = `<html><body>
<p>First</p>



<p>Second</p>
</body></html>`

	node, err := ExtractMainNode([]byte(rawHTML))
	if err != nil {
		t.Fatalf("ExtractMainNode: %v", err)
	}

	md, err := Transform(node, &ConvertConfig{})
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	// No 3+ consecutive newlines should remain.
	if strings.Contains(md, "\n\n\n") {
		t.Errorf("output contains 3+ consecutive newlines:\n%q", md)
	}
}

func TestTransformStripClasses(t *testing.T) {
	const rawHTML = `<html><body>
<div class="sidebar">Sidebar content</div>
<main><p>Main content</p></main>
</body></html>`

	node, err := ExtractMainNode([]byte(rawHTML))
	if err != nil {
		t.Fatalf("ExtractMainNode: %v", err)
	}

	cfg := &ConvertConfig{
		StripClasses: []string{"sidebar"},
	}

	md, err := Transform(node, cfg)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}

	if strings.Contains(md, "Sidebar content") {
		t.Errorf("sidebar class should be stripped, got:\n%s", md)
	}
	if !strings.Contains(md, "Main content") {
		t.Errorf("main content should remain, got:\n%s", md)
	}
}
