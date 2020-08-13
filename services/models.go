package services

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

type Card struct {
	UUID              string `json:"uuid" gorm:"primary_key"`
	SetID             string `json:"setCode" gorm:"index:pack_idx"`
	Name              string `json:"name"`
	Number            string `json:"number"`
	Layout            string `json:"layout"`
	Loyalty           string `json:"loyalty"`
	Power             string `json:"power"`
	Toughness         string `json:"toughness"`
	ConvertedManaCost int    `json:"cmc"`
	Type              string `json:"type"`
	ManaCost          string `json:"manaCost"`
	Rarity            string `json:"rarity" gorm:"index:pack_idx"`
	Side              string `json:"side"`
	IsAlternative     bool   `json:"isAlternative"`
	Color             string `json:"color"`
	ScryfallID        string `json:"scryfallId"`
	URL               string `json:"url"`
	Cubable           bool   `json:"-" gorm:"index:cubable_idx"`
	FaceName          string `json:"-" gorm:"index:cubable_idx"`
}

type CardResponse struct {
	Card

	Id   string `json:"cardId"`
	Foil bool   `json:"foil"`
}

type CardPool []CardResponse

type ChaosRequest struct {
	Players    uint `json:"players"`
	Packs      uint `json:"packs"`
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
