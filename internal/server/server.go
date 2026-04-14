package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/icetech233/gopress/internal/config"
	"github.com/icetech233/gopress/internal/markdown"
	"github.com/icetech233/gopress/internal/theme"
)

// DevServer serves markdown files as HTML on the fly.
type DevServer struct {
	Root       string
	SiteConfig *config.SiteConfig
	lr         *LiveReload
}

// Start starts the development server on the given port.
func (s *DevServer) Start(port int) error {
	lr, err := NewLiveReload()
	if err != nil {
		return fmt.Errorf("failed to init live reload: %w", err)
	}
	s.lr = lr
	s.lr.Watch(s.Root)

	mux := http.NewServeMux()

	// WebSocket route for HMR / Live Reload
	mux.HandleFunc("/ws", s.lr.ServeHTTP)

	// Virtual routes for external CSS / JS assets
	mux.HandleFunc("/assets/theme.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(theme.GetThemeCSS()))
	})
	mux.HandleFunc("/assets/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(theme.GetThemeJS(true))) // true for dev mode (injects WebSocket client)
	})

	// Handle all other requests
	mux.HandleFunc("/", s.handleRequest)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("GoPress Go Dev Server listening on http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *DevServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Security: Prevent directory traversal
	if strings.Contains(r.URL.Path, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Hot reload config on each request to pick up changes
	siteConfig, err := config.LoadConfig(s.Root)
	if err == nil {
		s.SiteConfig = siteConfig
	}

	// Determine the corresponding markdown file
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Check if it's requesting an HTML file
	if strings.HasSuffix(path, ".html") {
		mdPath := strings.TrimSuffix(path, ".html") + ".md"
		fullMdPath := filepath.Join(s.Root, mdPath)

		// Check if file exists
		if _, err := os.Stat(fullMdPath); err == nil {
			// Read and render Markdown
			result, err := markdown.Render(fullMdPath)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error rendering markdown: %v", err), http.StatusInternalServerError)
				return
			}

			// Generate HTML page
			html, err := theme.GenerateHTML(s.SiteConfig, result, path)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error generating HTML: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(html))
			return
		}
	}

	// Serve static files if not an HTML/MD route
	fullStaticPath := filepath.Join(s.Root, path)
	if stat, err := os.Stat(fullStaticPath); err == nil && !stat.IsDir() {
		http.ServeFile(w, r, fullStaticPath)
		return
	}

	http.NotFound(w, r)
}
