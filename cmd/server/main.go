package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		var msg Message
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Print("read:", err)
			return
		}
		err = c.WriteJSON(msg)
		if err != nil {
			log.Print("write:", err)
			return
		}
	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe(":8080", nil)
}
