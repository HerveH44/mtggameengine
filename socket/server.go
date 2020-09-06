package socketio

import (
	"net/http"

	engineio "github.com/googollee/go-engine.io"
)

// Server is a go-socket.io server.
type Server struct {
	handler *namespaceHandler
	eio     *engineio.Server
}

// NewServer returns a server.
func NewServer(c *engineio.Options) (*Server, error) {
	eio, err := engineio.NewServer(c)
	if err != nil {
		return nil, err
	}
	return &Server{
		handler: newHandler(),
		eio:     eio,
	}, nil
}

// Close closes server.
func (s *Server) Close() error {
	return s.eio.Close()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.eio.ServeHTTP(w, r)
}

// OnConnect set a handler function f to handle open event for
// namespace nsp.
func (s *Server) OnConnect(f func(Conn) error) {
	s.handler.OnConnect(f)
}

// OnDisconnect set a handler function f to handle disconnect event for
// namespace nsp.
func (s *Server) OnDisconnect(f func(Conn, string)) {
	s.handler.OnDisconnect(f)
}

// OnError set a handler function f to handle error for namespace nsp.
func (s *Server) OnError(f func(Conn, error)) {
	s.handler.OnError(f)
}

// OnEvent set a handler function f to handle event for namespace nsp.
func (s *Server) OnEvent(event string, f interface{}) {
	s.handler.OnEvent(event, f)
}

// Serve serves go-socket.io server
func (s *Server) Serve() error {
	for {
		conn, err := s.eio.Accept()
		if err != nil {
			return err
		}
		go s.serveConn(conn)
	}
}

// Broadcast broadcasts given event & args to all the connections in the room
func (s *Server) Broadcast(event string, args ...interface{}) bool {
	s.handler.broadcast.Send("/", event, args...)
	return true
}

// ConnectionsLen gives number of connections in the room
func (s *Server) ConnectionsLen() int {
	return s.handler.broadcast.Len("/")
}

func (s *Server) serveConn(c engineio.Conn) {
	_, err := newConn(c, s.handler)
	if err != nil {
		if s.handler.onError != nil {
			s.handler.onError(nil, err)
		}
		return
	}
}
