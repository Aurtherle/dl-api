package api

import (
	"encoding/json"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	Socks5Proxy   = os.Getenv("SOCKS5_PROXY")
	SourceCodeURL = "https://github.com/Abishnoi69/ytdl-api"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	infoMsg := "This is a simple API to download videos from YouTube and Instagram.\n" +
		"Usage:\n" +
		"/yt?url=<video_url> - download video from YouTube\n" +
		"Instagram: /ig?url=<post_id> - get download video URL from Instagram\n" +
		"Made with ❤️ by Abishnoi69\n" +
		"Source code: " + SourceCodeURL + "\n"

	switch r.URL.Path {
	case "/":
		if Socks5Proxy == "" {
			infoMsg += "NOTE: No SOCKS5 proxy configured; you might get rate limited by YouTube :("
		}

		_, err := fmt.Fprint(w, infoMsg)
		if err != nil {
			http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
			return
		}

	case "/yt":
		video, err := handleYouTube(w, r)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(video); err != nil {
			http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
		}

	default:
		http.NotFound(w, r)
	}
}

func handleYouTube(w http.ResponseWriter, r *http.Request) (map[string]string, error) {
	videoURL := r.URL.Query().Get("url")
	if videoURL == "" {
		return nil, fmt.Errorf("please provide a video URL\nUsage: /yt?url=<video_url>")
	}

	ytClient := youtube.Client{}
	if Socks5Proxy != "" {
		proxyURL, _ := url.Parse(Socks5Proxy)
		ytClient = youtube.Client{HTTPClient: &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}}
	}

	video, err := ytClient.GetVideo(videoURL)
	if err != nil {
		return nil, fmt.Errorf("error retrieving video: %v", err)
	}

	streamURL, err := ytClient.GetStreamURL(video, &video.Formats[0])
	if err != nil {
		return nil, fmt.Errorf("error generating stream URL: %v", err)
	}

	return map[string]string{
		"ID":          video.ID,
		"author":      video.Author,
		"duration":    video.Duration.String(),
		"thumbnail":   video.Thumbnails[0].URL,
		"description": video.Description,
		"stream_url":  streamURL,
		"title":       video.Title,
		"view_count":  fmt.Sprintf("%d", video.Views),
	}, nil
}
