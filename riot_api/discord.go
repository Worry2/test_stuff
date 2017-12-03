package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	discordHook = "https://discordapp.com/api/webhooks/384835834815447052/K6amIwt30YVjWBJFivZIR8UIBB8Qh-mUGcleVUQ0oTSTt5BJuR0eXRKZ1xJyqEmEzscF"
	avatarURL   = "http://ddragon.leagueoflegends.com/cdn/img/champion/tiles/gangplank_0.jpg"
)

func sendToDiscord(s string) {
	fmt.Println("Lähetys: ", s)
	fmt.Println(s)
	dm := discordgo.WebhookParams{
		Content:   fmt.Sprintf("%s", s),
		AvatarURL: avatarURL,
	}

	b, err := json.Marshal(dm)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(discordHook, "application/json", bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func printAndReadResponse(r *http.Response) string {
	s, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("status: %d, body: %s\n", r.StatusCode, s)
	return string(s)
}

func imageToDiscord(title, desc, champ string) {
	c := strings.Replace(strings.ToLower(champ), " ", "", -1)
	url := "http://ddragon.leagueoflegends.com/cdn/img/champion/tiles/" + c + "_0.jpg"
	fmt.Println(url)
	dm := discordgo.WebhookParams{
		AvatarURL: avatarURL,
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       title,
				Description: desc,
				Image: &discordgo.MessageEmbedImage{
					URL: url,
				},
			},
		},
	}

	b, err := json.Marshal(dm)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(discordHook, "application/json", bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
