package hash

import (
	"log"
	"mtggameengine/models"
	"testing"
)

func TestHash(t *testing.T) {
	main := make(map[string]int)
	main["blabla"] = 1
	main["blabla2"] = 2

	side := make(map[string]int)
	side["blabla"] = 1
	side["blabla2"] = 2

	event := models.HashEvent{
		Main: main,
		Side: side,
	}

	hash, err := MakeMWSHash(event)

	log.Println(hash)
	log.Println(err)

}
