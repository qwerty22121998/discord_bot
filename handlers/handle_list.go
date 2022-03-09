package handlers

import (
	"fmt"
	discord "github.com/bwmarrin/discordgo"
	"strings"
)

func (h *MusicHandler) handleList(s *discord.Session, q string, m *discord.Message) {
	musics := h.getQueue()
	playing := h.getPlaying()
	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("**Hiện tại, [%v](%v)** - `%v`\n", playing.Title, playing.URL, playing.Requester.Username))

	if len(musics) != 0 {
		msg.WriteString("\nTíp theo: \n")
		for i, music := range musics {
			msg.WriteString(fmt.Sprintf("**%d, [%v](%v)** - `%v`\n", i+1, music.Title, music.URL, music.Requester.Username))
		}
	}
	message(s, m.ChannelID, ":women_with_bunny_ears_partying: Danh sách", msg.String())
}
