package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"

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

	p.CurrentGameID = cgi.GameID
	games[cgi.GameID] = cgi

}

func endGame(g *riotapi.CurrentGameInfo) {
	if g == nil {
		return
	}
	fmt.Println("Peli loppui")
	sendMessage(newEmbedMessage("Peli loppui"))
	delete(games, g.GameID)
	delete(reportedGames, g.GameID)
}

const (
	teamNone = 0
	teamRed  = 200
	teamBlue = 100
)

func getOpposingTeam(teamID int) int {
	if teamID == teamBlue {
		return teamRed
	}
	return teamBlue
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

	ggTeam := findGoodGuysTeamID(cgi)
	ggPlayers := getPlayersOfTeam(ggTeam, cgi)
	ggReport := reportTeam(c, "Hyvikset", ggPlayers)
	ggReport.Author = &discordgo.MessageEmbedAuthor{Name: "Peli alkoi"}
	sendMessage(ggReport)
	bgPlayers := getPlayersOfTeam(getOpposingTeam(ggTeam), cgi)
	bgReport := reportTeam(c, "Pahikset", bgPlayers)
	sendMessage(bgReport)

	reportedGames[cgi.GameID] = true
}

func reportTeam(c *riotapi.Client, title string, cgi []riotapi.CurrentGameParticipant) *discordgo.MessageEmbed {
	fields := make([]*discordgo.MessageEmbedField, 0)
	for _, cgp := range cgi {
		mef, err := npcMessageEmbedField(c, &cgp)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		fields = append(fields, mef)
	}
	em := newEmbedMessage(title)
	em.Fields = fields
	return em
}

func npcMessageEmbedField(c *riotapi.Client, cgp *riotapi.CurrentGameParticipant) (*discordgo.MessageEmbedField, error) {
	champions, err := c.StaticData.Champions()
	if err != nil {
		return nil, err
	}
	rank, err := findPlayerRank(c, cgp.SummonerID)
	if err != nil {
		return nil, err
	}
	return &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s - %s", cgp.SummonerName, rank),
		Value:  champions.Data[cgp.ChampionID].Name,
		Inline: true}, nil
}

func findPlayerRank(c *riotapi.Client, summonerID int) (string, error) {
	si, err := c.Summoner.SummonerByID(summonerID)
	if err != nil {
		return "", err
	}
	recentMatches, err := c.Match.RecentMatchesByAccountID(si.AccountID)
	if err != nil {
		return "", err
	}
	if len(recentMatches.Matches) > 0 {
		match, err := c.Match.MatchByID(recentMatches.Matches[0].GameID)
		if err != nil {
			return "", err
		}
		var participantID int
		for _, ident := range match.ParticipantIdentities {
			if ident.Player.AccountID == si.AccountID {
				participantID = ident.ParticipantID
				break
			}
		}
		for _, p := range match.Participants {
			if p.ParticipantID == participantID {
				return p.HighestAchievedSeasonTier, nil
			}
		}
	}
	return "Not found", nil
}

func getPlayersOfTeam(teamID int, cgi *riotapi.CurrentGameInfo) []riotapi.CurrentGameParticipant {
	var team []riotapi.CurrentGameParticipant
	for _, cgp := range cgi.Participants {
		if cgp.TeamID == teamID {
			team = append(team, cgp)
		}
	}
	return team
}

func findGoodGuysTeamID(cgi *riotapi.CurrentGameInfo) int {
	for _, cgp := range cgi.Participants {
		if !isPlayerNPC(&cgp) {
			return cgp.TeamID
		}
	}
	return teamNone
}

func isPlayerNPC(cgp *riotapi.CurrentGameParticipant) bool {
	for _, p := range players {
		if p.ID == cgp.SummonerID {
			return false
		}
	}
	return true
}
