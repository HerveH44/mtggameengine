package game

import (
	"github.com/google/uuid"
	"log"
	"mtggameengine/models"
	socketio "mtggameengine/socket"
	"strings"
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

	lock      sync.RWMutex // access lock
	players   Players
	round     int
	PacksInfo string
}

type Players []Player

func (p *Players) Add(player Player) {
	*p = append(*p, player)
}

func (p *Players) indexOf(player Player) int {
	for i, pl := range *p {
		if player == pl {
			return i
		}
	}
	return -1
}

func (g *defaultGame) SetHost(hostId string) {
	g.HostID = hostId
}

func (g *defaultGame) Join(conn socketio.Conn) {
	// get write lock
	g.lock.Lock()
	defer g.lock.Unlock()

	for _, player := range g.players {
		// link conn to player
		if player.ID() == conn.ID() {
			log.Println(g.ID(), "player", conn.ID(), "re-joined the game")
			player.Err("only one window active")
			player.Attach(conn)
			g.greet(player)
			g.Room.Join(conn)
			g.meta()
			return
		}
	}

	if g.gameStarted() {
		conn.Err("game already started")
		return
	}

	if g.Seats == len(g.players) {
		conn.Err("game is already full")
		return
	}

	g.Room.Join(conn)

	//Pick Delegate?

	//If It's the host give him extra events
	player := newHuman(conn, conn.ID() == g.HostID)
	g.players.Add(player)
	g.greet(player)
	// broadcast
	g.meta()
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

type PlayerBasicInfo struct {
	IsHost bool     `json:"isHost"`
	Round  int      `json:"round"`
	Self   int      `json:"self"`
	Sets   []string `json:"sets"`
	GameId string   `json:"gameId"`
}

type BasicInfos struct {
	Type       string   `json:"type"`
	PacksInfos string   `json:"packsInfo"`
	Sets       []string `json:"sets"`
}

func (g *defaultGame) greet(player Player) {
	player.Set(PlayerBasicInfo{
		IsHost: player.IsHost(),
		Round:  g.round,
		Self:   g.players.indexOf(player),
		Sets:   g.Sets,
		GameId: g.ID(),
	})

	player.Emit("gameInfos", BasicInfos{
		Type:       g.Type,
		PacksInfos: g.PacksInfo,
		Sets:       g.Sets,
	})
}

type PlayerSpecificInfo struct {
	Name        string `json:"name"`
	Time        string `json:"time"`
	Packs       int    `json:"packs"`
	IsBot       bool   `json:"isBot"`
	IsConnected bool   `json:"isConnected"`
	Hash        string `json:"hash"`
}

type StateInfo struct {
	Players   *[]PlayerSpecificInfo `json:"players"`
	GameSeats int                   `json:"gameSeats"`
}

func (g *defaultGame) meta() {
	playersState := make([]PlayerSpecificInfo, 0)
	for _, p := range g.players {
		ps := PlayerSpecificInfo{
			Name:        p.Name(),
			Time:        "0",
			Packs:       0,
			IsBot:       p.IsBot(),
			IsConnected: p.IsConnected(),
			Hash:        "",
		}
		playersState = append(playersState, ps)
	}

	g.Room.Broadcast("set", StateInfo{
		Players:   &playersState,
		GameSeats: g.Seats,
	})
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
		PacksInfo:  strings.Join(gameRequest.Sets, " / "),
		players:    make(Players, 0),
	}
}
