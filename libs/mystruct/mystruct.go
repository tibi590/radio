package mystruct

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
    Conn *websocket.Conn
    Send chan []byte
    ConnLock sync.Mutex
    FrameQueue chan []byte
    Name string
}

func (client *Client) WriteToClient(messageType int, data []byte) error {
    client.ConnLock.Lock()
    defer client.ConnLock.Unlock()
    return client.Conn.WriteMessage(messageType, data)
}

type Button struct {
    Name      string
    Num       int
    IsToogle  bool
}

type IndexTemplate struct {
    Buttons []Button
}