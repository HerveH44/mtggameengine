package pl

import (
	"mtggameengine/models"
	socketio "mtggameengine/socket"
)

type Human struct {
	socketio.Conn
	player
	isConnected bool
	isHost      bool
}

func (h *Human) Name() string {
	return h.Conn.Name()
}

func (h *Human) Packs() *[]models.Pack {
	return h.player.Packs
}

func (h *Human) Time() int {
	return 0
}

func (h *Human) Hash() string {
	return ""
}

func (h *Human) IsBot() bool {
	return !h.isConnected
}

func (h *Human) IsHost() bool {
	return h.isHost
}

func (h *Human) Attach(conn socketio.Conn) {
	h.Conn = conn
	h.isConnected = true
}

func (h *Human) IsConnected() bool {
	return h.isConnected
}
