package socketio

import (
	"mtggameengine/socket/parser"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	engineio "github.com/googollee/go-engine.io"
)

// Conn is a connection in go-socket.io
type Conn interface {
	// ID returns session id
	ID() string
	Name() string
	SetName(name string)

	Close() error
	URL() url.URL
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	RemoteHeader() http.Header

	// Context of this connection. You can save one context for one
	// connection, and share it between all handlers. The handlers
	// is called in one goroutine, so no need to lock context if it
	// only be accessed in one connection.
	Context() interface{}
	SetContext(v interface{})
	Namespace() string
	Emit(msg string, v ...interface{})
	Set(interface{})
	Err(msg string)

	// Broadcast server side apis
	Join(room string)
	Leave(room string)
	LeaveAll()
	Rooms() []string

	// Attempt to handle event per conn
	OnEvent(event string, f interface{})
	Dispatch(event string, args []reflect.Value) ([]reflect.Value, error)
	RemoveEvent(event string)
	RemoveAllEvents()
}

type errorMessage struct {
	namespace string
	error
}

type writePacket struct {
	data []interface{}
}

type conn struct {
	engineio.Conn
	name      string
	encoder   *parser.Encoder
	decoder   *parser.Decoder
	errorChan chan errorMessage
	writeChan chan writePacket
	quitChan  chan struct{}
	handler   *namespaceHandler
	namespace *namespaceConn
	closeOnce sync.Once
	id        uint64
}

func newConn(c engineio.Conn, handler *namespaceHandler) (*conn, error) {
	url := c.URL()
	values := url.Query()
	name := values.Get("name")

	ret := &conn{
		Conn:      c,
		name:      name,
		encoder:   parser.NewEncoder(c),
		decoder:   parser.NewDecoder(c),
		errorChan: make(chan errorMessage),
		writeChan: make(chan writePacket),
		quitChan:  make(chan struct{}),
		handler:   handler,
	}
	if err := ret.connect(); err != nil {
		ret.Close()
		return nil, err
	}
	return ret, nil
}

func (c *conn) Name() string {
	return c.name
}

func (c *conn) SetName(name string) {
	c.name = name
}

func (c *conn) Close() error {
	var err error
	c.closeOnce.Do(func() {
		c.namespace.LeaveAll()
		if c.handler != nil && c.handler.onDisconnect != nil {
			c.handler.onDisconnect(c.namespace, "client namespace disconnect")
		}
		err = c.Conn.Close()
		close(c.quitChan)
	})
	return err
}

func (c *conn) connect() error {
	root := newNamespaceConn(c, "/", c.handler.broadcast)
	root.Join("/")
	root.SetContext(c.Conn.Context())

	go c.serveError()
	go c.serveWrite()
	go c.serveRead()

	c.handler.dispatch(root, "", nil)

	return nil
}

func (c *conn) write(args []reflect.Value) {
	data := make([]interface{}, len(args))
	for i := range data {
		data[i] = args[i].Interface()
	}
	pkg := writePacket{
		data: data,
	}
	select {
	case c.writeChan <- pkg:
	case <-c.quitChan:
		return
	}
}

func (c *conn) onError(namespace string, err error) {
	onErr := errorMessage{
		namespace: namespace,
		error:     err,
	}
	select {
	case c.errorChan <- onErr:
	case <-c.quitChan:
		return
	}
}

func (c *conn) parseArgs(types []reflect.Type) ([]reflect.Value, error) {
	return c.decoder.DecodeArgs(types)
}

func (c *conn) serveError() {
	defer c.Close()
	for {
		select {
		case <-c.quitChan:
			return
		case msg := <-c.errorChan:
			if c.handler != nil && c.handler.onError != nil {
				c.handler.onError(c.namespace, msg.error)
			}
		}
	}
}

func (c *conn) serveWrite() {
	defer c.Close()
	for {
		select {
		case <-c.quitChan:
			return
		case pkg := <-c.writeChan:
			if err := c.encoder.Encode(pkg.data); err != nil {
				c.onError("pkg.header.Namespace", err)
			}
		}
	}
}

func (c *conn) serveRead() {
	defer c.Close()
	var event string
	for {
		if err := c.decoder.DecodeHeader(&event); err != nil {
			c.onError("", err)
			return
		}

		// Connection event handling
		if c.namespace.HasEvent(event) {
			types := c.namespace.getTypes(event)
			args, err := c.decoder.DecodeArgs(types)
			if err != nil {
				c.onError("header.Namespace", err)
				return
			}
			ret, err := c.namespace.Dispatch(event, args)
			if err != nil {
				c.onError("header.Namespace", err)
				return
			}
			if len(ret) > 0 {
				c.write(ret)
			}
		} else {
			// Default handler
			types := c.handler.getTypes(event)
			args, err := c.decoder.DecodeArgs(types)
			if err != nil {
				c.onError("header.Namespace", err)
				return
			}
			ret, err := c.handler.dispatch(c.namespace, event, args)
			if err != nil {
				c.onError("header.Namespace", err)
				return
			}
			if len(ret) > 0 {
				c.write(ret)
			}
		}
	}
}
