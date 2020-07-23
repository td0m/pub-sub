package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// how often the server will ping clients for keepalive pongs
	pingPeriod = time.Second * 10
	// how long to wait for pongs
	// must be larger than pingPeriod
	pongTimeout = pingPeriod + time.Second*2
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
	eventsParam := r.URL.Query().Get("events")
	events := strings.Split(strings.ToLower(eventsParam), ",")
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	log.Println(events)
	go reader(c)
	go writer(c)
}

func reader(c *websocket.Conn) {
	defer c.Close()
	c.SetReadDeadline(time.Now().Add(pongTimeout))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	for {
		var msg Message
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("read: ", err)
		}
		log.Println(msg)
		// TODO: emit msg
	}
}

func writer(c *websocket.Conn) {
	defer c.Close()
	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case <-ticker.C:
			c.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe(":8080", nil)
}
