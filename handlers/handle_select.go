package handlers

import (
	"fmt"
	discord "github.com/bwmarrin/discordgo"
	"strconv"
)

func (h *MusicHandler) handleSelect(s *discord.Session, q string, m *discord.Message) {
	order, err := strconv.ParseInt(q, 10, 64)
	if err != nil {
		message(s, m.ChannelID, "", fmt.Sprintf("`%v` gõ sai rồi, gà quá :smirk:", m.Author.Username))
		return
	}
	order--
	list := h.cache.Get(m.Author.ID)
	if len(list) == 0 {
		message(s, m.ChannelID, "", fmt.Sprintf("`%v` bình tĩnh, chọn nhạc trước đã :man_tipping_hand:", m.Author.Username))
		return
	}
	if order < 0 || order >= int64(len(list)) {
		message(s, m.ChannelID, "", fmt.Sprintf("`%v` chọn đúng số nào :man_tipping_hand:", m.Author.Username))
		return
	}
	h.cache.Clear(m.Author.ID)
	message(s, m.ChannelID, "", fmt.Sprintf("`%v` đã thêm `%v` vào playlist", m.Author.Username, list[order].Title))
	list[order].ChannelID = m.ChannelID
	list[order].Requester = m.Author
	list[order].Session = s
	list[order].Connect = h.voice
	h.addMusic(&list[order])
}
