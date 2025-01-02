package service

import (
    "log"
    "os"
    "strings"
    "sync"
    "omegle-clone/internal/models"
    "github.com/gorilla/websocket"
)

type ChatService struct {
    clients      map[*models.Client]bool
    waitingQueue []*models.Client
    mutex        *sync.Mutex
}

var (
    chatService *ChatService
    once        sync.Once
)

func GetChatService() *ChatService {
    once.Do(func() {
        chatService = &ChatService{
            clients:      make(map[*models.Client]bool),
            waitingQueue: make([]*models.Client, 0),
            mutex:        &sync.Mutex{},
        }
    })
    return chatService
}

func (cs *ChatService) RegisterClient(client *models.Client) {
    cs.mutex.Lock()
    defer cs.mutex.Unlock()

    cs.clients[client] = true
    if len(cs.waitingQueue) > 0 {
        partner := cs.waitingQueue[0]
        cs.waitingQueue = cs.waitingQueue[1:]
        cs.pairClients(client, partner)
    } else {
        cs.waitingQueue = append(cs.waitingQueue, client)
    }

    go cs.handleMessages(client)
    go cs.handleSend(client)
}

// [internal/service/chat.go](internal/service/chat.go)
func (cs *ChatService) pairClients(client1, client2 *models.Client) {
    client1.Partner = client2
    client2.Partner = client1

    log.Printf("Paired Client %s with Client %s", client1.ID, client2.ID)

    // Notify both clients
    client1.SendChan <- []byte("Connected to a stranger!")
    client2.SendChan <- []byte("Connected to a stranger!")
}

func (cs *ChatService) handleMessages(client *models.Client) {
    defer func() {
        cs.disconnectClient(client)
    }()

    for {
        _, message, err := client.Conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("error: %v", err)
            }
            break
        }

        if client.Partner != nil {
            client.Partner.SendChan <- message
        }
    }
}

func (cs *ChatService) handleSend(client *models.Client) {
    for message := range client.SendChan {
        err := client.Conn.WriteMessage(websocket.TextMessage, message)
        if err != nil {
            log.Printf("error writing message: %v", err)
            break
        }
    }
}

func (cs *ChatService) disconnectClient(client *models.Client) {
    cs.mutex.Lock()
    defer cs.mutex.Unlock()

    // Remove from clients map
    delete(cs.clients, client)

    // Remove from waiting queue if present
    for i, c := range cs.waitingQueue {
        if c == client {
            cs.waitingQueue = append(cs.waitingQueue[:i], cs.waitingQueue[i+1:]...)
            break
        }
    }

    // Notify partner if exists
    if client.Partner != nil {
        client.Partner.SendChan <- []byte("Stranger has disconnected!")
        client.Partner.Partner = nil
        // Add partner back to waiting queue
        cs.waitingQueue = append(cs.waitingQueue, client.Partner)
    }

    close(client.SendChan)
    client.Conn.Close()
}

func (cs *ChatService) broadcastSystemMessage(message string) {
    for client := range cs.clients {
        client.SendChan <- []byte(message)
    }
}

func (cs *ChatService) GetWebSocketURL() string {
    ngrokURL := os.Getenv("NGROK_URL")
    if ngrokURL != "" {
        ngrokURL = strings.TrimRight(ngrokURL, "/")
        return strings.Replace(ngrokURL, "http", "ws", 1) + "/ws"
    }
    return "ws://localhost:8080/ws"
}