package pl

import "mtggameengine/models"

type Bot struct {
	player
}

func (b *Bot) ID() string {
	return ""
}

func (b Bot) IsConnected() bool {
	return false
}

func (b Bot) IsHost() bool {
	return false
}

func (b Bot) IsBot() bool {
	return true
}

func (b Bot) Time() int {
	return 0
}

func (b Bot) Hash() string {
	return ""
}

func NewBot() Player {
	packs := make([]*models.Pack, 0)
	return &Bot{player{
		name:  "bot",
		Packs: packs,
	}}
}
