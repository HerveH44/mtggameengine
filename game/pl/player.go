package pl

import (
	"mtggameengine/models"
)

type Player interface {
	ID() string
	Name() string
	IsConnected() bool
	IsHost() bool
	IsBot() bool
	GetPacks() []*models.Pack
	Time() int
	Hash() string
}

//Base type for bot and Human players
type player struct {
	name  string
	Packs []*models.Pack
}

func (p *player) GetPacks() []*models.Pack {
	return p.Packs
}

func (p *player) Name() string {
	return p.name
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
