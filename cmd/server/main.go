package main

import (
    "fmt"
    "log"
    "mime"
    "net/http"
    "path/filepath"
    "strings"
    "omegle-clone/internal/handlers"
    "omegle-clone/internal/service"
    "omegle-clone/pkg/utils"
    "github.com/rs/cors"
)

// getMIMEType returns the correct MIME type for a given file extension
func getMIMEType(path string) string {
    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
    case ".html":
        return "text/html"
    case ".css":
        return "text/css"
    case ".js":
        return "application/javascript"
    default:
        return mime.TypeByExtension(ext)
    }
}

func main() {
    // Initialize config and service
    config := utils.LoadConfig()
    chatService := service.GetChatService()

    // Create new mux
    mux := http.NewServeMux()

    // Create file server
    fs := http.FileServer(http.Dir("static"))

    // Handle routes
    mux.HandleFunc("/ws", handlers.HandleWebSocket)
    mux.HandleFunc("/wsurl", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, chatService.GetWebSocketURL())
    })
    mux.Handle("/", fs)

    // Setup CORS
    corsHandler := cors.New(cors.Options{
        AllowedOrigins:   []string{"*"},
        AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "X-Requested-With", "Origin"},
        ExposedHeaders:   []string{"Content-Length"},
        AllowCredentials: true,
    })

    // Create handler with CORS
    handler := corsHandler.Handler(mux)

    // Start server
    serverAddr := ":" + config.Port
    log.Printf("Server starting on %s", serverAddr)
    
    if err := http.ListenAndServe(serverAddr, handler); err != nil {
        log.Printf("Server error: %v", err)
        return
    }
}
