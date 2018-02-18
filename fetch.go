package main

import (
	"crypto/sha1"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"
)

func startFetchLoop() {
	for {
		log.Printf("Start to process")
		sleepTime := 15 * time.Minute

		ch := rssChannel{
			Title:       c.Snippet.Title,
			Author:      c.Snippet.CustomUrl,
			Description: c.Snippet.Description,
			Image: rssImage{
				URL:   c.Snippet.Thumbnails.High.Url,
				Link:  "https://www.youtube.com/channel/" + channelID,
				Title: c.Snippet.Title,
			},
			Language: "en-us",
			Link:     "https://www.youtube.com/channel/" + channelID,
			Items:    []rssItem{},
		}

		call := s.Search.List("id,snippet")
		call.ChannelId(channelID).MaxResults(10).Order("date").Type("video")
		response, err := call.Do()
		if err != nil {
			log.Printf("Unable to get channel detail: %v", err)
			sleepTime = 60 * time.Minute
			continue
		}

		mp3Files := map[string]bool{}

		for _, item := range response.Items {

			call2 := s.Videos.List("id,snippet,contentDetails")
			call2.Id(item.Id.VideoId)
			response2, err := call2.Do()
			if err != nil {
				log.Printf("Unable to get video detail: %v", err)
				sleepTime = 60 * time.Minute
				continue
			}

			video := response2.Items[0]

			fn, length, err := processVideo(video.Id)
			if err != nil {
				log.Printf("Unable to process video: %v", err)
				sleepTime = 60 * time.Minute
				continue
			}

			it := rssItem{
				Title:       video.Snippet.Title,
				Description: video.Snippet.Description,
				PubDate:     formatPubDate(video.Snippet.PublishedAt),
				Enclosure: rssEnclosure{
					URL:    baseURL + "/" + fn,
					Type:   "audio/mpeg",
					Length: length,
				},
				Duration: formatDuration(video.ContentDetails.Duration),
				GUID:     item.Id.VideoId,
			}
			ch.Items = append(ch.Items, it)

			mp3Files[fn] = true
		}

		saveFeedXML(ch)
		cleanUpFiles(mp3Files)

		log.Printf("Sleeping for %d minutes", sleepTime/time.Minute)
		time.Sleep(sleepTime)
	}
}

func processVideo(videoID string) (string, int64, error) {
	log.Printf("Processing Video ID: %s", videoID)

	h := sha1.New()
	io.WriteString(h, videoID)
	fn := fmt.Sprintf("%x.mp3", h.Sum(nil))
	fullFn := filepath.Join(dataDir, "public", "mp3", fn)

	info, err := os.Stat(fullFn)
	if !os.IsNotExist(err) {
		return fn, info.Size(), nil
	}

	user, err := user.Current()
	if err != nil {
		return "", 0, err
	}
	cmd := exec.Command(
		"docker",
		"run",
		"--rm",
		"-e", "DOWNLOAD_URI=https://www.youtube.com/watch?v="+videoID,
		"-v", "/etc/group:/etc/group:ro",
		"-v", "/etc/passwd:/etc/passwd:ro",
		"-v", filepath.Join(dataDir, "public", "mp3")+":"+"/data",
		"-u", user.Uid,
		"xinsnake/youtube2mp3",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Docker command failed")
		return "", 0, errors.New(err.Error() + "\n" + string(output))
	}
	err = os.Rename(filepath.Join(dataDir, "public", "mp3", "output.mp3"), fullFn)
	log.Printf("File saved to %s", fullFn)

	info2, err := os.Stat(fullFn)
	if os.IsNotExist(err) {
		return fn, 0, err
	}
	return fn, info2.Size(), nil
}

func saveFeedXML(ch rssChannel) {
	rr := rssRoot{
		Version:     "2.0",
		XmlnsItunes: "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Channel:     ch,
	}
	b, err := xml.MarshalIndent(rr, "", "  ")
	if err != nil {
		log.Printf("Unable to marshal XML: %v", err)
		return
	}
	finalXML := xml.Header + string(b)
	err = ioutil.WriteFile(filepath.Join(dataDir, "public", "feed.xml"), []byte(finalXML), 0644)
	if err != nil {
		log.Printf("Unable to write feed XML: %v", err)
	}
	log.Printf("Feed file saved to %s", filepath.Join(dataDir, "public", "feed.xml"))
}

func cleanUpFiles(mp3Files map[string]bool) {
	dataPath := filepath.Join(dataDir, "public", "mp3")
	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		log.Printf("Unable to read directory to clean: %v", err)
	}
	for _, file := range files {
		if mp3Files[file.Name()] {
			continue
		}
		log.Printf("Removing unused MP3 file %s", file.Name())
		err = os.Remove(filepath.Join(dataDir, "public", "mp3", file.Name()))
		if err != nil {
			log.Printf("Unable to remove MP3 file: %v", err)
		}
	}
}
