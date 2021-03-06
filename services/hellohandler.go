package services

import (
	"fmt"
	"mtggameengine/models"
	"mtggameengine/services/pool"
	socketio "mtggameengine/socket"
)

type HelloHandler func(conn socketio.Conn)

func NewHelloHandler(poolService pool.Service) HelloHandler {
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
