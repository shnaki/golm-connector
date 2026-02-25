package crawler

import (
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

// Normalize returns a canonical form of rawURL:
//   - scheme is lowercased
//   - host is lowercased
//   - fragment is removed
//   - trailing slash is normalised (root path keeps "/", others strip it)
//
// Returns "" on parse error.
func Normalize(rawURL string) string {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Fragment = ""
	if u.Path != "/" {
		u.Path = strings.TrimRight(u.Path, "/")
	}
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String()
}

// InScope reports whether target is within the scope defined by base.
// Scope is: same host AND target path has base path as a prefix.
func InScope(base, target string) bool {
	b, err := url.Parse(base)
	if err != nil {
		return false
	}
	t, err := url.Parse(target)
	if err != nil {
		return false
	}
	if !strings.EqualFold(b.Host, t.Host) {
		return false
	}
	scopePrefix := b.Path
	if !strings.HasSuffix(scopePrefix, "/") {
		scopePrefix = scopePrefix + "/"
	}
	return strings.HasPrefix(t.Path, scopePrefix) || t.Path == strings.TrimRight(scopePrefix, "/")
}

// ExtractLinks parses an HTML document and returns all href/src link URLs
// resolved against baseURL, de-duplicated and in document order.
func ExtractLinks(baseURL string, doc *html.Node) []string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	seen := make(map[string]struct{})
	var links []string

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			attr := ""
			switch n.Data {
			case "a":
				attr = getAttr(n, "href")
			case "link":
				attr = getAttr(n, "href")
			}
			if attr != "" {
				ref, err := url.Parse(attr)
				if err == nil {
					abs := base.ResolveReference(ref).String()
					abs = Normalize(abs)
					if abs != "" {
						if _, dup := seen[abs]; !dup {
							seen[abs] = struct{}{}
							links = append(links, abs)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return links
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// URLToFilename converts a URL to a safe relative file path ending in .html.
// e.g. https://example.com/docs/intro → example.com/docs/intro.html
func URLToFilename(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	p := strings.Trim(u.Path, "/")
	if p == "" {
		p = "index"
	}
	// Replace path separators that might cause issues; keep slashes for dirs.
	p = strings.ReplaceAll(p, "..", "__")
	if !strings.HasSuffix(p, ".html") {
		p += ".html"
	}
	return path.Join(u.Host, p)
}
