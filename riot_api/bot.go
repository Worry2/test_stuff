package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tahkapaa/test_stuff/riot_api/riotapi"
)

// Bot is a discordBot
type Bot struct {
	Discord  *discordgo.Session
	Channels map[string]*Channel
}

// Channel is a channel with followed players
type Channel struct {
	ChannelID string
	monitor   *PlayerMonitor
}

// PlayerMonitor monitors players
type PlayerMonitor struct {
	fpMutex         sync.Mutex
	FollowedPlayers map[int]Player
	games           map[int]*riotapi.CurrentGameInfo
	reportedGames   map[int]bool
	messageChan     chan monitorMessage
	ChannelID       string
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

func newBot(botToken string) (*Bot, error) {
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, err
	}
	bot := Bot{Discord: dg, Channels: make(map[string]*Channel)}

	bot.AddMessageHandler()

	err = dg.Open()
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

func newChannel(ID string) *Channel {
	pm := PlayerMonitor{
		FollowedPlayers: make(map[int]Player),
		games:           make(map[int]*riotapi.CurrentGameInfo),
		reportedGames:   make(map[int]bool),
		messageChan:     make(chan monitorMessage, 1),
		ChannelID:       ID,
	}

	go pm.monitorPlayers()

	return &Channel{
		ChannelID: ID,
		monitor:   &pm,
	}
}

// AddMessageHandler adds CreateMessage handler to the bot
func (b *Bot) AddMessageHandler() {
	b.Discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		timeStr := time.Now().Format(time.ANSIC)
		// Ignore all messages created by the bot itself
		// This isn't required in this specific example but it's a good practice.
		if m.Author.ID == s.State.User.ID {
			return
		}

		// If the message is "ping" reply with "Pong!"
		if m.Content == "ping" {
			s.ChannelMessageSend(m.ChannelID, "Yarrr!")
		}

		// If the message is "pong" reply with "Ping!"
		if m.Content == "pong" {
			s.ChannelMessageSend(m.ChannelID, "Yarrr!")
		}

		if strings.Contains(strings.ToLower(m.Content), "kapteeni") || strings.Contains(strings.ToLower(m.Content), "kapu") {
			s.ChannelMessageSend(m.ChannelID, "Yarrr!")
		}

		if m.Content == "?help" {
			handleHelp(m.ChannelID, m.Author.Username, timeStr, s)
			return
		}

		// Start following players
		if strings.HasPrefix(m.Content, "?follow") {
			_, ok := b.Channels[m.ChannelID]
			if !ok {
				b.Channels[m.ChannelID] = newChannel(m.ChannelID)
			}
			b.Channels[m.ChannelID].handleStartFollowing(m.Content, m.Author.Username, timeStr, s)
			return
		}

		// Stop following players
		if strings.HasPrefix(m.Content, "?remove") {
			b.Channels[m.ChannelID].handleStopFollowing(m.ChannelID, m.Content, m.Author.Username, timeStr)
			return
		}

		// list followed players
		if m.Content == "?list" {
			b.Channels[m.ChannelID].handleListFollowedPlayers(m.ChannelID, m.Author.Username, timeStr)
			return
		}

		if m.Content == "?joke" {
			s.ChannelMessageSend(m.ChannelID, jokes[rand.Intn(len(jokes))])
			return
		}
	})
}

func handleHelp(channelID, sender, timeStr string, s *discordgo.Session) {
	msg := discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title: "Available commands",
			Color: green,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "?follow", Value: "List o' summoner names t' be followed"},
				{Name: "?list", Value: "List o' summoners that are bein' followed"},
				{Name: "?remove", Value: "List o' summoner names that should nah be followed"},
				{Name: "?joke", Value: "Wants t' hear a joke?"},
			},
			Footer: newFooter(sender, timeStr),
		},
	}
	s.ChannelMessageSendComplex(channelID, &msg)
}

func (c *Channel) handleStartFollowing(summonerNames, sender, timeStr string, s *discordgo.Session) {
	st, err := s.ChannelMessageSendComplex(c.ChannelID, newWorkingMessage())
	if err != nil {
		fmt.Println(err)
		return
	}
	names := strings.Fields(summonerNames)
	if len(names) < 2 {
		s.ChannelMessageSendComplex(c.ChannelID, newErrorMessage("Give at least one summoner name", sender, timeStr))
		return
	}
	var addedSummoners []*discordgo.MessageEmbedField
	for _, name := range names[1:] {
		summoner, err := RC.Summoner.SummonerByName(name)
		if err != nil || summoner == nil {
			s.ChannelMessageSendComplex(c.ChannelID, newErrorMessage(fmt.Sprintf("Unable t' find summoner: %v", name), sender, timeStr))
			continue
		}
		rank, err := findPlayerRank(RC, summoner.ID)
		if err != nil {
			fmt.Println(err)
		}
		addedSummoners = append(addedSummoners, &discordgo.MessageEmbedField{Name: summoner.Name, Value: rank})

		c.monitor.messageChan <- monitorMessage{
			kind:    AddPlayer,
			player:  Player{Name: summoner.Name, ID: summoner.ID, Rank: rank},
			sender:  sender,
			timeStr: timeStr,
		}
	}
	if len(addedSummoners) > 0 {
		s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel: c.ChannelID,
			ID:      st.ID,
			Embed:   newAddedSummonersMessage("Now followin'", sender, timeStr, addedSummoners),
		})
	}
}

func (c *Channel) handleStopFollowing(channelID, summonerNames, sender, timeStr string) {
	if c == nil {
		DGBot.Discord.ChannelMessageSendComplex(channelID, newErrorMessage("No one is being followed", sender, timeStr))
		return
	}

	c.monitor.messageChan <- monitorMessage{
		kind:          RemovePlayer,
		summonerNames: summonerNames,
		sender:        sender,
		timeStr:       timeStr,
	}
}

func (c *Channel) handleListFollowedPlayers(channelID, sender, timeStr string) {
	if c == nil {
		DGBot.Discord.ChannelMessageSendComplex(channelID, newErrorMessage("No one is being followed", sender, timeStr))
		return
	}

	c.monitor.messageChan <- monitorMessage{
		kind:    ListPlayers,
		sender:  sender,
		timeStr: timeStr,
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
		delete(pm.FollowedPlayers, p.ID)
		removedSummoners = append(removedSummoners, &discordgo.MessageEmbedField{Name: p.Name, Value: p.Rank})
	}
	if len(removedSummoners) > 0 {
		DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &discordgo.MessageSend{
			Embed: newAddedSummonersMessage("Stopped followin'", sender, timeStr, removedSummoners),
		})
	}
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

func (pm *PlayerMonitor) monitorPlayers() {
	for {
		select {
		case msg := <-pm.messageChan:
			switch msg.kind {
			case AddPlayer:
				pm.FollowedPlayers[msg.player.ID] = msg.player
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
