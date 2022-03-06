package dto

import "github.com/bwmarrin/discordgo"

type Music struct {
	ID        string
	URL       string
	Title     string
	ChannelID string
	Requester *discordgo.User
	Session   *discordgo.Session
}
