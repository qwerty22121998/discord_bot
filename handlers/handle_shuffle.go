package handlers

import (
	discord "github.com/bwmarrin/discordgo"
)

func (h *MusicHandler) handleShuffle(s *discord.Session, q string, m *discord.Message) {
	h.shuffle()
	message(s, m.ChannelID, "", ":dizzy: shuffle")
	h.handleList(s, q, m)
	//h.control.Shuffle()
}
