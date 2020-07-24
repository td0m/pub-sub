package clients

import (
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/td0m/pub-sub/pkg/hub"
)

type wsClient struct {
	ws *websocket.Conn
}

func NewWsClient(host string, events []string) (*wsClient, error) {
	url := fmt.Sprintf("ws://%s?events=%s", host, strings.Join(events, ","))
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return &wsClient{ws: ws}, nil
}

func (self *wsClient) Emit(event string, payload interface{}) error {
	return self.ws.WriteJSON(hub.Message{Event: event, Payload: payload})
}

func (self *wsClient) ReadMessage() (*hub.Message, error) {
	var msg hub.Message
	err := self.ws.ReadJSON(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (self *wsClient) Close() error {
	return self.ws.Close()
}
