package pl

import "mtggameengine/models"

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

func (b *Bot) StartPicking(emptyPacks chan<- models.Pack) {
	go func() {
		for pack := range b.Packs {
			if len(pack) <= 0 {
				emptyPacks <- pack
			} else {
				pack.Pick(0)
				b.nextPlayer.AddPack(pack)
			}
		}
	}()

}

func NewBot() Player {
	return &Bot{player{
		name:  "bot",
		Packs: make(chan models.Pack, 1),
	}}
}
