package models

type HelloResponse struct {
	AvailableSets      *AvailableSetsMap `json:"availableSets,omitempty"`
	LatestSet          LatestSetResponse `json:"latestSet,omitempty"`
	MTGJsonVersion     MTGJsonVersion    `json:"mtgJsonVersion,omitempty"`
	BoosterRuleVersion string            `json:"boosterRulesVersion,omitempty"`
}

type MTGJsonVersion struct {
	Version string `json:"version,omitempty"`
	Date    string `json:"date,omitempty"`
}

type CreateGameRequest struct {
	Type            string   `json:"type"`
	Seats           int      `json:"seats"`
	Title           string   `json:"title"`
	IsPrivate       bool     `json:"isPrivate"`
	ChaosPackNumber int      `json:"chaosPacksNumber"`
	ModernOnly      bool     `json:"modernOnly"`
	TotalChaos      bool     `json:"totalChaos"`
	Sets            []string `json:"sets"`
	Cube            struct {
		List         string `json:"list"`
		Cards        int    `json:"cards"`
		Packs        int    `json:"packs"`
		CubePoolSize int    `json:"cubePoolSize"`
	} `json:"cube"`
}

type LobbyStats struct {
	Players     int `json:"numPlayers"`
	Games       int `json:"numGames"`
	ActiveGames int `json:"numActiveGames"`
}
