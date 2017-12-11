package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/tahkapaa/test_stuff/riot_api/riotapi"
)

var (
	Token   string
	RiotKey string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&RiotKey, "r", "", "Riot API key")
	flag.Parse()
}

// Player is a lol player
type Player struct {
	Name          string
	ID            int
	CurrentGameID int
	Rank          string
}

var games map[int]*riotapi.CurrentGameInfo
var reportedGames = make(map[int]bool)

func main() {
	rand.Seed(time.Now().UnixNano())
	games = make(map[int]*riotapi.CurrentGameInfo)

	bot, err := newBot(Token, RiotKey)
	if err != nil {
		fmt.Printf("Unable to create bot: %v\n", err)
		os.Exit(1)
	}
	go bot.monitorPlayers()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	bot.Discord.Close()
}


func handleMonitorPlayer(c *riotapi.Client, p Player) {
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
	// sendMessage(newEmbedMessage("Peli loppui"))
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

func (b *Bot) reportGames() {
	for _, game := range games {
		b.reportGame(game)
	}
}

func (b *Bot) reportGame(cgi *riotapi.CurrentGameInfo) {
	if reportedGames[cgi.GameID] {
		return
	}

	ggTeam := b.findGoodGuysTeamID(cgi)
	ggPlayers := getPlayersOfTeam(ggTeam, cgi)
	ggReport, err := reportTeam(b.RC, "Good guys", ggPlayers)
	if err != nil {
		fmt.Println(err)
		return
	}
	bgPlayers := getPlayersOfTeam(getOpposingTeam(ggTeam), cgi)
	bgReport, err := reportTeam(b.RC, "Bad guys", bgPlayers)
	if err != nil {
		fmt.Println(err)
		return
	}

	if ggTeam == teamRed {
		ggReport.Color = red
	} else {
		bgReport.Color = red
	}

	b.Discord.ChannelMessageSendComplex(b.ChannelID, &discordgo.MessageSend{Embed: ggReport})
	b.Discord.ChannelMessageSendComplex(b.ChannelID, &discordgo.MessageSend{Embed: bgReport})

	reportedGames[cgi.GameID] = true
}

func reportTeam(c *riotapi.Client, title string, cgi []riotapi.CurrentGameParticipant) (*discordgo.MessageEmbed, error) {
	fields := make([]*discordgo.MessageEmbedField, 0)
	for _, cgp := range cgi {
		mef, err := npcMessageEmbedField(c, &cgp)
		if err != nil {
			return nil, err
		}
		fields = append(fields, mef)
	}
	em := newEmbedMessage(title)
	em.Fields = fields
	return em, nil
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
		Inline: true,
	}, nil
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

func (b *Bot) findGoodGuysTeamID(cgi *riotapi.CurrentGameInfo) int {
	for _, cgp := range cgi.Participants {
		if !b.isPlayerNPC(&cgp) {
			return cgp.TeamID
		}
	}
	return teamNone
}

func (b *Bot) isPlayerNPC(cgp *riotapi.CurrentGameParticipant) bool {
	for _, p := range b.FollowedPlayers {
		if p.ID == cgp.SummonerID {
			return false
		}
	}
	return true
}
