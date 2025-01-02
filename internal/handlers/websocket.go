package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

const maxMessageSize = 1024 * 8
const pongWait = 60 * time.Second

var upgrader = websocket.Upgrader{
    ReadBufferSize:  maxMessageSize,
    WriteBufferSize: maxMessageSize,
    CheckOrigin:     func(r *http.Request) bool { return true },
}

type Client struct {
    conn    *websocket.Conn
    partner *Client
    send    chan []byte
    mu      sync.Mutex
}

type Message struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data"`
}

var (
    clients    = make(map[*Client]bool)
    register   = make(chan *Client)
    unregister = make(chan *Client)
    mutex      sync.RWMutex
)

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Upgrade HTTP connection to WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Upgrade error: %v", err)
        return
    }

    // Create new client
    client := &Client{
        conn: conn,
        send: make(chan []byte, 256),
    }

    // Register client
    register <- client

    // Start read/write pumps
    go client.writePump()
    go client.readPump()
}

func (c *Client) readPump() {
    defer func() {
        unregister <- c
        c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("read error: %v", err)
            }
            break
        }

        // Forward message to partner
        var msg map[string]interface{}
        if err := json.Unmarshal(message, &msg); err != nil {
            log.Printf("JSON parse error: %v", err)
            continue
        }

        // Process the valid JSON message
        c.mu.Lock()
        partner := c.partner
        c.mu.Unlock()

        if partner != nil {
            select {
            case partner.send <- message:
            default:
                log.Printf("Failed to send message to partner")
            }
        }
    }
}

func (c *Client) writePump() {
    defer c.conn.Close()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                return
            }

            err := c.conn.WriteMessage(websocket.TextMessage, message)
            if err != nil {
                return
            }
        }
    }
}

func init() {
    go handleConnections()
}

func handleConnections() {
    for {
        select {
        case client := <-register:
            mutex.Lock()
            var partner *Client
            for c := range clients {
                if c.partner == nil && c != client {
                    partner = c
                    break
                }
            }

            if partner != nil {
                client.partner = partner
                partner.partner = client
                msg := []byte(`{"type":"connected"}`)
                partner.send <- msg
                client.send <- msg
            }

            clients[client] = true
            mutex.Unlock()

        case client := <-unregister:
            mutex.Lock()
            if _, ok := clients[client]; ok {
                if client.partner != nil {
                    client.partner.partner = nil
                    client.partner.send <- []byte(`{"type":"disconnected"}`)
                }
                delete(clients, client)
                close(client.send)
            }
            mutex.Unlock()
        }
    }
}

func handleWebRTCSignaling(client *Client) {
    for {
        _, message, err := client.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("read error: %v", err)
            }
            break
        }

        // Validate message structure
        var msg struct {
            Type string          `json:"type"`
            SDP  json.RawMessage `json:"sdp,omitempty"`
            Candidate json.RawMessage `json:"candidate,omitempty"`
        }

        if err := json.Unmarshal(message, &msg); err != nil {
            log.Printf("JSON parse error: %v", err)
            continue
        }

        // Forward message to partner if exists
        client.mu.Lock()
        partner := client.partner
        client.mu.Unlock()

        if partner != nil {
            if err := partner.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                log.Printf("write error: %v", err)
                break
            }
        }
    }
}