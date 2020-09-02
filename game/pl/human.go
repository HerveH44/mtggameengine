package pl

import (
	"log"
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
	pack        *models.Pack
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
	h.Conn = conn
	h.isConnected = true
	if h.pack != nil {
		h.sendPack()
	}
}

func (h *Human) IsConnected() bool {
	return h.isConnected
}

func (h *Human) Kick() {
	// Should turn to a bot
	h.isConnected = false
}

func (h *Human) StartPicking(emptyPacks chan<- *models.Pack) {
	h.OnEvent("pick", func(conn socketio.Conn, index int) {
		if h.pack == nil || index >= len(*h.pack) {
			return
		}
		card := (*h.pack)[index]
		log.Println("pick", index)
		log.Println("card picked", card.Name)
		packToPass := (*h.pack).Pick(index)
		h.nextPlayer.AddPack(&packToPass)
		h.pickLock.Unlock()
	})

	go func() {
		for pack := range h.Packs {
			if len(*pack) <= 0 {
				emptyPacks <- pack
			} else {
				h.pickLock.Lock()
				h.pack = pack
				h.sendPack()
			}
		}
	}()
}

func (h *Human) sendPack() {
	h.Emit("pack", h.pack)
	h.PickNumber++
	h.Emit("pickNumber", h.PickNumber)
}

func NewHuman(conn socketio.Conn, isHost bool) *Human {
	return &Human{
		Conn: conn,
		player: &player{
			name:  conn.Name(),
			Packs: make(chan *models.Pack, 100),
		},
		isConnected: true,
		isHost:      isHost,
	}
}
