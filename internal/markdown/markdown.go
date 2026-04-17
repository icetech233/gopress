package markdown

import (
	"bytes"
	"fmt"
	"os"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	lru "github.com/hashicorp/golang-lru/v2"
	goldmark_meta "github.com/icetech233/gopress/pkg/goldmark-meta"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// RenderResult contains the generated HTML and frontmatter metadata.
type RenderResult struct {
	HTML  string
	Meta  map[string]interface{}
	Title string
}

var cache, _ = lru.New[string, *RenderResult](1024)

// Render parses markdown file and returns HTML and metadata, with LRU caching.
func Render(filePath string) (*RenderResult, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	cacheKey := fmt.Sprintf("%s:%d", filePath, stat.ModTime().UnixNano())
	if val, ok := cache.Get(cacheKey); ok {
		return val, nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	result, err := RenderBytes(content)
	if err == nil {
		cache.Add(cacheKey, result)
	}
	return result, err
}

// RenderBytes parses markdown bytes and returns HTML and metadata.
func RenderBytes(content []byte) (*RenderResult, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			goldmark_meta.Meta,
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
				),
				//highlighting.WithStyle("github"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	context := parser.NewContext()
	var buf bytes.Buffer

	if err := md.Convert(content, &buf, parser.WithContext(context)); err != nil {
		return nil, fmt.Errorf("failed to convert markdown: %w", err)
	}

	metaData := goldmark_meta.Get(context)

	// Extract title from meta or fall back to empty
	title := ""
	if metaData != nil {
		if t, ok := metaData["title"].(string); ok {
			title = t
		}
	}

	return &RenderResult{
		HTML:  buf.String(),
		Meta:  metaData,
		Title: title,
	}, nil
}
