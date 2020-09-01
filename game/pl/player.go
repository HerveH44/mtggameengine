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

type Players []Player

func (p *Players) Add(player Player) {
	*p = append(*p, player)
}

func (p *Players) IndexOf(player Player) int {
	for i, pl := range *p {
		if player.ID() == pl.ID() {
			return i
		}
	}
	return -1
}

func (p *Players) IndexOfID(id string) int {
	for i, pl := range *p {
		if id == pl.ID() {
			return i
		}
	}
	return -1
}

func (p *Players) Remove(index int) {
	arr := *p
	arr[index] = arr[len(arr)-1]
	*p = arr[:len(arr)-1]
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
