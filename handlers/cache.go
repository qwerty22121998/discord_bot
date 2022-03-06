package handlers

import (
	"github.com/qwerty22121998/discord_bot/dto"
	"sync"
)

type selectCache struct {
	mu   sync.Mutex
	data map[string][]dto.Music
}

func newCache() *selectCache {
	return &selectCache{
		data: make(map[string][]dto.Music),
	}
}

func (s *selectCache) Set(uid string, list []dto.Music) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[uid] = list
}

func (s *selectCache) Get(uid string) []dto.Music {
	s.mu.Lock()
	defer s.mu.Unlock()
	list, ok := s.data[uid]
	if !ok {
		return nil
	}
	return list
}

func (s *selectCache) Clear(uid string) {
	delete(s.data, uid)
}
