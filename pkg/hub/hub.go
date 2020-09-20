package hub

import "log"

const EVENT_CLIENT_UPDATE = "users!"

type clientUpdate struct {
	Clients map[string][]string `json:"clients"`
	Changed string              `json:"changed"`
	IsNew   bool                `json:"isNew"`
}

type Hub struct {
	emit       chan Message
	register   chan *Client
	unregister chan *Client
	clients    map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		emit:       make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    map[string]*Client{},
	}
}

func (self *Hub) Run() {
	for {
		select {
		case client := <-self.register:
			self.clients[client.id] = client
			log.Println("+ " + client.id)
			go self.emitClientUpdate(client.id, true)
		case client := <-self.unregister:
			if _, ok := self.clients[client.id]; ok {
				log.Println("- " + client.id)
				close(client.send)
				delete(self.clients, client.id)
				go self.emitClientUpdate(client.id, false)
			}
		case msg := <-self.emit:
			for _, client := range self.clients {
				if contains(client.events, msg.Event) {
					go func(c chan Message) { c <- msg }(client.send)
				}
			}
		}
	}
}

func (self *Hub) Clients() map[string][]string {
	c := map[string][]string{}
	for k, v := range self.clients {
		c[k] = v.Events()
	}
	return c
}

func contains(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

func (self *Hub) emitClientUpdate(id string, isNew bool) {
	clients := map[string][]string{}
	for id, client := range self.clients {
		clients[id] = client.events
	}
	self.emit <- Message{Event: EVENT_CLIENT_UPDATE, Payload: clientUpdate{Clients: clients, Changed: id, IsNew: isNew}}
}

func (self *Hub) Register(client *Client) {
	self.register <- client
}

func (self *Hub) Unregister(client *Client) {
	self.unregister <- client
}

func (self *Hub) Emit(msg Message) {
	self.emit <- msg
}
