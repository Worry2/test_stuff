package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tahkapaa/test_stuff/riot_api/riotapi"
)

// PlayerMonitor monitors players
type PlayerMonitor struct {
	fpMutex         sync.Mutex
	FollowedPlayers map[string]Player
	games           map[int]*riotapi.CurrentGameInfo
	reportedGames   map[int]bool
	messageChan     chan monitorMessage
	ChannelID       string
	db              DB
}

type monitorMessage struct {
	kind          msgType
	player        Player
	sender        string
	timeStr       string
	summonerNames string
}

type msgType int

const (
	AddPlayer msgType = iota
	RemovePlayer
	ListPlayers
)

func (pm *PlayerMonitor) save() {
	if err := pm.db.Save(&ChannelData{ID: pm.ChannelID, Summoners: pm.FollowedPlayers}); err != nil {
		fmt.Println("Unable to save channel data")
	}
}

func (pm *PlayerMonitor) monitorPlayers() {
	for {
		select {
		case msg := <-pm.messageChan:
			switch msg.kind {
			case AddPlayer:
				pm.FollowedPlayers[strconv.Itoa(msg.player.ID)] = msg.player
				pm.save()
				fmt.Printf("Added player '%v' to channel '%v'\n", msg.player.Name, pm.ChannelID)
			case RemovePlayer:
				pm.removePlayer(msg.summonerNames, msg.sender, msg.timeStr)
				fmt.Printf("'%v' from channel: %v\n", msg.summonerNames, pm.ChannelID)
			case ListPlayers:
				pm.listPlayers(msg.sender, msg.timeStr)
				fmt.Printf("list players for channel: %v\n", pm.ChannelID)
			}

		case <-time.After(time.Second * 60):
			for _, p := range pm.FollowedPlayers {
				pm.handleMonitorPlayer(RC, p)
			}
			pm.reportGames()
		}
	}
}

func (pm *PlayerMonitor) removePlayer(summonerNames, sender, timeStr string) {
	names := strings.Fields(summonerNames)
	if len(names) < 2 {
		DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, newErrorMessage("Give at least one summoner name", sender, timeStr))
		return
	}

	var players []Player
	for _, name := range names[1:] {
		var found bool
		for _, p := range pm.FollowedPlayers {
			if strings.ToLower(p.Name) == strings.ToLower(name) {
				players = append(players, p)
				found = true
				break
			}
		}
		if !found {
			DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, newErrorMessage(fmt.Sprintf("Unable t' find summoner: %v", name), sender, timeStr))
		}
	}
	var removedSummoners []*discordgo.MessageEmbedField
	for _, p := range players {
		delete(pm.FollowedPlayers, strconv.Itoa(p.ID))
		removedSummoners = append(removedSummoners, &discordgo.MessageEmbedField{Name: p.Name, Value: p.Rank})
	}
	if len(removedSummoners) > 0 {
		DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &discordgo.MessageSend{
			Embed: newAddedSummonersMessage("Stopped followin'", sender, timeStr, removedSummoners),
		})
	}
	pm.save()
}

func (pm *PlayerMonitor) listPlayers(sender, timeStr string) {
	var mefs []*discordgo.MessageEmbedField
	for _, p := range pm.FollowedPlayers {
		mefs = append(mefs, &discordgo.MessageEmbedField{Name: p.Name, Value: p.Rank})
	}
	msg := discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:  "Followed summoners",
			Color:  green,
			Fields: mefs,
			Footer: newFooter(sender, timeStr),
		},
	}
	DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &msg)
}

func (pm *PlayerMonitor) handleMonitorPlayer(c *riotapi.Client, p Player) {
	cgi, err := c.Spectator.ActiveGamesBySummoner(p.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if cgi == nil {
		if p.CurrentGameID > 0 {
			pm.endGame(pm.games[p.CurrentGameID])
			p.CurrentGameID = 0
		}
		return
	}

	if p.CurrentGameID > 0 {
		return
	}

	p.CurrentGameID = cgi.GameID
	pm.games[cgi.GameID] = cgi

}

func (pm *PlayerMonitor) endGame(g *riotapi.CurrentGameInfo) {
	if g == nil {
		return
	}
	fmt.Println("Peli loppui")
	// sendMessage(newEmbedMessage("Peli loppui"))
	delete(pm.games, g.GameID)
	delete(pm.reportedGames, g.GameID)
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
