package main

import (
	"fmt"
	"log"
	config2 "mtggameengine/config"
	"mtggameengine/models"
	"mtggameengine/services"
	socketio "mtggameengine/socket"
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
	helloHandler := services.NewHelloHandler(poolService)
	gameService := services.NewDefaultGameService(poolService)

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal("server error:", err)
	}
	server.OnConnect(func(s socketio.Conn) error {
		url := s.URL()
		values := url.Query()
		id := values.Get("id")
		name := values.Get("name")
		fmt.Println("Connnected id:", id, "name:", name)
		s.SetContext(HelloRequest{
			Id:   id,
			Name: name,
		})
		helloHandler(s)
		server.BroadcastToRoom("/", "/", "set", models.LobbyStats{Players: server.RoomLen("/", "/")})
		/**
		broadcast to all after join :
		["set",{"numPlayers":4,"numGames":5,"numActiveGames":3}]
		["set",{"roomInfo":[]}]
		*/
		return nil
	})
	server.OnEvent("create", func(s socketio.Conn, msg models.CreateGameRequest) {
		game, err := gameService.CreateGame(msg, s)
		if err != nil {
			s.Emit("error", err.Error())
			return
		}

		s.SetContext(game)
		s.Emit("route", "g/"+game.ID())

	})
	server.OnEvent("join", func(s socketio.Conn, msg string) {
		gameService.Join(msg, s)
	})
	server.OnError(func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})
	server.OnDisconnect(func(c socketio.Conn, reason string) {
		fmt.Println("socket disconnected", reason)
		c.Dispatch("leave", nil)
	})
	go server.Serve()
	defer server.Close()

	http.Handle("/engine.io/", server)
	http.Handle("/", http.FileServer(http.Dir(config.FrontendDir)))
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
