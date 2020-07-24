package main

import (
	"net/http"

	"github.com/td0m/pub-sub/pkg/handlers"
	"github.com/td0m/pub-sub/pkg/hub"
)

var h *hub.Hub

func main() {
	h = hub.NewHub()
	go h.Run()
	http.HandleFunc("/ws", handlers.NewWsHandler(h))
	http.ListenAndServe(":8080", nil)
}
