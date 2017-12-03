package main

import "testing"
import "github.com/bwmarrin/discordgo"

func TestSend(t *testing.T) {

	m := discordgo.MessageEmbed{Type: "testii", Title: "Joo", Color: 161,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{Name: "Embed", Value: "Value", Inline: false},
			&discordgo.MessageEmbedField{Name: "Embed2", Value: "Value2", Inline: true},
			&discordgo.MessageEmbedField{Name: "Embed3", Value: "Value3", Inline: true},
		},
		Footer:   &discordgo.MessageEmbedFooter{Text: "Testi", IconURL: avatarURL},
		Author:   &discordgo.MessageEmbedAuthor{Name: "Pasi", URL: avatarURL},
		Provider: &discordgo.MessageEmbedProvider{Name: "Miia", URL: avatarURL},
	}
	sendMessage(&m)
}
