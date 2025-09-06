package websocket

// IncomingMessage represents a message received from a client.
type IncomingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// OutgoingMessage represents a message sent to a client.
type OutgoingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
