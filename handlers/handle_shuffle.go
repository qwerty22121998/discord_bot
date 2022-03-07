package handlers

import (
	discord "github.com/bwmarrin/discordgo"
)

func (h *MusicHandler) handleShuffle(s *discord.Session, q string, m *discord.Message) {
	message(s, m.ChannelID, "", ":dizzy: shuffle")
	h.control.Shuffle()
}
