package game

import (
	socketio "mtggameengine/socket"
	"sync"
	"time"
)

type Room interface {
	Join(conn socketio.Conn)
	Leave(conn socketio.Conn)
	Broadcast(event string, v interface{})
}

type defaultRoom struct {
	messages    []Message
	connections Connections
	isPrivate   bool
	timeCreated time.Time
	lock        sync.RWMutex
}

func newRoom(isPrivate bool) Room {
	return &defaultRoom{
		messages:    make([]Message, 0),
		connections: make(Connections, 0),
		isPrivate:   isPrivate,
		timeCreated: time.Now(),
		lock:        sync.RWMutex{},
	}
}

type Connections []socketio.Conn

func (c *Connections) remove(conn socketio.Conn) {
	for i, co := range *c {
		if conn.ID() == co.ID() {
			arr := *c
			arr[i] = arr[len(arr)-1]
			*c = arr[:len(arr)-1]
			return
		}
	}
}

func (c *Connections) broadcast(event string, v interface{}) {
	for _, conn := range *c {
		conn.Emit(event, v)
	}
}

type Message struct {
	name string
	Text string
	time time.Time
}

func (d *defaultRoom) Join(conn socketio.Conn) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.connections = append(d.connections, conn)

	// Handle say
	conn.OnEvent("say", func(c socketio.Conn, msg string) {
		message := Message{
			name: conn.Name(),
			Text: msg,
			time: time.Now(),
		}

		d.lock.Lock()
		defer d.lock.Unlock()

		d.messages = append(d.messages, message)
		for _, conn := range d.connections {
			conn.Emit("hear", message)
		}
	})

	//Handle Name
	conn.OnEvent("name", func(c socketio.Conn, name string) {
		c.SetName(name[:15])
	})

	// Send all messages
	conn.Emit("chat", d.messages)
}

func (d *defaultRoom) Leave(conn socketio.Conn) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.connections.remove(conn)
	conn.RemoveEvent("say")
}

func (d *defaultRoom) Broadcast(event string, v interface{}) {
	d.connections.broadcast(event, v)
}
