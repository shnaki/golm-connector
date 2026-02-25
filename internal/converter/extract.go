package converter

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// ExtractMainNode parses htmlBytes and returns the *html.Node for the most
// specific "main content" element found, in priority order:
//
//  1. main > article
//  2. main
//  3. [role=main]
//  4. article
//  5. body
//  6. root document node (fallback)
func ExtractMainNode(htmlBytes []byte) (*html.Node, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlBytes))
	if err != nil {
		return nil, err
	}

	selectors := []string{
		"main article",
		"main",
		"[role=main]",
		"article",
		"body",
	}

	for _, sel := range selectors {
		if node := doc.Find(sel).First(); node.Length() > 0 {
			if n := node.Get(0); n != nil {
				return n, nil
			}
		}
	}

	// Absolute fallback: return the document node itself.
	return doc.Get(0), nil
}
