package game

import "mtggameengine/game/pl"

type PlayerSpecificInfo struct {
	Name        string  `json:"name"`
	Time        int     `json:"time"`
	Packs       int     `json:"packs"`
	IsBot       bool    `json:"isBot"`
	IsConnected bool    `json:"isConnected"`
	Hash        pl.Hash `json:"hash"`
}

type StateInfo struct {
	Players   *[]PlayerSpecificInfo `json:"players"`
	GameSeats int                   `json:"gameSeats"`
}

type StartRequest struct {
	AddBots        bool   `json:"addBots"`
	UseTimer       bool   `json:"useTimer"`
	TimerLength    string `json:"timerLength"`
	ShufflePlayers bool   `json:"shufflePlayers"`
}

type PlayerBasicInfo struct {
	IsHost   bool     `json:"isHost,omitempty"`
	Round    *int     `json:"round,omitempty"`
	Self     *int     `json:"self,omitempty"`
	Sets     []string `json:"sets,omitempty"`
	GameId   string   `json:"gameId,omitempty"`
	PackSize *int     `json:"packSize,omitempty"`
}

type BasicInfos struct {
	Type       string   `json:"type"`
	PacksInfos string   `json:"packsInfo"`
	Sets       []string `json:"sets"`
}
