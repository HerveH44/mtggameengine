package socketio

import (
	"reflect"
)

type namespaceHandler struct {
	onConnect    func(c Conn) error
	onDisconnect func(c Conn, msg string)
	onError      func(c Conn, err error)
	events       map[string]*funcHandler
	broadcast    Broadcast
}

func newHandler() *namespaceHandler {
	return &namespaceHandler{
		events:    make(map[string]*funcHandler),
		broadcast: NewBroadcast(),
	}
}

func (h *namespaceHandler) OnConnect(f func(Conn) error) {
	h.onConnect = f
}

func (h *namespaceHandler) OnDisconnect(f func(Conn, string)) {
	h.onDisconnect = f
}

func (h *namespaceHandler) OnError(f func(Conn, error)) {
	h.onError = f
}

func (h *namespaceHandler) OnEvent(event string, f interface{}) {
	h.events[event] = newEventFunc(f)
}

func (h *namespaceHandler) getTypes(event string) []reflect.Type {
	namespaceHandler := h.events[event]
	if namespaceHandler == nil {
		return nil
	}
	return namespaceHandler.argTypes
}

func (h *namespaceHandler) dispatch(c Conn, event string, args []reflect.Value) ([]reflect.Value, error) {
	// onConnect event
	if event == "" {
		var err error
		if h.onConnect != nil {
			err = h.onConnect(c)
		}
		return nil, err
	}

	funcHandler := h.events[event]
	if funcHandler == nil {
		return c.Dispatch(event, args)
	}
	return funcHandler.Call(append([]reflect.Value{reflect.ValueOf(c)}, args...))
}

type namespaceConn struct {
	*conn
	namespace string
	context   interface{}
	broadcast Broadcast
	events    map[string]*funcHandler
}

func newNamespaceConn(conn *conn, namespace string, broadcast Broadcast) *namespaceConn {
	return &namespaceConn{
		conn:      conn,
		namespace: namespace,
		broadcast: broadcast,
		events:    make(map[string]*funcHandler),
	}
}

func (c *namespaceConn) SetContext(v interface{}) {
	c.context = v
}

func (c *namespaceConn) Context() interface{} {
	return c.context
}

func (c *namespaceConn) Namespace() string {
	return c.namespace
}

func (c *namespaceConn) Emit(event string, v ...interface{}) {
	args := make([]reflect.Value, len(v)+1)
	args[0] = reflect.ValueOf(event)
	for i := 1; i < len(args); i++ {
		args[i] = reflect.ValueOf(v[i-1])
	}
	c.conn.write(args)
}

func (c *namespaceConn) Set(v interface{}) {
	c.Emit("set", v)
}

func (c *namespaceConn) Join(room string) {
	c.broadcast.Join(room, c)
}

func (c *namespaceConn) Leave(room string) {
	c.broadcast.Leave(room, c)
}

func (c *namespaceConn) LeaveAll() {
	c.broadcast.LeaveAll(c)
}

func (c *namespaceConn) Rooms() []string {
	return c.broadcast.Rooms(c)
}

func (c *namespaceConn) ID() string {
	url := c.URL()
	values := url.Query()
	id := values.Get("id")
	return id
}

func (c *namespaceConn) OnEvent(event string, f interface{}) {
	c.events[event] = newEventFunc(f)
}

func (c *namespaceConn) RemoveEvent(event string) {
	delete(c.events, event)
}

func (c *namespaceConn) Err(msg string) {
	c.Emit("error", msg)
}

func (c *namespaceConn) Dispatch(event string, args []reflect.Value) ([]reflect.Value, error) {
	funcHandler := c.events[event]
	if funcHandler == nil {
		return nil, nil
	}
	return funcHandler.Call(append([]reflect.Value{reflect.ValueOf(c)}, args...))
}
