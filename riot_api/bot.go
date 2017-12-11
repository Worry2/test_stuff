package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tahkapaa/test_stuff/riot_api/riotapi"
)

// Bot is a discordBot
type Bot struct {
	Discord *discordgo.Session
	RC      *riotapi.Client

	ChannelID       string
	FollowedPlayers map[int]Player
}

func newBot(botToken, riotKey string) (*Bot, error) {
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, err
	}
	c, err := riotapi.New(riotKey, "eune", 50, 20)
	if err != nil {
		log.Fatalf("unable to initialize riot api: %v", err)
	}
	bot := Bot{Discord: dg, RC: c, FollowedPlayers: make(map[int]Player)}

	bot.AddMessageHandler()

	err = dg.Open()
	if err != nil {
		return nil, err
	}

	// bot.FollowedPlayers = []*Player{
	// 	{Name: "Uxipaxa", ID: 24749077},
	// 	{Name: "Invataxi", ID: 31507600},
	// 	{Name: "Ignusnus", ID: 25251553},
	// 	{Name: "Opettaja", ID: 28490422},
	// }

	return &bot, nil
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

		if m.Content == "!help" {
			msg := discordgo.MessageSend{
				Embed: &discordgo.MessageEmbed{
					Title: "Available commands",
					Color: green,
					Fields: []*discordgo.MessageEmbedField{
						{Name: "!follow", Value: "List o' summoner names t' be followed"},
						{Name: "!list", Value: "List o' summoners that are bein' followed"},
						{Name: "!remove", Value: "List o' summoner names that should nah be followed"},
						{Name: "!joke", Value: "Wants t' hear a joke?"},
					},
					Footer: newFooter(m.Author.Username, timeStr),
				},
			}
			s.ChannelMessageSendComplex(m.ChannelID, &msg)
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

		// Start following players
		if strings.HasPrefix(m.Content, "!follow") {
			st, err := s.ChannelMessageSendComplex(m.ChannelID, newWorkingMessage())
			if err != nil {
				fmt.Println(err)
				return
			}
			names := strings.Fields(m.Content)
			if len(names) < 2 {
				s.ChannelMessageSendComplex(m.ChannelID, newErrorMessage("Give at least one summoner name", m.Author.Username, timeStr))
			}
			var addedSummoners []*discordgo.MessageEmbedField
			for _, name := range names[1:] {
				summoner, err := b.RC.Summoner.SummonerByName(name)
				if err != nil || summoner == nil {
					s.ChannelMessageSendComplex(m.ChannelID, newErrorMessage(fmt.Sprintf("Unable t' find summoner: %v", name), m.Author.Username, timeStr))
					continue
				}
				rank, err := findPlayerRank(b.RC, summoner.ID)
				if err != nil {
					fmt.Println(err)
				}
				addedSummoners = append(addedSummoners, &discordgo.MessageEmbedField{Name: summoner.Name, Value: rank})
				b.FollowedPlayers[summoner.ID] = Player{Name: summoner.Name, ID: summoner.ID, Rank: rank}
				b.ChannelID = m.ChannelID
			}
			if len(addedSummoners) > 0 {
				s.ChannelMessageEditComplex(&discordgo.MessageEdit{
					Channel: m.ChannelID,
					ID:      st.ID,
					Embed:   newAddedSummonersMessage("Now followin'", m.Author.Username, timeStr, addedSummoners),
				})
			}
		}

		// Stop following players
		if strings.HasPrefix(m.Content, "!remove") {
			names := strings.Fields(m.Content)
			if len(names) < 2 {
				s.ChannelMessageSendComplex(m.ChannelID, newErrorMessage("Give at least one summoner name", m.Author.Username, timeStr))
			}
			var players []Player
			for _, name := range names[1:] {
				for _, p := range b.FollowedPlayers {
					if strings.ToLower(p.Name) == strings.ToLower(name) {
						players = append(players, p)
						break
					}
					s.ChannelMessageSendComplex(m.ChannelID, newErrorMessage(fmt.Sprintf("Unable t' find summoner: %v", name), m.Author.Username, timeStr))
				}
			}
			var removedSummoners []*discordgo.MessageEmbedField
			for _, p := range players {
				delete(b.FollowedPlayers, p.ID)
				removedSummoners = append(removedSummoners, &discordgo.MessageEmbedField{Name: p.Name, Value: p.Rank})
			}
			if len(removedSummoners) > 0 {
				s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
					Embed: newAddedSummonersMessage("Stopped followin'", m.Author.Username, timeStr, removedSummoners),
				})
			}
		}

		// list followed players
		if m.Content == "!list" {
			var mefs []*discordgo.MessageEmbedField
			for _, p := range b.FollowedPlayers {
				mefs = append(mefs, &discordgo.MessageEmbedField{Name: p.Name, Value: p.Rank})
			}
			msg := discordgo.MessageSend{
				Embed: &discordgo.MessageEmbed{
					Title:  "Followed summoners",
					Color:  green,
					Fields: mefs,
					Footer: newFooter(m.Author.Username, timeStr),
				},
			}
			s.ChannelMessageSendComplex(m.ChannelID, &msg)
		}

		if m.Content == "!joke" {
			s.ChannelMessageSend(m.ChannelID, jokes[rand.Intn(len(jokes))])
		}
	})
}

func (b *Bot) monitorPlayers() {
	for {
		for _, p := range b.FollowedPlayers {
			handleMonitorPlayer(b.RC, p)
		}
		b.reportGames()
		time.Sleep(time.Second * 60)
	}
}
