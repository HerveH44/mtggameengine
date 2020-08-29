package services

import (
	"fmt"
	"github.com/google/uuid"
	"mtggameengine/models"
	socketio "mtggameengine/socket"
	"strings"
)

type Game interface {
	ID() string
}

type defaultGame struct {
	id        string
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

func (g *defaultGame) ID() string {
	return g.id
}

type GameService interface {
	CreateGame(game models.CreateGameRequest, conn socketio.Conn) (Game, error)
}

type defaultGameService struct {
	poolService PoolService
	games       map[string]Game
}

func NewDefaultGameService(service PoolService) GameService {
	return &defaultGameService{
		poolService: service,
		games:       make(map[string]Game),
	}
}

func (s *defaultGameService) CreateGame(gameRequest models.CreateGameRequest, conn socketio.Conn) (Game, error) {

	cubeList := strings.Split(gameRequest.Cube.List, "\n")

	// Validate cube request
	if gameRequest.Type == "cube draft" || gameRequest.Type == "cube sealed" {

		if gameRequest.Type == "cube draft" && gameRequest.Cube.Cards*gameRequest.Cube.Packs*gameRequest.Seats < len(cubeList) {
			return nil, fmt.Errorf("not enough cards")
		}

		if gameRequest.Type == "cube sealed" && gameRequest.Cube.CubePoolSize*gameRequest.Seats < len(cubeList) {
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

	game := defaultGame{
		id:         uuid.New().String(),
		HostID:     conn.ID(),
		Type:       gameRequest.Type,
		Title:      gameRequest.Title,
		Seats:      gameRequest.Seats,
		IsPrivate:  gameRequest.IsPrivate,
		Sets:       gameRequest.Sets,
		ModernOnly: gameRequest.ModernOnly,
		TotalChaos: gameRequest.TotalChaos,
		CubeList:   cubeList,
	}
	s.addGame(&game)
	return &game, nil
}

func (s *defaultGameService) addGame(g Game) {
	s.games[g.ID()] = g
}
