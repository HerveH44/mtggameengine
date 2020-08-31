package game

import (
	"github.com/google/uuid"
	"log"
	"mtggameengine/game/pl"
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

type Players []pl.Player

func (p *Players) Add(player pl.Player) {
	*p = append(*p, player)
}

func (p *Players) indexOf(player pl.Player) int {
	for i, pl := range *p {
		if player.ID() == pl.ID() {
			return i
		}
	}
	return -1
}

func (p *Players) indexOfID(id string) int {
	for i, pl := range *p {
		if id == pl.ID() {
			return i
		}
	}
	return -1
}

func (p *Players) Remove(index int) {
	arr := *p
	arr[index] = arr[len(arr)-1]
	*p = arr[:len(arr)-1]
}

func (g *defaultGame) SetHost(hostId string) {
	g.HostID = hostId
}

func (g *defaultGame) Join(conn socketio.Conn) {
	// get write lock
	g.lock.Lock()
	defer g.lock.Unlock()
	conn.RemoveAllEvents()

	for _, player := range g.players {
		// link conn to pl
		if player.ID() == conn.ID() {
			log.Println(g.ID(), "pl", conn.ID(), "re-joined the game")
			humanPlayer := player.(*pl.Human)
			humanPlayer.Err("only one window active")
			humanPlayer.Attach(conn)
			g.doJoin(humanPlayer)
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
	conn.OnEvent("leave", g.onConnectionExit)

	player := pl.NewHuman(conn, conn.ID() == g.HostID)
	g.players.Add(player)
	g.doJoin(player)
	if player.IsHost() {
		g.SetHostPermissions(player)
	}
}

func (g *defaultGame) onConnectionExit(c socketio.Conn) {
	// get write lock
	g.lock.Lock()
	defer g.lock.Unlock()

	log.Println(c.ID(), "left the game", g.ID())
	g.Room.Leave(c)
	c.RemoveEvent("leave")

	if g.gameStarted() {
		return
	}

	c.RemoveEvent("start") //a bit out of the blue?
	i := g.players.indexOfID(c.ID())
	g.players.Remove(i)
	g.broadcastPosition()
}

func (g *defaultGame) broadcastPosition() {
	for index, player := range g.players {
		if human, ok := player.(*pl.Human); ok {
			human.Set(PlayerBasicInfo{
				Self: index,
			})
		}
	}
}

func (g *defaultGame) ChangeName(c socketio.Conn, name string) {
	if len(name) > 14 {
		c.SetName(name[:15])
	}
	c.SetName(name)
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
	IsHost bool     `json:"isHost,omitempty"`
	Round  int      `json:"round,omitempty"`
	Self   int      `json:"self"`
	Sets   []string `json:"sets,omitempty"`
	GameId string   `json:"gameId,omitempty"`
}

type BasicInfos struct {
	Type       string   `json:"type"`
	PacksInfos string   `json:"packsInfo"`
	Sets       []string `json:"sets"`
}

func (g *defaultGame) greet(player *pl.Human) {
	info := PlayerBasicInfo{
		IsHost: player.IsHost(),
		Round:  g.round,
		Self:   g.players.indexOf(player),
		Sets:   g.Sets,
		GameId: g.ID(),
	}
	player.Set(info)

	player.Emit("gameInfos", BasicInfos{
		Type:       g.Type,
		PacksInfos: g.PacksInfo,
		Sets:       g.Sets,
	})
}

type PlayerSpecificInfo struct {
	Name        string `json:"name"`
	Time        int    `json:"time"`
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
			Time:        p.Time(),
			Packs:       len(*p.Packs()),
			IsBot:       p.IsBot(),
			IsConnected: p.IsConnected(),
			Hash:        p.Hash(),
		}
		playersState = append(playersState, ps)
	}

	g.Room.Broadcast("set", StateInfo{
		Players:   &playersState,
		GameSeats: g.Seats,
	})
}

func (g *defaultGame) doJoin(player *pl.Human) {
	g.greet(player)
	g.Room.Join(player)
	player.OnEvent("name", func(c socketio.Conn, name string) {
		g.meta()
	})
	g.meta()
}

type StartRequest struct {
	AddBots        bool   `json:"addBots"`
	UseTimer       bool   `json:"useTimer"`
	TimerLength    string `json:"timerLength"`
	ShufflePlayers bool   `json:"shufflePlayers"`
}

func (g *defaultGame) SetHostPermissions(player *pl.Human) {
	player.OnEvent("start", func(c socketio.Conn, msg StartRequest) {
		log.Println("start: ", msg)
	})
	player.OnEvent("kick", func(c socketio.Conn, index int) {
		if index < 0 || index >= len(g.players) {
			c.Err("player index is out of players range")
		}

		player := g.players[index]

		if player.IsBot() {
			return
		}

		human := player.(*pl.Human)
		log.Println(player.Name(), "is being kicked out from the game")
		if g.gameStarted() {
			human.Kick()
		} else {
			g.onConnectionExit(human)
		}

		human.Err("you were kicked")
		g.meta()
	})
	player.OnEvent("swap", func(c socketio.Conn, msg [2]int) {
		// get write lock
		g.lock.Lock()
		defer g.lock.Unlock()

		l := len(g.players)
		i, j := msg[0], msg[1]

		if j < 0 || j >= l {
			return
		}

		g.players[i], g.players[j] = g.players[j], g.players[i]
		g.broadcastPosition()
		g.meta()
	})
}

func CreateGame(gameRequest models.CreateGameRequest, cubeList []string) Game {
	return &defaultGame{
		Room:       newRoom(gameRequest.IsPrivate),
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
