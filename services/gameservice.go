package services

import (
	"fmt"
	"github.com/google/uuid"
	"mtggameengine/models"
	socketio "mtggameengine/socket"
	"strings"
)

type Game struct {
	ID        string
	Type      string
	Title     string
	Seats     int
	IsPrivate bool

	//Players
	HostID string

	// Regular
	Sets []string

	// Chaos
	ModernOnly bool
	TotalChaos bool

	//Cube
	CubeList []string
}

type GameService interface {
	CreateGame(game models.CreateGameRequest, conn socketio.Conn) (*Game, error)
}

type defaultGameService struct {
	poolService PoolService
}

func NewDefaultGameService(service PoolService) GameService {
	return &defaultGameService{poolService: service}
}

func (s *defaultGameService) CreateGame(game models.CreateGameRequest, conn socketio.Conn) (*Game, error) {

	cubeList := strings.Split(game.Cube.List, "\n")

	// Validate cube request
	if game.Type == "cube draft" || game.Type == "cube sealed" {

		if game.Type == "cube draft" && game.Cube.Cards*game.Cube.Packs*game.Seats < len(cubeList) {
			return nil, fmt.Errorf("not enough cards")
		}

		if game.Type == "cube sealed" && game.Cube.CubePoolSize*game.Seats < len(cubeList) {
			return nil, fmt.Errorf("not enough cards")
		}

		missingCards, err := s.poolService.CheckCubeList(cubeList)
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
		HostID:     conn.ID(),
		Type:       game.Type,
		Title:      game.Title,
		Seats:      game.Seats,
		IsPrivate:  game.IsPrivate,
		Sets:       game.Sets,
		ModernOnly: game.ModernOnly,
		TotalChaos: game.TotalChaos,
		CubeList:   cubeList,
	}, nil
}
