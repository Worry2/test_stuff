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
