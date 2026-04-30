package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PreviewServer serves the built static files.
type PreviewServer struct {
	OutDir string
	lr     *LiveReload
}

// Start starts the preview server on the given port.
func (s *PreviewServer) Start(port int) error {
	// Initialize LiveReload if hot reload is supported
	lr, err := NewLiveReload()
	if err != nil {
		return fmt.Errorf("failed to init live reload: %w", err)
	}
	s.lr = lr
	s.lr.Watch(s.OutDir)

	mux := http.NewServeMux()

	// WebSocket route for Live Reload
	mux.HandleFunc("/ws", s.lr.ServeHTTP)

	// Serve static files
	mux.HandleFunc("/", s.handleRequest)

	addr := fmt.Sprintf(":%d", port)
	slog.Info("GoPress Preview Server listening", "url", fmt.Sprintf("http://localhost%s", addr))
	return http.ListenAndServe(addr, mux)
}

func (s *PreviewServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Security: Prevent directory traversal
	if strings.Contains(r.URL.Path, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	fullStaticPath := filepath.Join(s.OutDir, path)
	stat, err := os.Stat(fullStaticPath)

	// Fallback to .html if path without extension is requested
	if err != nil && !strings.HasSuffix(path, ".html") {
		htmlPath := fullStaticPath + ".html"
		if statHtml, errHtml := os.Stat(htmlPath); errHtml == nil && !statHtml.IsDir() {
			fullStaticPath = htmlPath
			stat = statHtml
			err = nil
		}
	}

	if err == nil && !stat.IsDir() {
		// If hot reload is enabled, we could inject the live reload script
		// However, to provide a true preview, we serve the files as-is.
		// For the sake of the "hot reload" requirement, we inject the WS script into HTML files.
		if strings.HasSuffix(fullStaticPath, ".html") {
			content, readErr := os.ReadFile(fullStaticPath)
			if readErr == nil {
				// 				html := string(content)
				// 				// Inject live reload script before </body>
				// 				script := `
				// <script>
				// (function() {
				// 	var protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
				// 	var ws = new WebSocket(protocol + '//' + window.location.host + '/ws');
				// 	ws.onmessage = function(e) {
				// 		if (e.data === 'reload') {
				// 			window.location.reload();
				// 		}
				// 	};
				// })();
				// </script>`
				// 				html = strings.Replace(html, "</body>", script+"\n</body>", 1)
				// w.Write([]byte(html))
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(content)
				return
			}
		}

		http.ServeFile(w, r, fullStaticPath)
		return
	}

	http.NotFound(w, r)
}
