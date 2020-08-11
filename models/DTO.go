package models

type HelloResponse struct {
	AvailableSets      map[string]interface{} `json:"availableSets"`
	LatestSet          interface{}            `json:"latestSet"`
	MTGJsonVersion     MTGJsonVersion         `json:"mtgJsonVersion"`
	BoosterRuleVersion string                 `json:"boosterRulesVersion"`
}

type MTGJsonVersion struct {
	Version string `json:"version"`
	Date    string `json:"date"`
}
