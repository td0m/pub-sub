package main

import (
	"log"
	"math/rand"
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

type Client struct {
	id     string
	send   chan Message
	events []string
}

func NewClient(id string, events []string) Client {
	return Client{id: id, events: events, send: make(chan Message)}
}

type Hub struct {
	emit       chan Message
	register   chan *Client
	unregister chan *Client
	clients    map[string]*Client
}

func (self *Hub) Run() {
	for {
		select {
		case client := <-self.register:
			self.clients[client.id] = client
			log.Println("+ " + client.id)
			go self.emitClientUpdate()
		case client := <-self.unregister:
			if _, ok := self.clients[client.id]; ok {
				log.Println("- " + client.id)
				close(client.send)
				delete(self.clients, client.id)
				go self.emitClientUpdate()
			}
		case msg := <-self.emit:
			for _, v := range self.clients {
				go func(c chan Message) { c <- msg }(v.send)
			}
		}

	}
}

const EVENT_CLIENT_UPDATE = "users!"

func (self *Hub) emitClientUpdate() {
	clients := map[string][]string{}
	for id, client := range self.clients {
		clients[id] = client.events
	}
	self.emit <- Message{Event: EVENT_CLIENT_UPDATE, Payload: clients}
}

func (self *Hub) Register(client *Client) {
	self.register <- client
}

func (self *Hub) Unregister(client *Client) {
	self.unregister <- client
}

var hub *Hub

func NewHub() *Hub {
	return &Hub{
		emit:       make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    map[string]*Client{},
	}
}

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
	client := NewClient(randId(10), events)
	hub.Register(&client)
	go reader(c, &client)
	go writer(c, &client)
}

func reader(c *websocket.Conn, client *Client) {
	defer c.Close()
	defer hub.Unregister(client)
	defer c.SetReadDeadline(time.Now().Add(pongTimeout))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	for {
		var msg Message
		err := c.ReadJSON(&msg)
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return
			}
			log.Printf("error: %v", err)
		}
		hub.emit <- msg
		// TODO: emit msg
	}
}

func writer(c *websocket.Conn, client *Client) {
	defer c.Close()
	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case <-ticker.C:
			c.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case msg := <-client.send:
			c.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.WriteJSON(msg)
			if err != nil {
				return
			}
		}
	}
}

func main() {
	hub = NewHub()
	go hub.Run()
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe(":8080", nil)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randId(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
