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
	GetPacksCount() int
	AddPack(pack *models.Pack)
	Time() int
	Hash() string
	SetNextPlayer(nextPlayer Player)
	StartPicking(emptyPacks chan<- *models.Pack)
	StopPicking()
}

//Base type for bot and Human players
type player struct {
	name       string
	Packs      chan *models.Pack
	nextPlayer Player
}

func (p *player) SetNextPlayer(nextPlayer Player) {
	p.nextPlayer = nextPlayer
}

func (p *player) GetPacksCount() int {
	return len(p.Packs)
}

func (p *player) Name() string {
	return p.name
}

func (p *player) AddPack(pack *models.Pack) {
	p.Packs <- pack
}

func (p *player) StopPicking() {
	close(p.Packs)
	p.Packs = make(chan *models.Pack)
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
