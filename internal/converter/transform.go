package converter

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var reMultiNewline = regexp.MustCompile(`\n{3,}`)

// Transform takes a content node, strips unwanted elements, converts it to
// Markdown, and normalises consecutive blank lines.
func Transform(node *html.Node, cfg *ConvertConfig) (string, error) {
	// Wrap the node in a goquery selection for easy removal operations.
	doc := goquery.NewDocumentFromNode(node)

	// Strip configured tags.
	for _, tag := range cfg.StripTags {
		doc.Find(tag).Remove()
	}

	// Strip configured CSS classes.
	for _, cls := range cfg.StripClasses {
		doc.Find("." + cls).Remove()
	}

	// Always strip images and inline SVG.
	doc.Find("img, svg, picture").Remove()

	// Render the cleaned node back to HTML for the converter.
	var buf bytes.Buffer
	if err := html.Render(&buf, doc.Get(0)); err != nil {
		return "", fmt.Errorf("transform: render html: %w", err)
	}

	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
		),
	)

	md, err := conv.ConvertString(buf.String())
	if err != nil {
		return "", fmt.Errorf("transform: convert: %w", err)
	}

	// Collapse 3+ consecutive newlines to exactly 2.
	md = reMultiNewline.ReplaceAllString(md, "\n\n")
	md = strings.TrimSpace(md) + "\n"

	return md, nil
}
