package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"sfsEdgeStore/pathutil"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handleWebIndex 处理 Web 界面的索引请求
func (s *Server) handleWebIndex(w http.ResponseWriter, r *http.Request) {
	webDir, err := pathutil.Join("web")
	if err != nil {
		webDir = "web"
	}

	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		webDir = "web"
	}

	switch r.URL.Path {
	case "/", "/dashboard":
		indexFile := filepath.Join(webDir, "index.html")
		if _, err := os.Stat(indexFile); os.IsNotExist(err) {
			http.Error(w, "Web interface not found. Please ensure the 'web' directory exists.", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, indexFile)
	case "/subscription-topics":
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		s.handleStaticFiles(w, r)
	}
}

// handleStaticFiles 处理 Web 界面的静态文件请求
func (s *Server) handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	log.Printf("Static file request: %s", r.URL.Path)

	webDir, err := pathutil.Join("web")
	if err != nil {
		webDir = "web"
	}

	if _, err := os.Stat(webDir); err != nil {
		log.Println("Web directory not found")
		http.Error(w, "Web directory not found", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(webDir, "static", filepath.FromSlash(r.URL.Path[len("/static/"):]))
	log.Printf("Trying to serve file: %s", filePath)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("File not found: %s", filePath)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	ext := filepath.Ext(filePath)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".woff":
		w.Header().Set("Content-Type", "font/woff")
	case ".woff2":
		w.Header().Set("Content-Type", "font/woff2")
	case ".ttf":
		w.Header().Set("Content-Type", "font/ttf")
	case ".eot":
		w.Header().Set("Content-Type", "application/vnd.ms-fontobject")
	}

	log.Printf("Serving file: %s", filePath)
	http.ServeFile(w, r, filePath)
}

// handleWebSocket 处理 WebSocket 连接请求
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	s.wsManager.Register(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			s.wsManager.Unregister(conn)
			break
		}
	}
}
