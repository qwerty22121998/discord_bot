package handlers

import (
	discord "github.com/bwmarrin/discordgo"
)

func (h *MusicHandler) handleSkip(s *discord.Session, q string, m *discord.Message) {
	h.message(s, m.ChannelID, "", ":middle_finger: next bài")
	h.control.skip <- true
}
