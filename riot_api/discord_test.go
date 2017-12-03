package main

import "testing"
import "github.com/bwmarrin/discordgo"

func TestSend(t *testing.T) {

	m := discordgo.MessageEmbed{Type: "testii", Title: "OMG! Kuka vetäs ässän?", Color: 13125190,
		// Fields: []*discordgo.MessageEmbedField{
		// 	&discordgo.MessageEmbedField{Name: "Embed", Value: "Value", Inline: false},
		// 	&discordgo.MessageEmbedField{Name: "Embed2", Value: "Value2", Inline: true},
		// 	&discordgo.MessageEmbedField{Name: "Embed3", Value: "Value3", Inline: true},
		// },
		Footer: &discordgo.MessageEmbedFooter{Text: "Cpt. Selviö", IconURL: avatarURL},
	}

	sendMessages([]*discordgo.MessageEmbed{&m, &m})
}
