package models

type HelloResponse struct {
	AvailableSets      map[string]interface{} `json:"availableSets,omitempty"`
	LatestSet          interface{}            `json:"latestSet,omitempty"`
	MTGJsonVersion     MTGJsonVersion         `json:"mtgJsonVersion,omitempty"`
	BoosterRuleVersion string                 `json:"boosterRulesVersion,omitempty"`
}

type MTGJsonVersion struct {
	Version string `json:"version,omitempty"`
	Date    string `json:"date,omitempty"`
}

type CreateGame struct {
	Type       string   `json:"type"`
	Seats      int      `json:"seats"`
	Title      string   `json:"title"`
	IsPrivate  bool     `json:"isPrivate"`
	ModernOnly bool     `json:"modernOnly"`
	TotalChaos bool     `json:"totalChaos"`
	Sets       []string `json:"sets"`
}
