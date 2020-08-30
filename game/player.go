package game

import socketio "mtggameengine/socket"

type Player interface {
	socketio.Conn
	Attach(conn socketio.Conn)
	IsConnected() bool
	IsHost() bool
	IsBot() bool
}

type human struct {
	socketio.Conn
	isConnected bool
	isHost      bool
}

func (h *human) IsBot() bool {
	return !h.isConnected
}

func (h *human) IsHost() bool {
	return h.isHost
}

func (h *human) Attach(conn socketio.Conn) {
	h.Conn = conn
	h.isConnected = true
}

func (h *human) IsConnected() bool {
	return h.isConnected
}

func newHuman(conn socketio.Conn, isHost bool) Player {

	return &human{
		Conn:        conn,
		isConnected: true,
		isHost:      isHost,
	}
}
