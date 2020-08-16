package services

import (
	"github.com/google/uuid"
	"mtggameengine/models"
	"strings"
)

type Game struct {
	ID        string
	Type      string
	Title     string
	Seats     int
	IsPrivate bool

	// Regular
	Sets []string

	// Chaos
	ModernOnly bool
	TotalChaos bool

	//Cube
	CubeList []string
}

func CreateGame(game models.CreateGame) Game {
	return Game{
		ID:         uuid.New().String(),
		Type:       game.Type,
		Title:      game.Title,
		Seats:      game.Seats,
		IsPrivate:  game.IsPrivate,
		Sets:       game.Sets,
		ModernOnly: game.ModernOnly,
		TotalChaos: game.TotalChaos,
		CubeList:   strings.Split(game.CubeList, "\n"),
	}
}
