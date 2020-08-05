package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/td0m/pub-sub/pkg/clients"
)

func main() {
	token := os.Getenv("TOKEN")
	if len(token) == 0 {
		fmt.Println("Please provide a TOKEN")
		return
	}
	host := "localhost:8080/ws"
	events := []string{"users!"}
	client, err := clients.NewWsClient(host, token, events, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			msg, err := client.ReadMessage()
			if err != nil {
				return
			}
			fmt.Println(msg)
		}
	}()

	go func() {
		defer close(done)
		for {
			event, payload := readLine("event"), readLine("payload")
			client.Emit(event, payload)
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
