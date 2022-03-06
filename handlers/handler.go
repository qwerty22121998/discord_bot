package handlers

import (
	"fmt"
	discord "github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"strings"
	"sync"
)

const CMDPrefix = "."

var (
	ErrNoVoiceChannel = func(name string) error {
		return fmt.Errorf("Join voice trước đã bạn `%v` ơi", name)
	}
)

type MusicHandler struct {
	mu      sync.Mutex
	cache   *selectCache
	voice   *discord.VoiceConnection
	control *musicControl
}

func NewMusicHandler() *MusicHandler {
	return &MusicHandler{
		mu:      sync.Mutex{},
		cache:   newCache(),
		control: newControl(),
	}
}

func (h *MusicHandler) Join(s *discord.Session, gid, cid string) error {
	voice, err := s.ChannelVoiceJoin(gid, cid, false, true)
	if err != nil {
		return err
	}
	h.voice = voice
	return nil
}

func (h *MusicHandler) checkVoice(s *discord.Session, gid string, user *discord.User) error {
	if h.voice != nil {
		return nil
	}
	gid, cid, err := h.getCurrentUserVoiceChannel(s, gid, user)
	if err != nil {
		return err
	}
	if err := h.Join(s, gid, cid); err != nil {
		return err
	}
	return nil
}

func (h *MusicHandler) getCurrentUserVoiceChannel(s *discord.Session, gid string, user *discord.User) (string, string, error) {
	guild, err := s.State.Guild(gid)
	if err != nil {
		return "", "", err
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == user.ID {
			return vs.GuildID, vs.ChannelID, nil
		}
	}
	return "", "", ErrNoVoiceChannel(user.Username)
}

func (h *MusicHandler) message(s *discord.Session, cid, title, desc string) error {
	_, err := s.ChannelMessageSendComplex(cid, &discord.MessageSend{
		Embeds: []*discord.MessageEmbed{
			{
				Title:       title,
				Description: desc,
			},
		},
	})
	return err
}

func (h *MusicHandler) parseCommand(msg string) (string, string) {
	args := strings.SplitN(strings.TrimPrefix(msg, CMDPrefix), " ", 2)
	cmd := args[0]
	var query string
	if len(args) > 1 {
		query = args[1]
	}
	return cmd, query
}

func (h *MusicHandler) Handle(s *discord.Session, m *discord.MessageCreate) {
	if m.Author.Bot {
		return
	}
	if strings.HasPrefix(m.Content, CMDPrefix) {
		cmd, arg := h.parseCommand(m.Content)
		zap.S().Infow("receive command", "user", m.Author.Username, "command", cmd, "arg", arg)
		switch cmd {
		case "search":
			h.handleSearch(s, arg, m.Message)
		case "play":
			h.handleSelect(s, arg, m.Message)
		case "skip":
			h.handleSkip(s, arg, m.Message)
		case "shuffle":
			h.handleShuffle(s, arg, m.Message)
			//case "test":
			//	h.handleTest(s, arg, m.Message)
			//}
		}
	}
}
