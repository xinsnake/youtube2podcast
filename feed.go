package main

import (
	"encoding/xml"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

type rssRoot struct {
	XMLName     xml.Name   `xml:"rss"`
	Version     string     `xml:"version,attr"`
	XmlnsItunes string     `xml:"xmlns:itunes,attr"`
	Channel     rssChannel `xml:"channel"`
}

type rssChannel struct {
	ID          string    `xml:"id,attr"`
	Title       string    `xml:"title"`
	Author      string    `xml:"author"`
	Description string    `xml:"description"`
	Image       rssImage  `xml:"image"`
	Language    string    `xml:"language"`
	Link        string    `xml:"link"`
	Items       []rssItem `xml:"item"`
}

type rssImage struct {
	URL   string `xml:"url"`
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

type rssItem struct {
	Title       string       `xml:"title"`
	Description string       `xml:"description"`
	PubDate     string       `xml:"pubDate"`
	Enclosure   rssEnclosure `xml:"enclosure"`
	Duration    string       `xml:"itunes:duration"`
	GUID        string       `xml:"guid"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length int64  `xml:"length,attr"`
}

func getCurrentFeeds() ([]rssRoot, error) {
	var results []rssRoot
	files, err := ioutil.ReadDir(cfg.DataDir)
	if err != nil {
		return results, err
	}
	re := regexp.MustCompile("^feed-[a-z0-9]+\\.xml$")
	for _, file := range files {
		fileName := file.Name()
		if !re.MatchString(fileName) {
			continue
		}
		xmlString, err := ioutil.ReadFile(filepath.Join(cfg.DataDir, file.Name()))
		if err != nil {
			return results, err
		}
		var xmlObj rssRoot
		err = xml.Unmarshal(xmlString, &xmlObj)
		if err != nil {
			return results, err
		}
		results = append(results, xmlObj)
	}
	return results, nil
}
