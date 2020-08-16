package services

import (
	"fmt"
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

type GameService interface {
	CreateGame(game models.CreateGameRequest) (*Game, error)
}

type defaultGameService struct {
	poolService PoolService
}

func NewDefaultGameService(service PoolService) GameService {
	return &defaultGameService{poolService: service}
}

func (s *defaultGameService) CreateGame(game models.CreateGameRequest) (*Game, error) {

	if game.Type == "cube draft" || game.Type == "cube sealed" {
		missingCards, err := s.poolService.CheckCubeList(strings.Split(game.Cube.List, "\n"))
		if err != nil {
			fmt.Println(err.Error())
			return nil, fmt.Errorf("unexpected server error")
		}

		if len(missingCards) > 0 {
			return nil, fmt.Errorf("invalid cards: %s", strings.Join(missingCards[:10], "; "))
		}
	}

	return &Game{
		ID:         uuid.New().String(),
		Type:       game.Type,
		Title:      game.Title,
		Seats:      game.Seats,
		IsPrivate:  game.IsPrivate,
		Sets:       game.Sets,
		ModernOnly: game.ModernOnly,
		TotalChaos: game.TotalChaos,
		CubeList:   strings.Split(game.Cube.List, "\n"),
	}, nil
}
