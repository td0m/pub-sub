package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

func main() {
	host := "localhost:8080/ws"
	events := []string{"users!"}
	url := fmt.Sprintf("ws://%s?events=%s", host, strings.Join(events, ","))
	ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		fmt.Println(err)
		return
	}
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			event, payload := readLine("event"), readLine("payload")
			ws.WriteJSON(map[string]string{"event": event, "payload": payload})
		}
	}()

	go func() {
		defer close(done)
		for {
			var msg interface{}
			err := ws.ReadJSON(&msg)
			if err != nil {
				close(done)
			}
			fmt.Println(msg)
		}
	}()

	<-done
}

func readLine(field string) string {
	r := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", field)
	text, _ := r.ReadString('\n')
	return text[:len(text)-1]
}
