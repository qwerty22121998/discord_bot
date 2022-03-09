package parser

import (
	"context"
	"fmt"
	ytdl "github.com/kkdai/youtube/v2"
	"github.com/qwerty22121998/discord_bot/dto"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"html"
	"io"
)

var yt *youtube.Service
var dlClient *ytdl.Client

func InitParser(APIKey string) error {
	var err error
	yt, err = youtube.NewService(context.Background(), option.WithAPIKey(APIKey))
	if err != nil {
		return err
	}
	dlClient = &ytdl.Client{
		Debug: true,
	}
	return nil
}

func SearchMusic(kw string, maxSize int64) ([]dto.Music, error) {
	req := yt.Search.List([]string{"id", "snippet"}).Q(kw).MaxResults(maxSize)
	resp, err := req.Do()
	if err != nil {
		return nil, err
	}
	res := make([]dto.Music, 0, maxSize)
	for _, item := range resp.Items {
		res = append(res, dto.Music{
			URL:   fmt.Sprintf("https://www.youtube.com/watch?v=%v", item.Id.VideoId),
			Title: html.UnescapeString(item.Snippet.Title),
			ID:    item.Id.VideoId,
		})
	}
	return res, nil
}

func GetMusic(url string) (io.ReadCloser, error) {
	vid, err := dlClient.GetVideo(url)
	if err != nil {
		return nil, err
	}
	audioChannels := vid.Formats.WithAudioChannels()
	stream, _, err := dlClient.GetStream(vid, &audioChannels[0])
	if err != nil {
		return nil, err
	}
	return stream, nil
}
