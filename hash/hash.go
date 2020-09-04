package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"log"
	"mtggameengine/models"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type opts struct {
	algo      hash.Hash
	separator string
	prefix    string
	name      func(string) string
	digest    func(string) string
}

var (
	cock = opts{
		algo:      sha1.New(),
		separator: ";",
		prefix:    "SB:",
		name: func(name string) string {
			return strings.ToLower(name)
		},
		digest: func(data string) string {
			log.Println(data)
			parseInt, err := strconv.ParseInt(data[:10], 16, 64)
			if err != nil {
				log.Println(err)
				return ""
			}
			return strconv.FormatInt(parseInt, 32)
		},
	}
	mws = opts{
		algo:      md5.New(),
		separator: "",
		prefix:    "#",
		name: func(name string) string {
			re := regexp.MustCompile("[^A-Z]")
			upperName := strings.ToUpper(name)
			return re.ReplaceAllString(upperName, "")
		},
		digest: func(data string) string {
			return data[:8]
		},
	}
)

func makeHash(deck models.HashEvent, opt opts) (string, error) {
	data := makeData(deck, opt)
	return digest(opt, data)
}

func digest(opt opts, data string) (string, error) {
	h := opt.algo
	if _, err := h.Write([]byte(data)); err != nil {
		return "", err
	}
	sum := h.Sum(nil)
	digest := hex.EncodeToString(sum)
	return opt.digest(digest), nil
}

func makeData(deck models.HashEvent, opt opts) string {
	items := make([]string, 0)
	prefix := ""
	cards := deck.Main
	for cardName, num := range cards {
		item := prefix + opt.name(cardName)
		for i := 0; i < num; i++ {
			items = append(items, item)
		}
	}
	cardsSide := deck.Side
	prefix = opt.prefix
	for cardName, num := range cardsSide {
		item := prefix + opt.name(cardName)
		for i := 0; i < num; i++ {
			items = append(items, item)
		}
	}

	sort.Strings(items)
	return strings.Join(items, opt.separator)
}

func MakeCockatriceHash(deck models.HashEvent) (string, error) {
	return makeHash(deck, cock)
}

func MakeMWSHash(deck models.HashEvent) (string, error) {
	return makeHash(deck, mws)
}
