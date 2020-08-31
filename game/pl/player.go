package pl

import (
	"mtggameengine/models"
	socketio "mtggameengine/socket"
)

type Player interface {
	ID() string
	Name() string
	IsConnected() bool
	IsHost() bool
	IsBot() bool
	Packs() *[]models.Pack
	Time() int
	Hash() string
}

//Base type for bot and Human players
type player struct {
	Name  string
	Packs *[]models.Pack
}

func NewHuman(conn socketio.Conn, isHost bool) *Human {

	packs := make([]models.Pack, 0)
	return &Human{
		Conn: conn,
		player: player{
			Name:  conn.Name(),
			Packs: &packs,
		},
		isConnected: true,
		isHost:      isHost,
	}
}
