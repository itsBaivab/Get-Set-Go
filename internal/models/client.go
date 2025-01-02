package models

import (
    "github.com/google/uuid"
    "github.com/gorilla/websocket"
)

type Client struct {
    Conn     *websocket.Conn
    Partner  *Client
    SendChan chan []byte
    ID       string
    IsVideo  bool
}

func NewClient(conn *websocket.Conn) *Client {
    return &Client{
        Conn:     conn,
        SendChan: make(chan []byte, 256),
        ID:       generateID(),
    }
}

func generateID() string {
    return uuid.New().String()
}

func (c *Client) Close() {
    close(c.SendChan)
    c.Conn.Close()
}