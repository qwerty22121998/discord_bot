package handlers

import (
	discord "github.com/bwmarrin/discordgo"
)

func (h *MusicHandler) handleSkip(s *discord.Session, q string, m *discord.Message) {
	if err := h.skip(); err != nil {
		message(s, m.ChannelID, "", err.Error())
		return
	}
	message(s, m.ChannelID, "", ":middle_finger: next bài")
}
