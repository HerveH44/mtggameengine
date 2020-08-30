package game

import (
	"github.com/google/uuid"
	"log"
	"mtggameengine/models"
	socketio "mtggameengine/socket"
	"sync"
)

type Game interface {
	ID() string
	SetHost(hostId string)
	Join(conn socketio.Conn)
}

type defaultGame struct {
	Room
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

	lock    sync.RWMutex // access lock
	players map[string]Player
	round   int
}

func (g *defaultGame) SetHost(hostId string) {
	g.HostID = hostId
}

func (g *defaultGame) Join(conn socketio.Conn) {
	// get write lock
	g.lock.Lock()
	defer g.lock.Unlock()

	playerId := conn.ID()

	player, ok := g.players[playerId]
	if ok {
		log.Println(g.ID(), "player", playerId, "re-joined the game")
		// Greet
		// Broadcast infos
		// link conn to player?
		return
	}

	if g.gameStarted() {
		conn.Err("game already started")
		return
	}

	g.Room.Join(conn)

	//Pick Delegate?
	//Create HumanPlayer
	//If It's the host give him extra events
	//add player to list of players
	// greet
	// broadcast
}

func (g *defaultGame) ID() string {
	return g.id
}

func (g *defaultGame) gameStarted() bool {
	return g.round != 0
}

func (g *defaultGame) gameFinished() bool {
	return g.round == -1
}

func CreateGame(gameRequest models.CreateGameRequest, cubeList []string) Game {
	return &defaultGame{
		Room:       &defaultRoom{},
		id:         uuid.New().String(),
		Type:       gameRequest.Type,
		Title:      gameRequest.Title,
		Seats:      gameRequest.Seats,
		IsPrivate:  gameRequest.IsPrivate,
		Sets:       gameRequest.Sets,
		ModernOnly: gameRequest.ModernOnly,
		TotalChaos: gameRequest.TotalChaos,
		CubeList:   cubeList,
	}
}
