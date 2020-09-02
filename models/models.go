package models

type SetResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type LatestSetResponse struct {
	SetResponse
	Type string `json:"type"`
}

type VersionResponse struct {
	Date    string `json:"date"`
	Version string `json:"version"`
}

type RegularRequest struct {
	Players int      `json:"players"`
	Sets    []string `json:"sets"`
}

type ChaosRequest struct {
	Players    int  `json:"players"`
	Packs      int  `json:"packs"`
	Modern     bool `json:"modern"`
	TotalChaos bool `json:"totalChaos"`
}

type CubeRequest struct {
	Cubelist       []string `json:"list"`
	Players        uint     `json:"players"`
	PlayerPackSize uint     `json:"playerPackSize"`
	Packs          uint     `json:"packs"`
}

type CubeListRequest struct {
	Cubelist []string `json:"list"`
}

type CubeListErrorResponse struct {
	Error []string `json:"error"`
}

type AvailableSetsMap map[string][]SetResponse
