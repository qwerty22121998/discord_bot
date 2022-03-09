package handlers

import (
	"errors"
	"fmt"
	discord "github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/qwerty22121998/discord_bot/dto"
	"github.com/qwerty22121998/discord_bot/parser"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"strings"
	"sync"
)

const CMDPrefix = "."

var (
	ErrNoVoiceChannel = func(name string) error {
		return fmt.Errorf("Join voice trước đã bạn `%v` ơi", name)
	}
	ErrNoPlaying = errors.New(":hear_no_evil: Không có bài nào thì skip nyc à ")
)

type MusicHandler struct {
	cache *selectCache
	voice *discord.VoiceConnection

	queue     []*dto.Music
	muQue     sync.Mutex
	playing   *dto.Music
	muPlaying sync.Mutex
	chanPlay  chan bool
	chanSkip  chan bool
	chanDone  chan bool
}

func NewMusicHandler() *MusicHandler {
	return &MusicHandler{
		cache: newCache(),
		voice: nil,

		queue:     make([]*dto.Music, 0),
		muQue:     sync.Mutex{},
		playing:   nil,
		muPlaying: sync.Mutex{},
		chanPlay:  make(chan bool, 1024),
		chanSkip:  make(chan bool),
		chanDone:  make(chan bool),
	}
}

func (h *MusicHandler) Start() {
	for {
		select {
		case <-h.chanPlay:
			h.play()
		}
	}
}

func (h *MusicHandler) pop() *dto.Music {
	h.muQue.Lock()
	h.muPlaying.Lock()
	defer h.muQue.Unlock()
	defer h.muPlaying.Unlock()
	playing := h.queue[0]
	h.queue = h.queue[1:]
	return playing
}

func (h *MusicHandler) play() {
	nextSong := h.pop()
	h.setPlaying(nextSong)
	defer func() {
		h.setPlaying(nil)
	}()

	reader, err := parser.GetMusic(nextSong.URL)
	if err != nil {
		zap.S().Errorw("error when get music", "id", nextSong.ID, "error", err)
		return
	}
	dcaOp := dca.StdEncodeOptions
	dcaOp.RawOutput = true
	dcaOp.Bitrate = 96
	dcaOp.Application = "lowdelay"
	opusReader, err := dca.EncodeMem(reader, dcaOp)
	if err != nil {
		zap.S().Errorw("error when create opus reader", "error", err)
		return
	}
	done := make(chan error)
	stream := dca.NewStream(opusReader, nextSong.Connect, done)
	defer func() {
		stream.SetPaused(true)
		opusReader.Cleanup()
		stream.SetPaused(false)
	}()
	for {
		select {
		case err := <-done:
			if err != io.EOF {
				zap.S().Errorw("error when playing", "error", err)
			}
			return
		case <-h.chanSkip:
			return
		}
	}

}

func (h *MusicHandler) getQueue() []*dto.Music {
	h.muQue.Lock()
	defer h.muQue.Unlock()
	return h.queue
}

func (h *MusicHandler) skip() error {
	playing := h.getPlaying()
	if playing == nil {
		return ErrNoPlaying
	}
	h.chanSkip <- true
	return nil
}

func (h *MusicHandler) shuffle() {
	h.muQue.Lock()
	defer h.muQue.Unlock()
	rand.Shuffle(len(h.queue), func(i, j int) {
		h.queue[i], h.queue[j] = h.queue[j], h.queue[i]
	})
}

func (h *MusicHandler) addMusic(music *dto.Music) {
	h.muQue.Lock()
	defer h.muQue.Unlock()
	h.queue = append(h.queue, music)
	h.chanPlay <- true
}

func (h *MusicHandler) getPlaying() *dto.Music {
	h.muPlaying.Lock()
	defer h.muPlaying.Unlock()
	return h.playing
}

func (h *MusicHandler) setPlaying(music *dto.Music) {
	h.muPlaying.Lock()
	defer h.muPlaying.Unlock()
	h.playing = music
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

func message(s *discord.Session, cid, title, desc string) error {
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
		case "queue":
			h.handleList(s, arg, m.Message)
			//case "test":
			//	h.handleTest(s, arg, m.Message)
			//}
		}
	}
}
