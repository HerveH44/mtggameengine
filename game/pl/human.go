package pl

import (
	"mtggameengine/models"
	socketio "mtggameengine/socket"
	"sync"
)

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
	h.onPack(h.handlePack)
	return h
}

func (h *Human) Name() string {
	return h.Conn.Name()
}

func (h *Human) Time() int {
	return 0
}

func (h *Human) Hash() string {
	return ""
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

	// Draft
	h.OnEvent("handlePack", h.onPick)
	if h.pack != nil {
		h.Emit("pack", h.pack)
	}
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
