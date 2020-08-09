package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/td0m/pub-sub/pkg/auth"
	"github.com/td0m/pub-sub/pkg/hub"
)

func NewHTTPHandler(h *hub.Hub) http.HandlerFunc {
	return auth.WithClaims(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value("claims").(*auth.Claims)
		log.Println(claims)
		var msg hub.Message
		json.NewDecoder(r.Body).Decode(&msg)
		h.Emit(msg)
	})
}
