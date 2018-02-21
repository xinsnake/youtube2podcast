package main

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/xinsnake/youtube2podcast/config"
)

const youtubeChannelURI = "https://www.youtube.com/channel/"
const youtubeVideoURI = "https://www.youtube.com/watch?v="

func fireOffFetchLoop() {
	log.Printf("Fire off fetch loop")

	for index, ch := range cfg.Channels {
		log.Printf("Fire off channel %s", ch.ID)

		if index != 0 {
			rand.Seed(time.Now().UnixNano())
			delay := rand.Intn(10)
			log.Printf("Delay randomly for %d seconds", delay)
			time.Sleep(time.Duration(delay) * time.Second)
		}

		go fetchChannel(ch)
	}
}

func fetchChannel(ch config.Channel) {
	success := true

	for {
		if !success {
			log.Printf("Detecte previous failure for channel %s, waiting for 1 interval", ch.ID)
			time.Sleep(time.Duration(ch.RefreshInterval) * time.Second)
		}

		log.Printf("Start to process channel %s", ch.ID)

		if ch.Retain > 50 {
			log.Printf("Warning: channel %s retain greater than 50 is not supported yet, set to 50", ch.ID)
			ch.Retain = 50
		}

		chListCall := yService.Channels.List("id,snippet").Id(ch.ChannelID).MaxResults(1)
		chListResp, err := chListCall.Do()
		if err != nil {
			log.Printf("Error: unable to get channel %s: %v", ch.ID, err)
			success = false
			continue
		}
		channel := chListResp.Items[0]

		searchListCall := yService.Search.List("id,snippet").
			ChannelId(ch.ChannelID).MaxResults(int64(ch.Retain)).Order("date").Type("video")
		searchListResponse, err := searchListCall.Do()
		if err != nil {
			log.Printf("Error: unable to get latest videos in channel %s: %v", ch.ChannelID, err)
			success = false
			continue
		}

		rssCh := rssChannel{
			Title:       channel.Snippet.Title,
			Author:      channel.Snippet.CustomUrl,
			Description: channel.Snippet.Description,
			Image: rssImage{
				URL:   channel.Snippet.Thumbnails.High.Url,
				Link:  youtubeChannelURI + channel.Id,
				Title: channel.Snippet.Title,
			},
			Language: "en-us",
			Link:     youtubeChannelURI + channel.Id,
			Items:    []rssItem{},
		}

		for _, item := range searchListResponse.Items {

			videoListCall := yService.Videos.List("id,snippet,contentDetails").Id(item.Id.VideoId)
			videoListResponse, err := videoListCall.Do()
			if err != nil {
				log.Printf("Error: unable to get video detail %s => %s: %v",
					ch.ChannelID, item.Id.VideoId, err)
				success = false
				continue
			}

			video := videoListResponse.Items[0]

			fn, length, err := processVideo(video.Id)
			if err != nil {
				log.Printf("Error: unable to process video %s => %s: %v",
					ch.ChannelID, item.Id.VideoId, err)
				success = false
				continue
			}

			pubDate, err := formatPubDate(video.Snippet.PublishedAt)
			if err != nil {
				log.Printf("Error: unable to process video %s => %s: %v",
					ch.ChannelID, item.Id.VideoId, err)
				success = false
				continue
			}

			rssItem := rssItem{
				Title:       video.Snippet.Title,
				Description: video.Snippet.Description,
				PubDate:     pubDate,
				Enclosure: rssEnclosure{
					URL:    cfg.BaseURL + "/" + fn,
					Type:   "audio/mpeg",
					Length: length,
				},
				Duration: formatDuration(video.ContentDetails.Duration),
				GUID:     item.Id.VideoId,
			}

			rssCh.Items = append(rssCh.Items, rssItem)
		}

		err = saveFeedXML(ch.ID, rssCh)
		if err != nil {
			log.Printf("Error: to save channel feed XML %s: %v", ch.ChannelID, err)
			success = false
			continue
		}

		success = true

		log.Printf("Channel %s fetch sleeping for %d seconds", ch.ID, ch.RefreshInterval)
		time.Sleep(time.Duration(ch.RefreshInterval) * time.Second)
	}
}

func processVideo(videoID string) (string, int64, error) {
	log.Printf("Processing Video ID: %s", videoID)

	h := sha1.New()
	io.WriteString(h, videoID)
	videoHash := fmt.Sprintf("%x", h.Sum(nil))

	mp3FileName := fmt.Sprintf("%s.mp3", videoHash)
	mp3FullPath := filepath.Join(cfg.DataDir, mp3FileName)

	info, err := os.Stat(mp3FullPath)
	if !os.IsNotExist(err) {
		return mp3FileName, info.Size(), nil
	}

	videoURI := youtubeVideoURI + videoID
	videoFullPath := filepath.Join(cfg.DataDir, fmt.Sprintf("%s.%%(ext)s", videoHash))

	youtubeDlCmd := exec.Command(
		cfg.Exec.Youtubedl,
		"-o", videoFullPath,
		"-x", "--audio-format", "mp3",
		videoURI)
	_, err = youtubeDlCmd.CombinedOutput()
	if err != nil {
		return "", 0, err
	}

	info, err = os.Stat(mp3FullPath)
	if os.IsNotExist(err) {
		return mp3FileName, 0, err
	}
	return mp3FileName, info.Size(), nil
}

func saveFeedXML(ID string, rssCh rssChannel) error {
	rssRt := rssRoot{
		Version:     "2.0",
		XmlnsItunes: "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Channel:     rssCh,
	}
	b, err := xml.MarshalIndent(rssRt, "", "  ")
	if err != nil {
		return err
	}
	finalXML := xml.Header + string(b)
	finalXMLPath := filepath.Join(cfg.DataDir, fmt.Sprintf("feed-%s.xml", ID))
	return ioutil.WriteFile(finalXMLPath, []byte(finalXML), 0644)
}
