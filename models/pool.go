package models

type Card struct {
	Id                string `json:"cardId"`
	Foil              bool   `json:"foil"`
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
}

type Pack []Card

type Pool []Pack

func (p *Pool) Shift() Pack {
	pack := (*p)[0]
	*p = (*p)[:len(*p)]
	return pack
}
