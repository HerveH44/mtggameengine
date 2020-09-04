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

func (b *Bot) Hash() (h Hash) {
	return h
}

func (b *Bot) pick(pack models.Pack) {
	rand.Seed(time.Now().UnixNano())
	pack.Pick(rand.Intn(len(pack)))
	b.pass(pack)
}

func NewBot() Player {
	bot := Bot{player{
		name:        "bot",
		Packs:       make(chan models.Pack, 1),
		stopPicking: make(chan bool),
	}}
	bot.onPack(bot.pick)
	return &bot
}
