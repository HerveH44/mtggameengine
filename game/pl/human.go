package pl

import (
	"mtggameengine/hash"
	"mtggameengine/models"
	socketio "mtggameengine/socket"
	"sort"
	"sync"
)

type Hash struct {
	Cockatrice string `json:"cock"`
	MWS        string `json:"mws"`
}

type Human struct {
	socketio.Conn
	*player
	isConnected bool
	isHost      bool
	UseTimer    bool
	TimerLength string
	PickNumber  int
	pickLock    sync.Mutex
	pack        models.Pack
	pool        models.Cards
	hash        Hash
}

func NewHuman(conn socketio.Conn, isHost bool) *Human {
	h := &Human{
		Conn: conn,
		player: &player{
			name:        conn.Name(),
			Packs:       make(chan models.Pack, 100),
			stopPicking: make(chan bool),
		},
		isConnected: true,
		isHost:      isHost,
		pool:        make(models.Cards, 0),
	}
	h.OnEvent("pick", h.onPick)
	h.OnEvent("hash", h.onHash)
	h.onPack(h.handlePack)
	return h
}

func (h *Human) Name() string {
	return h.Conn.Name()
}

func (h *Human) Time() int {
	return 0
}

func (h *Human) Hash() Hash {
	return h.hash
}

func (h *Human) IsBot() bool {
	return !h.isConnected
}

func (h *Human) IsHost() bool {
	return h.isHost
}

func (h *Human) Attach(conn socketio.Conn) {
	if conn != h.Conn {
		h.Conn.Close()
	}
	h.Conn = conn
	h.isConnected = true
	h.sendPool()

	h.OnEvent("hash", h.onHash)

	// Draft
	h.OnEvent("handlePack", h.onPick)
	if h.pack != nil {
		h.Emit("pack", h.pack)
	}
}

func countByName(pool models.Cards) map[string]int {
	ret := make(map[string]int)
	for _, card := range pool {
		if _, ok := ret[card.Name]; ok {
			ret[card.Name]++
		} else {
			ret[card.Name] = 1
		}
	}
	return ret
}

var BASIC = []string{"Forest", "Island", "Mountain", "Plains", "Swamp"}

func Include(arr []string, val string) bool {
	sort.Strings(arr)
	i := sort.SearchStrings(arr, val)
	if i >= len(arr) || arr[i] != val {
		return false
	}
	return true
}

func (h *Human) onHash(_ socketio.Conn, deck models.HashEvent) {
	if h.checkDeck(deck) {
		if calculatedHash, err := calcHash(deck); err == nil {
			h.hash = calculatedHash
		}
	}
}

func calcHash(deck models.HashEvent) (h Hash, err error) {
	if h.Cockatrice, err = hash.MakeCockatriceHash(deck); err != nil {
		return
	}
	if h.MWS, err = hash.MakeMWSHash(deck); err != nil {
		return
	}
	return
}

func (h *Human) checkDeck(deck models.HashEvent) bool {
	poolByName := countByName(h.pool)
	cards := deck.Main
	for cardName, num := range cards {
		if Include(BASIC, cardName) {
			continue
		}

		if _, ok := poolByName[cardName]; !ok {
			return false
		}
		poolByName[cardName] -= num
		if poolByName[cardName] < 0 {
			return false
		}
	}
	cards2 := deck.Side
	for cardName, num := range cards2 {
		if Include(BASIC, cardName) {
			continue
		}

		if _, ok := poolByName[cardName]; !ok {
			return false
		}
		poolByName[cardName] -= num
		if poolByName[cardName] < 0 {
			return false
		}
	}

	return true
}

func (h *Human) IsConnected() bool {
	return h.isConnected
}

func (h *Human) Kick() {
	// Should turn to a bot
	h.isConnected = false
}

func (h *Human) onPick(_ socketio.Conn, index int) {
	if h.pack == nil || index >= len(h.pack) {
		return
	}
	defer h.pickLock.Unlock()
	card := h.pack[index]
	h.Emit("add", card)
	h.pool.Push(card)
	h.pack.Pick(index)
	h.pass(h.pack)
}

func (h *Human) handlePack(pack models.Pack) {
	h.pickLock.Lock()
	h.sendPack(pack)
}

func (h *Human) sendPack(pack models.Pack) {
	h.pack = pack
	h.Emit("pack", h.pack)
	h.PickNumber++
	h.Emit("pickNumber", h.PickNumber)
}

func (h *Human) AddPool(pool models.Pool) {
	for _, p := range pool {
		h.pool = append(h.pool, p...)
	}
	h.sendPool()
}

func (h *Human) sendPool() {
	h.Emit("pool", h.pool)
}
