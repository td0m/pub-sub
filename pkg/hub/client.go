package hub

type Client struct {
	id     string
	send   chan Message
	events []string
}

func NewClient(id string, events []string) Client {
	return Client{id: id, events: events, send: make(chan Message)}
}

func (self *Client) Message() chan Message {
	return self.send
}
