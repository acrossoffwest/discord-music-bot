package main

import (
	"context"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"log"
	"net/url"
)

func PrepareToken(rawUrlOrToken string) string {
	parsedUrl, err := url.Parse(rawUrlOrToken)
	if err != nil {
		return rawUrlOrToken
	}
	queryParams, err := url.ParseQuery(parsedUrl.RawQuery)

	if err != nil || len(queryParams["list"]) == 0 {
		return rawUrlOrToken
	}

	return queryParams["list"][0]
}

func LoadPlaylist(rawUrl string) []YoutubeVideoInfo {
	token := PrepareToken(rawUrl)
	nextPageToken := ""
	var videos []YoutubeVideoInfo
	for {
		videosPage := LoadPlaylistPage(token, nextPageToken)
		if len(videosPage.Videos) == 0 {
			return videos
		}
		videos = append(videos, videosPage.Videos...)
		if videosPage.NextPageToken == "" {
			break
		}
		nextPageToken = videosPage.NextPageToken
	}
	return videos
}

func LoadPlaylistPage(playlistId string, nextPageToken string) YoutubeVideosListPageWrapper {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(o.YoutubeToken))
	if err != nil {
		log.Fatal("Error new service: ", err)
	}
	part := []string{"id", "snippet"}
	call := service.PlaylistItems.List(part)
	call.PlaylistId(playlistId)
	log.Println("---------------- Next Page --------------", nextPageToken)
	if nextPageToken != "" {
		call = call.PageToken(nextPageToken)
	}
	resp, err := call.Do()

	if err != nil {
		log.Fatal(err)
	}

	var videos []YoutubeVideoInfo
	for _, playlistItem := range resp.Items {
		youtubeVideoInfo := YoutubeVideoInfo{
			Id:    playlistItem.Snippet.ResourceId.VideoId,
			Title: playlistItem.Snippet.Title,
		}
		videos = append(videos, youtubeVideoInfo)
		log.Println("%v, (%v)\r\n", youtubeVideoInfo.Title, youtubeVideoInfo.Id)
	}

	return YoutubeVideosListPageWrapper{
		Videos:        videos,
		NextPageToken: resp.NextPageToken,
	}
}
