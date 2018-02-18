package main

import (
	"log"
	"os"

	youtube "google.golang.org/api/youtube/v3"
)

var channelID string // Y2M_CHANNEL_ID
var dataDir string   // Y2M_DATA_DIR
var baseURL string   // Y2M_BASE_URL
var s *youtube.Service
var c *youtube.Channel

func main() {
	channelID = os.Getenv("Y2M_CHANNEL_ID")
	dataDir = os.Getenv("Y2M_DATA_DIR")
	baseURL = os.Getenv("Y2M_BASE_URL")

	log.Printf("Running preflight tasks")

	checkDocker()
	createDirs()
	prepareYoutubeChannel()

	startFetchLoop()
}
