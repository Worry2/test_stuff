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
	ggReport := reportTeam(b.RC, "Good guys", ggPlayers)
	bgPlayers := getPlayersOfTeam(getOpposingTeam(ggTeam), cgi)
	bgReport := reportTeam(b.RC, "Bad guys", bgPlayers)

	if ggTeam == teamRed {
		ggReport.Color = red
	} else {
		bgReport.Color = red
	}

	b.Discord.ChannelMessageSendComplex(b.ChannelID, &discordgo.MessageSend{Embed: ggReport})
	b.Discord.ChannelMessageSendComplex(b.ChannelID, &discordgo.MessageSend{Embed: bgReport})

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
