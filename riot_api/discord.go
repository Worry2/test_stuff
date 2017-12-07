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

const (
	red    = 16750480
	blue   = 2926540
	green  = 5308359
	yellow = 16773522
)

func sendToDiscord(s string) {
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

func newWorkingMessage() *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title: "Runnin' on some errands...",
			Color: yellow,
		},
	}
}

func newEmbedMessage(title string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: title,
		Color: blue,
	}
}

func newSuccessMessage(title, requestor, msgTime string) *discordgo.MessageEmbed {
	return newMessage(title, requestor, msgTime, green)
}

func newErrorMessage(title, requestor, msgTime string) *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Embed: newMessage(title, requestor, msgTime, red),
	}
}

func newMessage(title, requestor, msgTime string, color int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:  title,
		Color:  color,
		Footer: newFooter(requestor, msgTime),
	}
}

func newFooter(requestor, msgTime string) *discordgo.MessageEmbedFooter {
	return &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Requested by: %s | %v", requestor, msgTime),
	}
}

func newAddedSummonersMessage(title, requestor, msgTime string, fields []*discordgo.MessageEmbedField) *discordgo.MessageEmbed {
	msgEmbed := newSuccessMessage(title, requestor, msgTime)
	msgEmbed.Fields = fields
	return msgEmbed
}

func sendMessages(embeds []*discordgo.MessageEmbed) {
	dm := discordgo.WebhookParams{
		AvatarURL: avatarURL,
		Embeds:    embeds,
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

func sendMessage(m *discordgo.MessageEmbed) {
	dm := discordgo.WebhookParams{
		AvatarURL: avatarURL,
		Embeds:    []*discordgo.MessageEmbed{m},
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
