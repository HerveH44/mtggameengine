package services

import (
	"fmt"
	socketio "github.com/herveh44/go-socket.io"
	"mtggameengine/models"
)

type HelloHandler func(conn socketio.Conn)

func NewHelloHandler(poolService PoolService) HelloHandler {
	return func(conn socketio.Conn) {
		version, err := poolService.GetVersion()
		if err != nil {
			fmt.Println("newhellohandler: error while fetching version", err)
			return
		}

		availableSetsMap, err := poolService.GetAvailableSets()
		if err != nil {
			fmt.Println("newhellohandler: error while fetching availableSetsMap", err)
			return
		}

		latestSet, err := poolService.GetLatestSet()
		if err != nil {
			fmt.Println("newhellohandler: error while fetching latestSet", err)
			return
		}

		response := models.HelloResponse{
			AvailableSets: availableSetsMap,
			LatestSet:     latestSet,
			MTGJsonVersion: models.MTGJsonVersion{
				Version: version.Version,
				Date:    version.Date,
			},
			BoosterRuleVersion: version.Version,
		}

		conn.Emit("set", response)
	}
}
