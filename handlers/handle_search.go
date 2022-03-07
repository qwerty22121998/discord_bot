package handlers

import (
	"fmt"
	discord "github.com/bwmarrin/discordgo"
	"github.com/qwerty22121998/discord_bot/parser"
	"strings"
)

func (h *MusicHandler) handleSearch(s *discord.Session, q string, m *discord.Message) {
	if err := h.checkVoice(s, m.GuildID, m.Author); err != nil {
		message(s, m.ChannelID, "Lỗi", err.Error())
		return
	}
	musics, err := parser.SearchMusic(q, 5)
	if err != nil {
		message(s, m.ChannelID, "Lỗi", err.Error())
		return
	}
	content := strings.Builder{}
	for i, music := range musics {
		content.WriteString(fmt.Sprintf("**%d, [%v](%v)**\n", i+1, music.Title, music.URL))
	}
	h.cache.Set(m.Author.ID, musics)
	message(s, m.ChannelID,
		fmt.Sprintf(":love_you_gesture: Kết quả cho `%v` bởi `%v`\nChọn bài bằng cách `/play <order>`", q, m.Author.Username),
		content.String(),
	)
}
