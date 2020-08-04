package clients

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/td0m/pub-sub/pkg/hub"
)

type wsClient struct {
	ws *websocket.Conn
}

func NewWsClient(host, token string, events []string, secure bool) (*wsClient, error) {
	protocol := "ws"
	if secure {
		protocol = "wss"
	}
	url := fmt.Sprintf("%s://%s?events=%s", protocol, host, strings.Join(events, ","))
	ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{
		"Authorization": []string{"Bearer " + token},
	})
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
