package pl

import (
	"mtggameengine/models"
	socketio "mtggameengine/socket"
)

type Human struct {
	socketio.Conn
	*player
	isConnected bool
	isHost      bool
	UseTimer    bool
	TimerLength string
	PickNumber  int
}

func (h *Human) Name() string {
	return h.Conn.Name()
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

func (h *Human) Kick() {
	// Should turn to a bot
	h.isConnected = false
}

func (h *Human) StartPicking(emptyPacks chan<- *models.Pack) {
	go func() {
		for pack := range h.Packs {
			if len(*pack) <= 0 {
				emptyPacks <- pack
			} else {
				h.Emit("pack", pack)
				h.PickNumber++
				h.Emit("pickNumber", h.PickNumber)
			}
		}
	}()
}

func NewHuman(conn socketio.Conn, isHost bool) *Human {
	return &Human{
		Conn: conn,
		player: &player{
			name:  conn.Name(),
			Packs: make(chan *models.Pack, 1),
		},
		isConnected: true,
		isHost:      isHost,
	}
}
