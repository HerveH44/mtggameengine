package main

import (
	"fmt"
	socketio "github.com/herveh44/go-socket.io"
	"log"
	"mtggameengine/models"
	"net/http"
)

type HelloRequest struct {
	Id   string
	Name string
}

func main() {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal("server error:", err)
	}
	server.OnConnect("/", func(s socketio.Conn) error {
		url := s.URL()
		values := url.Query()
		id := values.Get("id")
		name := values.Get("name")
		fmt.Println("id:", id, "name:", name)
		s.SetContext(HelloRequest{
			Id:   id,
			Name: name,
		})
		fmt.Println("connected:", s.ID())

		response := models.HelloResponse{
			MTGJsonVersion:     models.MTGJsonVersion{Version: "4.0.1", Date: "asd"},
			BoosterRuleVersion: "asd",
		}
		s.Emit("set", response)
		return nil
	})
	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})
	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		return "recv " + msg
	})
	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})
	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})
	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})
	go server.Serve()
	defer server.Close()

	http.Handle("/engine.io/", server)
	http.Handle("/", http.FileServer(http.Dir("/home/noname/projects/dr4ft/built")))
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
