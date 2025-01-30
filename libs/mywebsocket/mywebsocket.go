package mywebsocket

import (
	"radio_site/libs/myconst"
	"radio_site/libs/myfile"
	"radio_site/libs/myhelper"
	"radio_site/libs/myparallel"
	"radio_site/libs/mystruct"
	"strings"
	"sync/atomic"

	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	holdingCommandPrefix = "h"
	userListCommandPrefix = "u"
)

var (
	socketReadError = []byte("RE")
	socketWriteError = []byte("WE")
	socketClosed = []byte("closed")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		host := r.Header.Get("X-Host")
		if host == "" {
			host = r.Host
		}

		if strings.Contains(host, "localhost") {
			return true
		}

		u, err := url.Parse(r.Header["Origin"][0])
		if err != nil {
			log.Println("origin error: ", err)
			return false
		}

		return strings.EqualFold(u.Hostname(), host)
	},
}

var (
	Clients sync.Map
	ClientCount atomic.Int64 // sync.Map doesnt have any len function
)

var ButtonsHeld sync.Map

func clientsToString() string {
	var builder strings.Builder
	Clients.Range(func(key, value any) bool {
		builder.WriteString(key.(*mystruct.Client).Name)
		builder.WriteByte(',')
		return true
	})
	return builder.String()
}

func holdingClientsToString() string {
	var builder strings.Builder
	ButtonsHeld.Range(func(key, value any) bool {
		builder.WriteString(value.(*mystruct.Client).Name)
		builder.WriteByte(';')
		builder.WriteString(strconv.Itoa(key.(int)))
		builder.WriteByte(',')
		return true
	})
	return builder.String()
}

func applyHeldButtons(statuses []byte) []byte {
	ButtonsHeld.Range(func(key, value any) bool {
		pin := key.(int)
		myhelper.InvertStatusByte(statuses, pin)
		return true
	})
	return statuses
}

func startFrameSender(client *mystruct.Client) {
	defer close(client.FrameQueue)
	for frame := range client.FrameQueue {
		if err := client.WriteToClient(websocket.BinaryMessage, frame); err != nil {
			return
		}
	}
}

func addClient(client *mystruct.Client) {
	Clients.Store(client, struct{}{})
	ClientCount.Add(1)
	go startFrameSender(client)
	go readMessages(client)
	log.Printf("%s connected. Total clients: %d", client.Name, ClientCount.Load())
}

func removeClient(client *mystruct.Client) {
	Clients.Delete(client)
	users := clientsToString()
	ClientCount.Add(-1)

	client.Conn.Close()

	ButtonsHeld.Range(func(key, value any) bool {
		if value != client { return true }

		ButtonsHeld.Delete(key)
		return false
	})

	usersHolding := holdingClientsToString()

	log.Printf("%s disconnected. Total clients: %d", client.Name, ClientCount.Load())
	broadcast([]byte(userListCommandPrefix + users))
	statuses := myfile.ReadPinStatuses()
	if statuses == nil {
		broadcast(socketReadError)
		return
	}
	broadcast([]byte(holdingCommandPrefix + usersHolding))
	broadcast(applyHeldButtons(statuses))
}

func broadcast(text []byte) {
	Clients.Range(func(key, value any) bool {
		key.(*mystruct.Client).Send <- text
		return true
	})
}

func WsHandler(res http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	name := req.Header.Get("X-User")
	if name == "" {
		name = req.Header.Get("X-Real-IP")
		if name == "" {
			name = req.RemoteAddr
		}
		log.Println(name, "has no name")
	} else {
		name += " (" + req.Header.Get("X-Real-IP") + ")"
	}

	client := &mystruct.Client{
		Conn: conn,
		Send: make(chan []byte),
		FrameQueue: make(chan []byte, 5),
		Name: name,
	}

	addClient(client)
	defer removeClient(client)

	// wait maximum readTimeout second for pong
	conn.SetReadDeadline(time.Now().Add(myconst.READ_TIMEOUT))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(myconst.READ_TIMEOUT))
		return nil
	})

	// ping every heartbeatTimeout second
	heartbeatTicker := time.NewTicker(myconst.HEARTBEAT_TIMEOUT)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-heartbeatTicker.C:
			// send ping
			if err := client.WriteToClient(websocket.PingMessage, nil); err != nil {
				log.Println(client.Name, "ping failed. Closing connection:", err)
				return
			}
		case message, ok := <-client.Send:
			if !ok {
				// Channel closed, terminate connection
				log.Println(client.Name, "channel closed")
				client.WriteToClient(websocket.TextMessage, socketClosed)
				return
			}

			// Send the message to the client
			if err := client.WriteToClient(websocket.TextMessage, message); err != nil {
				log.Println(client.Name, "write error, closing connection:", err)
				return
			}
		}
	}
}

func readMessages(client *mystruct.Client) {
	defer close(client.Send)

	statuses := myfile.ReadPinStatuses()
	if statuses == nil {
		broadcast(socketReadError)
		return
	}
	client.Send <- applyHeldButtons(statuses)
	users := clientsToString()

	usersHolding := holdingClientsToString()
	broadcast([]byte(holdingCommandPrefix + usersHolding))
	broadcast([]byte(userListCommandPrefix + users))

	for {
		msgType, message, err := client.Conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			return
		} else if err != nil {
			log.Println("Read error:", err)
			return
		}

		if msgType != websocket.TextMessage {
			continue
		}

		// check if message is number and in range of max pin number
		pin, err := strconv.Atoi(string(message))
		if err != nil || pin >= myconst.MAX_NUMBER_OF_PINS {
			continue
		}

		modes := myfile.ReadPinModes()
		isToggleButton := modes[pin] == 'T'

		var statuses []byte

		if !isToggleButton {
			statuses = myfile.ReadPinStatuses()
			if statuses == nil {
				broadcast(socketReadError)
				return
			}

			value, loaded := ButtonsHeld.LoadOrStore(pin, client)
			if value != client { continue }       // if button is not held by requesting user, deny it
			if loaded { ButtonsHeld.Delete(pin) } // if button already held by requesting user, release it

			usersHolding := holdingClientsToString()
			broadcast([]byte(holdingCommandPrefix + usersHolding))
		} else {
			statuses = myhelper.TogglePinStatus(pin)
			if statuses == nil {
				broadcast(socketWriteError)
				return
			}
		}

		applyHeldButtons(statuses)
		myparallel.WritePort(statuses)
		broadcast(statuses)
	}
}
