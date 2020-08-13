package main

import (
	"fmt"
	socketio "github.com/herveh44/go-socket.io"
	"log"
	config2 "mtggameengine/config"
	"mtggameengine/models"
	"mtggameengine/services"
	"net/http"
)

type HelloRequest struct {
	Id   string
	Name string
}

func main() {
	config, err := config2.Setup()
	if err != nil {
		log.Fatal("load config error:", err)
	}

	poolService := services.NewPoolService(config.PoolServiceBaseURL)

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal("server error:", err)
	}
	server.OnConnect(func(s socketio.Conn) error {
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
			MTGJsonVersion:     models.MTGJsonVersion{Version: "asd", Date: "asd"},
			BoosterRuleVersion: "asd",
		}

		if version, err := poolService.GetVersion(); err == nil {
			response.MTGJsonVersion = models.MTGJsonVersion{
				Version: version.Version,
				Date:    version.Date,
			}
			response.BoosterRuleVersion = version.Version
		}

		s.Emit("set", response)
		return nil
	})
	server.OnEvent("create", func(s socketio.Conn, msg models.CreateGame) {
		fmt.Println("create:", msg)
		s.Emit("route", "g/asd")
	})
	server.OnEvent("join", func(s socketio.Conn, msg string) {
		fmt.Println("join:", msg)
	})
	server.OnError(func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})
	server.OnDisconnect(func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})
	go server.Serve()
	defer server.Close()

	http.Handle("/engine.io/", server)
	http.Handle("/", http.FileServer(http.Dir(config.FrontendDir)))
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
