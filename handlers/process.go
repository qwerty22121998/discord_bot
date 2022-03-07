package handlers

import (
	"fmt"
	"github.com/qwerty22121998/dca"
	"github.com/qwerty22121998/discord_bot/dto"
	"github.com/qwerty22121998/discord_bot/parser"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"sync"
	"time"
)

var op *dca.EncodeOptions

func init() {
	rand.Seed(time.Now().UnixNano())
	op = dca.StdEncodeOptions
	op.RawOutput = true
	op.Bitrate = 96
	op.Application = "lowdelay"
}

type musicControl struct {
	skip      chan bool
	muQueue   sync.Mutex
	queue     []*dto.Music
	muPlaying sync.Mutex
	playing   chan *dto.Music
}

func newControl() *musicControl {
	return &musicControl{
		skip:    make(chan bool),
		queue:   make([]*dto.Music, 0),
		playing: make(chan *dto.Music),
	}
}

func (h *musicControl) Add(music *dto.Music) {
	zap.S().Infow("add", "title", music.Title, "id", music.ID)
	defer zap.S().Infow("added")
	h.muQueue.Lock()
	defer h.muQueue.Unlock()
	h.queue = append(h.queue, music)
}

func (h *musicControl) Shuffle() {
	zap.S().Infow("shuffle")
	defer zap.S().Infow("shuffled")
	h.muQueue.Lock()
	defer h.muQueue.Unlock()
	rand.Shuffle(len(h.queue), func(i, j int) {
		h.queue[i], h.queue[j] = h.queue[j], h.queue[i]
	})
}

func (h *musicControl) Play(music *dto.Music) {
	zap.S().Infow("play", "title", music.Title, "id", music.ID)
	defer zap.S().Infow("end", "title", music.Title, "id", music.ID)
	h.muPlaying.Lock()
	defer func() {
		h.playing = nil
		defer h.muPlaying.Unlock()
	}()

	file, err := parser.GetMusic(music.URL)
	if err != nil {
		zap.S().Error("parser.GetMusic", "err", err)
		return
	}
	sess, err := dca.EncodeMem(file, op)
	if err != nil {
		zap.S().Error("error when decode", "err", err)
		return
	}
	sig := make(chan error)
	stream := dca.NewStream(sess, music.Connect, sig)
	defer stream.SetPaused(false)
	defer sess.Cleanup()
	message(music.Session, music.ChannelID, ":headphones: Bài hát hiện tại",
		fmt.Sprintf("**%v** theo yêu cầu của `%v`", music.Title, music.Requester.Username),
	)
	for {
		select {
		case err := <-sig:
			if err != io.EOF {
				zap.S().Errorw("error while playing", "error", err)
				return
			}
			return
		case <-h.skip:
			stream.SetPaused(true)
			zap.S().Infow("skipped", "name", music.Title, "url", music.URL)
			return
		}
	}
}

func (h *musicControl) Skip() {
	h.skip <- true
}

func (h *musicControl) Start() {
	for music := range h.playing {
		h.Play(music)
	}
}
