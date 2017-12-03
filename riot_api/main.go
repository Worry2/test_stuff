package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tahkapaa/test_stuff/riot_api/riotapi"
)

const (
	apiKey = "RGAPI-8b46dbb2-825e-47d3-8b42-6065baaf018f"
)

var players = []*Player{
	{Name: "Uxipaxa", ID: 24749077},
	{Name: "Invataxi", ID: 31507600},
	{Name: "Ignusnus", ID: 25251553},
	{Name: "Opettaja", ID: 28490422},
}

// Player is a lol player
type Player struct {
	Name          string
	ID            int
	CurrentGameID int
	Champion      riotapi.Champion
}

var games map[int]*riotapi.CurrentGameInfo
var reportedGames = make(map[int]bool)

func main() {
	games = make(map[int]*riotapi.CurrentGameInfo)
	c, err := riotapi.New(apiKey, "eune", 50, 20)
	if err != nil {
		log.Fatalf("unable to initialize riot api: %v", err)
	}

	// sendToDiscord("Iltaa kaikki, olen täällä taas!")
	monitorPlayers(c)
}

func monitorPlayers(c *riotapi.Client) {
	for {
		for _, p := range players {
			handleMonitorPlayer(c, p)
		}
		reportGames(c, players)
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
		if p.CurrentGameID > 0 {
			endGame(games[p.CurrentGameID])
			p.CurrentGameID = 0
		}
		return
	}

	if p.CurrentGameID > 0 {
		return
	}

	for _, playerInfo := range cgi.Participants {
		if playerInfo.SummonerID == p.ID {
			champions, err := c.StaticData.Champions()
			if err != nil {
				fmt.Println(err)
				return
			}
			p.Champion = champions.Data[playerInfo.ChampionID]
			fmt.Println(playerInfo.Perks)
			break
		}
	}

	// if !p.InGame {
	// 	//imageToDiscord(p.Name, fmt.Sprintf("Pelaa hahmolla %s", p.Champion.Name), p.Champion.Name)
	// 	fmt.Println(p.Name, fmt.Sprintf("Pelaa hahmolla %s", p.Champion.Name), p.Champion.Name)

	// }
	p.CurrentGameID = cgi.GameID
	games[cgi.GameID] = cgi

}

func endGame(g *riotapi.CurrentGameInfo) {
	if g == nil {
		return
	}
	fmt.Println("Peli loppui")
	delete(games, g.GameID)
	delete(reportedGames, g.GameID)
}

func reportGames(c *riotapi.Client, xp []*Player) {
	for _, game := range games {
		reportGame(c, game)
	}
}

func reportGame(c *riotapi.Client, cgi *riotapi.CurrentGameInfo) {
	if reportedGames[cgi.GameID] {
		return
	}

	for _, playerInfo := range cgi.Participants {
		champions, err := c.StaticData.Champions()
		if err != nil {
			fmt.Println(err)
			return
		}
		champion := champions.Data[playerInfo.ChampionID]
		fmt.Println(playerInfo.SummonerName, fmt.Sprintf("Pelaa hahmolla %s", champion.Name), champion.Name)
	}
	reportedGames[cgi.GameID] = true
}
