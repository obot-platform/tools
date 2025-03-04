package clean

import (
	"bytes"
	"fmt"

	md "github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/PuerkitoBio/goquery"
)

var tagsToRemove = []string{
	"script, style, noscript, meta, head",
	"header", "footer", "nav", "aside", ".header", ".top", ".navbar", "#header",
	".footer", ".bottom", "#footer", ".sidebar", ".side", ".aside", "#sidebar",
	".modal", ".popup", "#modal", ".overlay", ".ad", ".ads", ".advert", "#ad",
	".lang-selector", ".language", "#language-selector", ".social", ".social-media",
	".social-links", "#social", ".menu", ".navigation", "#nav", ".breadcrumbs",
	"#breadcrumbs", "#search-form", ".search", "#search", ".share", "#share",
	".widget", "#widget", ".cookie", "#cookie",
}

func Clean(in []byte) ([]byte, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(in))
	if err != nil {
		return nil, fmt.Errorf("failed to parse html input: %v", err)
	}
	for _, tag := range tagsToRemove {
		doc.Find(tag).Remove()
	}
	html, err := doc.Html()
	if err != nil {
		return nil, fmt.Errorf("failed to get html from document: %v", err)
	}
	return []byte(html), nil
}

func ToMarkdown(in []byte) ([]byte, error) {
	converter := md.NewConverter(md.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin()))
	return converter.ConvertReader(bytes.NewBuffer(in))
}
