package handlers

import (
	discord "github.com/bwmarrin/discordgo"
	"github.com/qwerty22121998/discord_bot/dto"
)

func (h *MusicHandler) handleTest(s *discord.Session, q string, m *discord.Message) {
	if err := h.checkVoice(s, m.GuildID, m.Author); err != nil {
		panic(err)
	}
	h.checkVoice(s, m.GuildID, m.Author)
	h.control.Add(&dto.Music{
		ID:        "HtDzVSgjjEc",
		URL:       "https://www.youtube.com/watch?v=HtDzVSgjjEc",
		Title:     "3D 10 Second countdown with voice and sound effects",
		ChannelID: m.ChannelID,
		Requester: m.Author,
		Session:   s,
		Connect:   h.voice,
	})
}
