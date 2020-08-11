package socket

import (
	"encoding/json"
	"fmt"
	engineio "github.com/googollee/go-engine.io"
	"io/ioutil"
	"log"
	"mtggameengine/models"
	"net/http"
	"sync"
)

type Server struct {
	eio *engineio.Server
}

func NewServer(c *engineio.Options) (*Server, error) {
	eio, err := engineio.NewServer(c)
	if err != nil {
		return nil, err
	}
	return &Server{
		eio: eio,
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.eio.ServeHTTP(w, r)
}

// Close closes server.
func (s *Server) Close() error {
	return s.eio.Close()
}

// Serve serves go-socket.io server
func (s *Server) Serve() error {
	for {
		conn, err := s.eio.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			return err
		}
		go s.serveConn(conn)
	}
}

func (s *Server) serveConn(conn engineio.Conn) {
	defer conn.Close()
	fmt.Println(conn.ID(), conn.RemoteAddr(), "->", conn.LocalAddr(), "with", conn.RemoteHeader())

	type arg struct {
		typ  engineio.FrameType
		data []byte
	}
	data := make(chan arg, 10)
	closeChan := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer fmt.Println(conn.ID(), "write quit")
		defer wg.Done()

		for {
			select {
			case d := <-data:
				w, err := conn.NextWriter(d.typ)
				if err != nil {
					log.Println("next write error:", err)
					return
				}
				if _, err := w.Write(d.data); err != nil {
					w.Close()
					log.Println("write error:", err)
					return
				}
				if err := w.Close(); err != nil {
					log.Println("write close error:", err)
					return
				}
			case <-closeChan:
				return
			}
		}
	}()

	response := models.HelloResponse{
		MTGJsonVersion:     models.MTGJsonVersion{Version: "4.0.1", Date: "asd"},
		BoosterRuleVersion: "asd",
	}

	marshal, _ := json.Marshal(response)

	data <- arg{
		typ:  engineio.TEXT,
		data: []byte(fmt.Sprintf("[\"set\", %s]", marshal)),
	}

	for {
		typ, r, err := conn.NextReader()
		if err != nil {
			log.Println("next read error:", err)
			break
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			r.Close()
			log.Println("read all error:", err)
			break
		}
		switch typ {
		case engineio.BINARY:
			fmt.Println("read: binary", b)
		case engineio.TEXT:
			fmt.Println("read: text", string(b))
		}
		data <- arg{
			typ:  typ,
			data: b,
		}
		if err := r.Close(); err != nil {
			log.Println("close reader error:", err)
			break
		}
	}

	close(closeChan)
	wg.Wait()
}
