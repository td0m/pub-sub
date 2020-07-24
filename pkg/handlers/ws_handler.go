package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/td0m/pub-sub/pkg/auth"
	"github.com/td0m/pub-sub/pkg/hub"
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

type wsHandler struct {
	hub    *hub.Hub
	conn   *websocket.Conn
	client hub.Client
}

func NewWsHandler(h *hub.Hub) http.HandlerFunc {
	return auth.WithClaims(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value("claims").(*auth.Claims)
		eventsParam := r.URL.Query().Get("events")
		events := strings.Split(strings.ToLower(eventsParam), ",")

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}

		handler := wsHandler{
			hub:    h,
			client: hub.NewClient(claims.Id, events),
			conn:   conn,
		}
		h.Register(&handler.client)
		go handler.reader()
		go handler.writer()
	})
}

func (self *wsHandler) reader() {
	defer self.conn.Close()
	defer self.hub.Unregister(&self.client)
	defer self.conn.SetReadDeadline(time.Now().Add(pongTimeout))
	self.conn.SetPongHandler(func(string) error {
		self.conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	for {
		var msg hub.Message
		err := self.conn.ReadJSON(&msg)
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return
			}
			log.Printf("error: %v", err)
		}
		self.hub.Emit(msg)
	}
}

func (self *wsHandler) writer() {
	defer self.conn.Close()
	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case <-ticker.C:
			self.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := self.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case msg, ok := <-self.client.Message():
			if !ok {
				self.conn.WriteMessage(websocket.CloseMessage, []byte("error"))
				return
			}
			self.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := self.conn.WriteJSON(msg)
			if err != nil {
				return
			}
		}
	}
}
