package main

import (
	"log"
	"mtggameengine/socket"
	"net/http"
)

type HelloRequest struct {
	Id   string
	Name string
}

func main() {
	server, err := socket.NewServer(nil)
	if err != nil {
		log.Fatal("server error:", err)
	}

	go server.Serve()
	defer server.Close()

	http.Handle("/engine.io/", server)
	http.Handle("/", http.FileServer(http.Dir("/home/noname/projects/dr4ft/built")))
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
