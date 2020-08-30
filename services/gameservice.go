package services

import (
	"fmt"
	"mtggameengine/game"
	"mtggameengine/models"
	socketio "mtggameengine/socket"
	"strings"
)

type GameService interface {
	CreateGame(game models.CreateGameRequest, conn socketio.Conn) (game.Game, error)
	Join(gameId string, conn socketio.Conn)
}

type defaultGameService struct {
	poolService PoolService
	games       map[string]game.Game
}

func (s *defaultGameService) Join(gameId string, conn socketio.Conn) {
	// check if game exists
	if _, ok := s.games[gameId]; !ok {
		conn.Emit("error", fmt.Sprint("game", gameId, "does not exist"))
		return
	}

	// add the connection to the rooms connection map
	s.games[gameId].Join(conn)
}

func NewDefaultGameService(service PoolService) GameService {
	return &defaultGameService{
		poolService: service,
		games:       make(map[string]game.Game),
	}
}

func (s *defaultGameService) CreateGame(gameRequest models.CreateGameRequest, conn socketio.Conn) (game.Game, error) {

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

	newGame := game.CreateGame(gameRequest, cubeList)
	newGame.SetHost(conn.ID())
	s.addGame(newGame)
	return newGame, nil
}

func (s *defaultGameService) addGame(g game.Game) {
	s.games[g.ID()] = g
}
