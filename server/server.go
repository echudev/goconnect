package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketSender struct {
	clients map[*websocket.Conn]bool
}

func NewWebSocketSender() *WebSocketSender {
	return &WebSocketSender{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (wss *WebSocketSender) SendData(data map[string]float64) {
	for client := range wss.clients {
		err := client.WriteJSON(data)
		if err != nil {
			log.Printf("Error sending data: %v", err)
			client.Close()
			delete(wss.clients, client)
		}
	}
}

func HandleConnections(w http.ResponseWriter, r *http.Request, wss *WebSocketSender) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	wss.clients[ws] = true
}

func StartServer(dataChan <-chan map[string]float64) {
	wss := NewWebSocketSender()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		HandleConnections(w, r, wss)
	})

	go func() {
		for data := range dataChan {
			wss.SendData(data)
		}
	}()

	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
