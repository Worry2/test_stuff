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

func imageToDiscord(s string) {
	dm := discordgo.WebhookParams{
		AvatarURL: avatarURL,
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: s,
				Image: &discordgo.MessageEmbedImage{URL: "http://ddragon.leagueoflegends.com/cdn/img/champion/tiles/" + strings.ToLower(s) + "_0.jpg"},
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
