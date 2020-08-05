package main

import (
	"net/http"
	"os"

	"github.com/td0m/pub-sub/pkg/auth"
	"github.com/td0m/pub-sub/pkg/handlers"
	"github.com/td0m/pub-sub/pkg/hub"
)

var h *hub.Hub

func init() {
	secret := os.Getenv("JWT_SECRET")
	auth.Init(secret)
}

func main() {
	h = hub.NewHub()
	go h.Run()
	http.HandleFunc("/ws", handlers.NewWsHandler(h))
	http.ListenAndServe(":8080", nil)
}
