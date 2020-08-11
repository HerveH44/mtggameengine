package main

import (
	"encoding/json"
	"fmt"
	engineio "github.com/googollee/go-engine.io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type HelloRequest struct {
	Id   string
	Name string
}

type MTGJsonVersion struct {
	Version string `json:"version"`
	Date    string `json:"date"`
}

type HelloResponse struct {
	AvailableSets      map[string]interface{} `json:"availableSets"`
	LatestSet          interface{}            `json:"latestSet"`
	MTGJsonVersion     MTGJsonVersion         `json:"mtgJsonVersion"`
	BoosterRuleVersion string                 `json:"boosterRulesVersion"`
}

func main() {
	server, err := engineio.NewServer(nil)
	if err != nil {
		log.Fatal("server error:", err)
	}

	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				log.Fatalln("accept error:", err)
				return
			}

			go func(conn engineio.Conn) {
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

				response := HelloResponse{
					MTGJsonVersion:     MTGJsonVersion{Version: "4.0.1", Date: "asd"},
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
			}(conn)
		}
	}()

	http.Handle("/engine.io/", server)
	http.Handle("/", http.FileServer(http.Dir("/home/noname/projects/dr4ft/built")))
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
