package pl

import (
	"mtggameengine/models"
)

type Bot struct {
	player
}

func (b *Bot) ID() string {
	return ""
}

func (b *Bot) Name() string {
	panic("bot")
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

func (b Bot) Packs() *[]models.Pack {
	panic("implement me")
}

func (b Bot) Time() int {
	panic("implement me")
}

func (b Bot) Hash() string {
	panic("implement me")
}
