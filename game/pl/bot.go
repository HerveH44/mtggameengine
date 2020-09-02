package pl

import (
	"math/rand"
	"mtggameengine/models"
	"time"
)

type Bot struct {
	player
}

func (b *Bot) ID() string {
	return ""
}

func (b *Bot) IsConnected() bool {
	return false
}

func (b *Bot) IsHost() bool {
	return false
}

func (b *Bot) IsBot() bool {
	return true
}

func (b *Bot) Time() int {
	return 0
}

func (b *Bot) Hash() string {
	return ""
}

func (b *Bot) StartPicking() {
	go func() {
		for pack := range b.Packs {
			if len(pack) <= 0 {
				continue
			} else {
				rand.Seed(time.Now().UnixNano())
				pack.Pick(rand.Intn(len(pack)))
				b.pass(pack)
			}
		}
	}()

}

func (b *Bot) StopPicking() {
	close(b.Packs)
	b.Packs = make(chan models.Pack, 1)
}

func NewBot() Player {
	return &Bot{player{
		name:  "bot",
		Packs: make(chan models.Pack, 1),
	}}
}
