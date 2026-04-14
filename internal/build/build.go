package build

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/icetech233/gopress/internal/config"
	"github.com/icetech233/gopress/internal/markdown"
	"github.com/icetech233/gopress/internal/theme"
	"golang.org/x/sync/errgroup"
)

// Builder holds context for building static site.
type Builder struct {
	Root       string
	OutDir     string
	SiteConfig *config.SiteConfig
}

// Build traverses the root directory and builds HTML for all .md files concurrently.
func (b *Builder) Build() error {
	outDir := b.OutDir
	if outDir == "" {
		outDir = filepath.Join(b.Root, ".gopress", "dist")
	}

	fmt.Printf("Building site into %s...\n", outDir)

	// Clean out dir
	os.RemoveAll(outDir)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write Lean Chunk Assets
	assetsDir := filepath.Join(outDir, "assets")
	os.MkdirAll(assetsDir, 0755)
	os.WriteFile(filepath.Join(assetsDir, "theme.css"), []byte(theme.GetThemeCSS()), 0644)
	os.WriteFile(filepath.Join(assetsDir, "app.js"), []byte(theme.GetThemeJS(false)), 0644) // false for prod

	var eg errgroup.Group
	sem := make(chan struct{}, 64) // Concurrency limit

	// Walk through root directory
	err := filepath.WalkDir(b.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories (like .git, .gopress, node_modules)
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." && d.Name() != ".gopress" {
				return filepath.SkipDir
			}
			if d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Process markdown files
		if strings.HasSuffix(d.Name(), ".md") {
			relPath, _ := filepath.Rel(b.Root, path)

			// Ignore .gopress internals
			if strings.HasPrefix(relPath, ".gopress") {
				return nil
			}

			pathCopy := path
			relPathCopy := relPath
			isReadme := d.Name() == "README.md"

			eg.Go(func() error {
				sem <- struct{}{}
				defer func() { <-sem }()

				// Render Markdown
				result, err := markdown.Render(pathCopy)
				if err != nil {
					return fmt.Errorf("failed to render %s: %w", pathCopy, err)
				}

				// Generate HTML
				relPath, _ := filepath.Rel(b.Root, pathCopy)
				urlPath := "/" + filepath.ToSlash(strings.TrimSuffix(relPath, ".md")) + ".html"

				htmlContent, err := theme.GenerateHTML(b.SiteConfig, result, urlPath)
				if err != nil {
					return fmt.Errorf("failed to generate HTML for %s: %w", pathCopy, err)
				}

				// Write to output directory
				outPath := filepath.Join(outDir, strings.TrimSuffix(relPathCopy, ".md")+".html")
				if isReadme {
					outPath = filepath.Join(outDir, filepath.Dir(relPathCopy), "index.html")
				}

				if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
					return fmt.Errorf("failed to create dir %s: %w", filepath.Dir(outPath), err)
				}

				if err := os.WriteFile(outPath, []byte(htmlContent), 0644); err != nil {
					return fmt.Errorf("failed to write %s: %w", outPath, err)
				}

				fmt.Printf("✓ Rendered %s\n", relPathCopy)
				return nil
			})
		} else if !strings.HasPrefix(d.Name(), ".") {
			// Copy static assets synchronously to avoid complex file descriptor limits on large repos
			relPath, _ := filepath.Rel(b.Root, path)
			if strings.HasPrefix(relPath, ".gopress") {
				return nil
			}

			outPath := filepath.Join(outDir, relPath)
			if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
				return err
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if err := os.WriteFile(outPath, content, 0644); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	fmt.Println("Build complete!")
	return nil
}
