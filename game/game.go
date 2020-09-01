package game

import (
	"github.com/google/uuid"
	"log"
	"math/rand"
	"mtggameengine/game/pl"
	"mtggameengine/models"
	socketio "mtggameengine/socket"
	"strings"
	"sync"
	"time"
)

type Game interface {
	ID() string
	SetHost(hostId string)
	Join(conn socketio.Conn)
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
		players:    make(pl.Players, 0),
	}
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
	players   pl.Players
	round     int
	PacksInfo string
}

func (g *defaultGame) ID() string {
	return g.id
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
	conn.OnEvent("leave", g.onSocketLeave)

	player := pl.NewHuman(conn, conn.ID() == g.HostID)
	g.players.Add(player)
	g.doJoin(player)
	if player.IsHost() {
		g.setHostPermissions(player)
	}
}

func (g *defaultGame) onSocketLeave(c socketio.Conn) {
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
	i := g.players.IndexOfID(c.ID())
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

func (g *defaultGame) gameStarted() bool {
	return g.round != 0
}

func (g *defaultGame) gameFinished() bool {
	return g.round == -1
}

func (g *defaultGame) greet(player *pl.Human) {
	info := PlayerBasicInfo{
		IsHost: player.IsHost(),
		Round:  g.round,
		Self:   g.players.IndexOf(player),
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

func (g *defaultGame) meta() {
	playersState := make([]PlayerSpecificInfo, 0)
	for _, p := range g.players {
		ps := PlayerSpecificInfo{
			Name:        p.Name(),
			Time:        p.Time(),
			Packs:       len(p.GetPacks()),
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

func (g *defaultGame) setHostPermissions(player *pl.Human) {
	player.OnEvent("start", g.start)
	player.OnEvent("kick", g.kick)
	player.OnEvent("swap", g.swap)
}

func (g *defaultGame) start(c socketio.Conn, startRequest StartRequest) {
	if startRequest.AddBots {
		g.addBots()
	}

	if startRequest.ShufflePlayers {
		g.shufflePlayers()
	}

	g.broadcastPosition()
	g.meta()
	g.round++
}

func (g *defaultGame) addBots() {
	// get write lock
	g.lock.Lock()
	defer g.lock.Unlock()

	for i := len(g.players); i < g.Seats; i++ {
		g.players.Add(pl.NewBot())
	}
}

func (g *defaultGame) shufflePlayers() {
	// get write lock
	g.lock.Lock()
	defer g.lock.Unlock()

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(g.players), func(i, j int) { g.players[i], g.players[j] = g.players[j], g.players[i] })
}

func (g *defaultGame) kick(c socketio.Conn, index int) {
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
		g.onSocketLeave(human)
	}

	human.Err("you were kicked")
	g.meta()
}

func (g *defaultGame) swap(_ socketio.Conn, msg [2]int) {
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
}
