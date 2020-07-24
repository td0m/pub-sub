package hub

import "log"

const EVENT_CLIENT_UPDATE = "users!"

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
			for _, v := range self.clients {
				go func(c chan Message) { c <- msg }(v.send)
			}
		}

	}
}

type clientUpdate struct {
	Clients map[string][]string `json:"clients"`
	Changed string              `json:"changed"`
	IsNew   bool                `json:"isNew"`
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
