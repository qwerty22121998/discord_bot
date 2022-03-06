package handlers

import (
	discord "github.com/bwmarrin/discordgo"
)

func (h *MusicHandler) handleShuffle(s *discord.Session, q string, m *discord.Message) {
	h.message(s, m.ChannelID, "", ":dizzy:")
	h.control.shuffle <- true
}
