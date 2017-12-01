package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tahkapaa/test_stuff/riot_api/riotapi"
)

const (
	apiKey = "RGAPI-4e5b5478-e606-47ee-ac4d-b6easdfdasd"
)

var players = []*Player{
	{Name: "Uxipaxa", ID: 24749077},
	{Name: "Invataxi", ID: 31507600},
	{Name: "Ignusnus", ID: 25251553},
	{Name: "Opettaja", ID: 28490422},
}

// Player is a lol player
type Player struct {
	Name     string
	ID       int
	InGame   bool
	Champion riotapi.Champion
}

func main() {
	c, err := riotapi.New(apiKey, "eune", 50, 20)
	if err != nil {
		log.Fatalf("unable to initialize riot api: %v", err)
	}

	sendToDiscord("Iltaa kaikki, olen täällä taas!")
	monitorPlayers(c)
}

func monitorPlayers(c *riotapi.Client) {
	for {
		for _, p := range players {
			handleMonitorPlayer(c, p)
		}
		time.Sleep(time.Second * 60)
	}
}

func handleMonitorPlayer(c *riotapi.Client, p *Player) {
	cgi, err := c.Spectator.ActiveGamesBySummoner(p.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if cgi == nil {
		if p.InGame {
			sendToDiscord(p.Name + " lopetti pelin")
		}
		p.InGame = false
		return
	}

	if p.InGame {
		return
	}

	for _, playerInfo := range cgi.Participants {
		champions, err := c.StaticData.Champions()
		if err != nil {
			fmt.Println(err)
			return
		}
		p.Champion = champions.Data[playerInfo.ChampionID]
	}

	if !p.InGame {
		sendToDiscord(fmt.Sprintf("%v meni peliin", p.Name))
		imageToDiscord(p.Champion.Name)
	}
	p.InGame = true
}
