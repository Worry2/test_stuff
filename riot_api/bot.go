package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tahkapaa/test_stuff/riot_api/riotapi"
)

// Bot is a discordBot
type Bot struct {
	Discord  *discordgo.Session
	Channels map[string]*Channel
	db       DB
}

// Channel is a channel with followed players
type Channel struct {
	ChannelID string
	monitor   *PlayerMonitor
}

func newBot(botToken string, db DB) (*Bot, error) {
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, err
	}
	bot := Bot{
		Discord:  dg,
		Channels: make(map[string]*Channel),
		db:       db,
	}

	bot.AddMessageHandler()

	err = dg.Open()
	if err != nil {
		return nil, err
	}

	bot.readChannelsFromDB()
	return &bot, nil
}

func (b *Bot) readChannelsFromDB() {
	channels, err := b.db.Get()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize channels from db: %v", err))
	}
	for _, channel := range channels {
		b.addChannel(channel.ID, channel.Summoners)
	}
}

func (b *Bot) addChannel(ID string, players map[string]Player) {
	if players == nil {
		players = make(map[string]Player)
	}
	pm := PlayerMonitor{
		FollowedPlayers: players,
		games:           make(map[int]*riotapi.CurrentGameInfo),
		reportedGames:   make(map[int]bool),
		messageChan:     make(chan monitorMessage, 1),
		ChannelID:       ID,
		db:              b.db,
	}

	go pm.monitorPlayers()

	b.Channels[ID] = &Channel{
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
				b.addChannel(m.ChannelID, nil)
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
		DGBot.Discord.ChannelMessageSendComplex(channelID, newErrorMessage("No one be bein' followed", sender, timeStr))
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
		DGBot.Discord.ChannelMessageSendComplex(channelID, newErrorMessage("No one be bein' followed", sender, timeStr))
		return
	}

	c.monitor.messageChan <- monitorMessage{
		kind:    ListPlayers,
		sender:  sender,
		timeStr: timeStr,
	}
}
