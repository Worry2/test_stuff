package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/tahkapaa/test_stuff/riot_api/riotapi"
)

const runesURL = "http://ec2-54-171-95-164.eu-west-1.compute.amazonaws.com:3000/runesReforged/"

var (
	Token   string
	RiotKey string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&RiotKey, "r", "", "Riot API key")
	flag.Parse()
}

// RC is the riotapi client
var RC *riotapi.Client

// DGBot is the discord bot
var DGBot *Bot

// Player is a lol player
type Player struct {
	Name          string
	ID            int
	CurrentGameID int
	Rank          string
}

func main() {
	rand.Seed(time.Now().UnixNano())

	c, err := riotapi.New(RiotKey, "eune", 50, 20)
	if err != nil {
		log.Fatalf("unable to initialize riot api: %v", err)
	}
	RC = c

	bot, err := newBot(Token, New("channels"))
	if err != nil {
		fmt.Printf("Unable to create bot: %v\n", err)
		os.Exit(1)
	}
	DGBot = bot

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	bot.Discord.Close()
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

func (pm *PlayerMonitor) reportGames() {
	for _, game := range pm.games {
		pm.reportGame(game)
	}
}

func (pm *PlayerMonitor) reportGame(cgi *riotapi.CurrentGameInfo) {
	if pm.reportedGames[cgi.GameID] {
		return
	}

	ggTeam := pm.findGoodGuysTeamID(cgi)
	ggPlayers := getPlayersOfTeam(ggTeam, cgi)
	ggReport, err := reportTeam(RC, "Good guys", ggPlayers)
	if err != nil {
		fmt.Println(err)
		return
	}
	bgPlayers := getPlayersOfTeam(getOpposingTeam(ggTeam), cgi)
	bgReport, err := reportTeam(RC, "Bad guys", bgPlayers)
	if err != nil {
		fmt.Println(err)
		return
	}

	if ggTeam == teamRed {
		ggReport.Color = red
	} else {
		bgReport.Color = red
	}

	DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &discordgo.MessageSend{Embed: ggReport})
	DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &discordgo.MessageSend{Embed: bgReport})

	bgrunes := reportRunes(bgPlayers)
	for _, msg := range bgrunes {
		DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &discordgo.MessageSend{Embed: msg})
	}
	ggrunes := reportRunes(ggPlayers)
	for _, msg := range ggrunes {
		DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &discordgo.MessageSend{Embed: msg})
	}

	pm.reportedGames[cgi.GameID] = true
}

func (pm *PlayerMonitor) findGoodGuysTeamID(cgi *riotapi.CurrentGameInfo) int {
	for _, cgp := range cgi.Participants {
		if !pm.isPlayerNPC(&cgp) {
			return cgp.TeamID
		}
	}
	return teamNone
}

func (pm *PlayerMonitor) isPlayerNPC(cgp *riotapi.CurrentGameParticipant) bool {
	for _, p := range pm.FollowedPlayers {
		if p.ID == cgp.SummonerID {
			return false
		}
	}
	return true
}

func reportTeam(c *riotapi.Client, title string, cgp []riotapi.CurrentGameParticipant) (*discordgo.MessageEmbed, error) {
	fields := make([]*discordgo.MessageEmbedField, 0)
	for _, gp := range cgp {
		mef, err := npcMessageEmbedField(c, &gp)
		if err != nil {
			return nil, err
		}
		fields = append(fields, mef)
	}
	em := newEmbedMessage(title)
	em.Fields = fields
	return em, nil
}

func reportRunes(cgp []riotapi.CurrentGameParticipant) []*discordgo.MessageEmbed {
	var mex []*discordgo.MessageEmbed
	for _, gp := range cgp {
		fmt.Println(gp.SummonerName, fmt.Sprintf("%s%s%d", runesURL, "perkStyle/", gp.Perks.PerkStyle))
		msg := discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    gp.SummonerName,
				IconURL: fmt.Sprintf("%s%s%d.png", runesURL, "perkStyle/", gp.Perks.PerkStyle),
			},
			Title: gp.SummonerName,
		}
		mex = append(mex, &msg)
	}
	return mex
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
