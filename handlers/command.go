package handlers

import discord "github.com/bwmarrin/discordgo"

type Command interface {
	Handler() func(s *discord.Session, q string, m *discord.Message)
	Command() string
	Help() string
}

type command struct {
	handler func(s *discord.Session, q string, m *discord.Message)
	command string
	help    string
}

func (c *command) Handler() func(s *discord.Session, q string, m *discord.Message) {
	return c.handler
}

func (c *command) Help() string {
	return c.help
}

func (c *command) Command() string {
	return c.command
}
