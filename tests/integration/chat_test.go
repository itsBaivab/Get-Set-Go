package integration

import (
    "net/http/httptest"
    "strings"
    "testing"
    "omegle-clone/internal/handlers"

    "github.com/gorilla/websocket"
)

func TestWebSocketConnection(t *testing.T) {
    s := httptest.NewServer(http.HandlerFunc(handlers.HandleWebSocket))
    defer s.Close()

    // Convert http://... to ws://...
    url := "ws" + strings.TrimPrefix(s.URL, "http")

    // Connect to the server
    ws, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        t.Fatalf("%v", err)
    }
    defer ws.Close()
}