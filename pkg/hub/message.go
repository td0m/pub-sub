package hub

type Message struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}
