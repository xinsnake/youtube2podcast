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
	"regexp"
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
	if ch.Retain > 50 {
		log.Printf("Warning: channel %s retain greater than 50 is not supported yet, set to 50", ch.ID)
		ch.Retain = 50
	}

	re := regexp.MustCompile("^[a-z]+$")
	if !re.MatchString(ch.ID) {
		log.Printf("Warning: channel %s contains none a-z charactors, you may encounter erros", ch.ID)
	}

	success := true

	for {
		if !success {
			log.Printf("Detecte previous failure for channel %s, waiting for 1 interval", ch.ID)
			time.Sleep(time.Duration(ch.RefreshInterval) * time.Second)
		}

		log.Printf("Start to process channel %s", ch.ID)

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
			ID:          ch.ID,
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

		mp3s := make(map[string]bool)

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

			fn, length, err := processVideo(ch.ID, video.Id)
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
			mp3s[fn] = true
		}

		err = saveFeedXML(ch.ID, rssCh)
		if err != nil {
			log.Printf("Error: unable to save channel feed XML %s: %v", ch.ChannelID, err)
			success = false
			continue
		}

		err = cleanUp(ch.ID, mp3s)
		if err != nil {
			log.Printf("Error: unable to clean up unused files %s: %v", ch.ChannelID, err)
			success = false
			continue
		}

		success = true

		log.Printf("Channel %s fetch sleeping for %d seconds", ch.ID, ch.RefreshInterval)
		time.Sleep(time.Duration(ch.RefreshInterval) * time.Second)
	}
}

func processVideo(chID, videoID string) (string, int64, error) {
	log.Printf("Processing Video ID: %s", videoID)

	h := sha1.New()
	io.WriteString(h, videoID)
	videoHash := fmt.Sprintf("%x", h.Sum(nil))

	mp3FileName := fmt.Sprintf("%s-%s.mp3", chID, videoHash)
	mp3FullPath := filepath.Join(cfg.DataDir, mp3FileName)

	info, err := os.Stat(mp3FullPath)
	if !os.IsNotExist(err) {
		return mp3FileName, info.Size(), nil
	}

	videoURI := youtubeVideoURI + videoID
	videoFullPath := filepath.Join(cfg.DataDir, fmt.Sprintf("%s-%s.%%(ext)s", chID, videoHash))

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

func saveFeedXML(chID string, rssCh rssChannel) error {
	log.Printf("Saving feed XML file for channel %s", chID)

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
	finalXMLPath := filepath.Join(cfg.DataDir, fmt.Sprintf("feed-%s.xml", chID))
	return ioutil.WriteFile(finalXMLPath, []byte(finalXML), 0644)
}

func cleanUp(chID string, mp3s map[string]bool) error {
	log.Printf("Cleaning unused mp3 files for channel %s", chID)

	files, err := ioutil.ReadDir(cfg.DataDir)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(fmt.Sprintf("^%s-[a-f0-9]+\\.mp3$", chID))
	for _, file := range files {
		fileName := file.Name()
		if !re.MatchString(fileName) || mp3s[fileName] {
			continue
		}
		log.Printf("Removing unused MP3 file %s", fileName)
		err = os.Remove(filepath.Join(cfg.DataDir, fileName))
		if err != nil {
			return err
		}
	}
	return nil
}
