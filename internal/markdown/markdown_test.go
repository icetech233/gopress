package markdown_test

import (
	"testing"

	"github.com/icetech233/gopress/internal/markdown"
)

func TestRender(t *testing.T) {
	filePath := "../../docs/history/themeable-logo.md"

	resp, err := markdown.Render(filePath)
	if err != nil {
		t.Fatalf("failed to render markdown: %v", err)
	}
	t.Logf("resp:%v", resp)
}
