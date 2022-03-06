package handlers

import (
	"fmt"
	"github.com/jonas747/dca"
	"github.com/qwerty22121998/discord_bot/dto"
	"github.com/qwerty22121998/discord_bot/parser"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type musicControl struct {
	add     chan *dto.Music
	skip    chan bool
	next    chan bool
	current chan *dto.Music
	shuffle chan bool
}

func newControl() *musicControl {
	return &musicControl{
		add:     make(chan *dto.Music),
		skip:    make(chan bool),
		next:    make(chan bool),
		current: make(chan *dto.Music),
		shuffle: make(chan bool),
	}
}

func (h *MusicHandler) Start() {
	c := h.control
	op := dca.StdEncodeOptions
	op.RawOutput = true
	op.Bitrate = 96
	op.Application = "lowdelay"

	queue := make([]*dto.Music, 0)
	skip := make(chan bool)
	var playing *dto.Music
	var muPlaying sync.Mutex
	var muListChange sync.Mutex

	go func() {
		for {
			select {
			case <-c.shuffle:
				func() {
					zap.S().Infow("shuffle signal")
					muListChange.Lock()
					defer muListChange.Unlock()
					rand.Shuffle(len(queue), func(i, j int) {
						queue[i], queue[j] = queue[j], queue[i]
					})
				}()
			}
		}
	}()

	go func() {
		for {
			select {
			case music := <-c.add:
				func() {
					muListChange.Lock()
					zap.S().Infow("add signal", "title", music.Title, "url", music.URL)
					queue = append(queue, music)
					muListChange.Unlock()
					c.next <- true
				}()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-c.next:
				func() {
					muListChange.Lock()
					zap.S().Infow("next signal")
					if len(queue) == 0 {
						muListChange.Unlock()
						return
					}
					muPlaying.Lock()
					playing = queue[0]
					queue = queue[1:]
					muPlaying.Unlock()
					muListChange.Unlock()
					c.current <- playing
				}()
			}
		}
	}()

	go func() {
		for {
			select {
			case music := <-c.current:
				func() {
					zap.S().Infow("current signal", "title", music.Title, "url", music.Title)
					file, err := parser.GetMusic(music.URL)
					if err != nil {
						zap.S().Error("error when get music", "err", err)
						return
					}
					sess, err := dca.EncodeMem(file, op)
					if err != nil {
						zap.S().Error("error when decode", "err", err)
						return
					}
					sig := make(chan error)
					stream := dca.NewStream(sess, h.voice, sig)
					defer stream.SetPaused(false)
					defer sess.Cleanup()
					h.message(music.Session, music.ChannelID, ":headphones: Bài hát hiện tại",
						fmt.Sprintf("**%v** theo yêu cầu của `%v`", music.Title, music.Requester.Username),
					)
					for {
						select {
						case err := <-sig:
							if err != io.EOF {
								zap.S().Errorw("error while streaming music", "error", err)
								return
							}
							zap.S().Infow("song ended", "name", music.Title, "url", music.URL)
							return
						case <-skip:

							stream.SetPaused(true)
							zap.S().Infow("skip signal in goroutine")
							zap.S().Infow("song skipped", "name", music.Title, "url", music.URL)
							return
						}
					}
				}()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-c.skip:
				func() {
					zap.S().Infow("skip signal")
					muPlaying.Lock()
					if playing == nil {
						muPlaying.Unlock()
						return
					}
					playing = nil
					muPlaying.Unlock()
					skip <- true
				}()
			}
		}

	}()

}
