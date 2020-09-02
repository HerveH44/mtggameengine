package game

import (
	"github.com/google/uuid"
	"log"
	"math/rand"
	"mtggameengine/game/pl"
	"mtggameengine/models"
	"mtggameengine/services/pool"
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

func CreateGame(gameRequest models.CreateGameRequest, cubeList []string, service pool.PoolService) Game {
	return &defaultGame{
		Room:        newRoom(gameRequest.IsPrivate),
		id:          uuid.New().String(),
		Type:        gameRequest.Type,
		Title:       gameRequest.Title,
		Seats:       gameRequest.Seats,
		IsPrivate:   gameRequest.IsPrivate,
		Sets:        gameRequest.Sets,
		ModernOnly:  gameRequest.ModernOnly,
		TotalChaos:  gameRequest.TotalChaos,
		CubeList:    cubeList,
		PacksInfo:   strings.Join(gameRequest.Sets, " / "),
		players:     make(pl.Players, 0),
		poolService: service,
		rounds:      calcRounds(gameRequest),
		delta:       -1,
	}
}

func calcRounds(request models.CreateGameRequest) int {
	switch request.Type {
	case "draft":
		fallthrough
	case "decadent draft":
		fallthrough
	case "sealed":
		return len(request.Sets)

	case "cube draft":
		fallthrough
	case "cube sealed":
		return request.Cube.Packs

	case "chaos draft":
		fallthrough
	case "chaos sealed":
		return request.ChaosPackNumber
	}
	return 0
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
	PacksInfo string

	poolService pool.PoolService
	pool        models.Pool

	// TODO: ref this. don't think we need it as is
	round int

	// Things only for draft mode
	useTimer       bool
	timerLength    string
	rounds         int
	packCount      int
	delta          int //to check the order of pack passing
	emptyPacksChan chan *models.Pack
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
			Packs:       p.GetPacksCount(),
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
	g.useTimer = startRequest.UseTimer
	g.timerLength = startRequest.TimerLength

	if startRequest.AddBots {
		g.addBots()
	}

	if startRequest.ShufflePlayers {
		g.shufflePlayers()
	}

	g.createPool()

	//Handle sealed or draft or other?
	if g.Type == "draft" {
		g.handleDraft()
	}

	g.broadcastPosition()
	g.meta()
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

func (g *defaultGame) createPool() {
	switch g.Type {
	case "draft":
		regularPool, err := g.poolService.MakeRegularPool(models.RegularRequest{
			Players: len(g.players),
			Sets:    g.Sets,
		})
		if err != nil {
			log.Println("Could not fetch regularPool", err)
		}
		g.pool = regularPool
	}
}

func (g *defaultGame) handleDraft() {
	for _, p := range g.players {
		if !p.IsBot() {
			human := p.(*pl.Human)
			human.UseTimer = g.useTimer
			human.TimerLength = g.timerLength
		}
	}
	g.meta()
	g.startRound()
}

func (g *defaultGame) startRound() {
	if g.round != 0 {

	}

	g.round++
	if g.round == g.rounds {
		g.endGame()
		return
	}

	log.Println(g.id, "new round started")

	g.packCount = len(g.players)
	g.delta *= -1

	// Give packs to every player
	for i, player := range g.players {
		player.SetNextPlayer(g.getNextPlayer(i))
		pack := g.pool.Shift()
		player.AddPack(&pack)
		if !player.IsBot() {
			human := player.(*pl.Human)
			human.PickNumber = 0
			human.Set(PlayerBasicInfo{
				PackSize: len(pack),
				Round:    g.round,
			})
		}
	}

	g.emptyPacksChan = make(chan *models.Pack, len(g.players))
	for _, player := range g.players {
		player.StartPicking(g.emptyPacksChan)
	}

	go func() {
		count := 0
		for range g.emptyPacksChan {
			count++
			// emptied all the packs of the round
			if count == cap(g.emptyPacksChan) {
				for _, player := range g.players {
					player.StopPicking()
				}
			}
		}
	}()

	g.meta()
}

func (g *defaultGame) getNextPlayer(playerIndex int) pl.Player {
	playersLen := len(g.players)
	nextIndex := playerIndex + g.delta
	index := (nextIndex%playersLen + playersLen) % playersLen
	return g.players[index]
}

func (g *defaultGame) endGame() {
	log.Println(g.id, "game ended")

	g.round = -1
	g.meta()
}
