package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// LiveReload manages WebSocket connections and watches files for changes.
type LiveReload struct {
	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex
	watcher   *fsnotify.Watcher
}

// NewLiveReload initializes a new LiveReload instance.
func NewLiveReload() (*LiveReload, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &LiveReload{
		clients: make(map[*websocket.Conn]bool),
		watcher: watcher,
	}, nil
}

// Watch recursively watches a directory for changes and broadcasts reload events.
func (lr *LiveReload) Watch(root string) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			return lr.watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Println("Warning: Failed to watch some directories:", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-lr.watcher.Events:
				if !ok {
					return
				}
				// We care about write and create operations
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					log.Println("File changed, triggering reload:", event.Name)
					lr.broadcast("reload")
				}
			case err, ok := <-lr.watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()
}

// ServeHTTP upgrades the HTTP connection to a WebSocket.
func (lr *LiveReload) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("websocket upgrade error:", err)
		return
	}

	lr.clientsMu.Lock()
	lr.clients[c] = true
	lr.clientsMu.Unlock()

	defer func() {
		lr.clientsMu.Lock()
		delete(lr.clients, c)
		lr.clientsMu.Unlock()
		c.Close()
	}()

	// Keep connection alive until client disconnects
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}

// broadcast sends a message to all connected WebSocket clients.
func (lr *LiveReload) broadcast(msg string) {
	lr.clientsMu.Lock()
	defer lr.clientsMu.Unlock()
	for client := range lr.clients {
		if err := client.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			client.Close()
			delete(lr.clients, client)
		}
	}
}
