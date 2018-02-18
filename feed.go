package main

import "encoding/xml"

type rssRoot struct {
	XMLName     xml.Name   `xml:"rss"`
	Version     string     `xml:"version,attr"`
	XmlnsItunes string     `xml:"xmlns:itunes,attr"`
	Channel     rssChannel `xml:"channel"`
}

type rssChannel struct {
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
