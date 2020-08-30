package game

import (
	socketio "mtggameengine/socket"
	"sync"
	"time"
)

type Room interface {
	Join(conn socketio.Conn)
}

type defaultRoom struct {
	messages    []Message
	connections Connections
	isPrivate   bool
	timeCreated time.Time
	lock        sync.RWMutex
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
	conn.OnEvent("say", func(msg string) {
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
	conn.OnEvent("name", func(name string) {
		conn.SetName(name[:15])
	})

	//Handle exit
	conn.OnEvent("exit", func() {
		d.lock.Lock()
		defer d.lock.Unlock()
		d.connections.remove(conn)

		conn.RemoveEvent("say")
		conn.RemoveEvent("exit")
	})

	// Send all messages
	conn.Emit("chat", d.messages)
}
