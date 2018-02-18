package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	youtube "google.golang.org/api/youtube/v3"
)

func checkDocker() {
	log.Printf("Checking docker runtime")

	cmd := exec.Command("docker", "run", "--rm", "hello-world")
	output, err := cmd.CombinedOutput()
	if err != nil {
		handleStartupError(
			errors.New(err.Error()+"\n"+string(output)),
			"Unable to get a working Docker environment")
	}
}

func createDirs() {
	log.Printf("Creating directory structure")

	err := os.MkdirAll(filepath.Join(dataDir, "public", "mp3"), 0755)
	handleStartupError(err, "Unalbe to create MP3 directory")
}

func prepareYoutubeChannel() {
	log.Printf("Check youtube connection")

	ctx := context.Background()
	b, err := ioutil.ReadFile(filepath.Join(dataDir, "client_secret.json"))
	handleStartupError(err, "Unable to read client secret file")

	cl, err := google.ConfigFromJSON(b, youtube.YoutubeReadonlyScope)
	handleStartupError(err, "Unable to parse client secret file to config")

	s, err = youtube.New(getClient(ctx, cl))
	handleStartupError(err, "Error creating YouTube client")

	call := s.Channels.List("id,snippet")
	call.Id(channelID)
	result, err := call.Do()
	handleStartupError(err, "")

	c = result.Items[0]
}
