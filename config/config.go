package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	BaseURL      string `json:"baseUrl"`
	Port         string `json:"port"`
	StaticDir    string `json:"staticDir"`
	DataDir      string `json:"dataDir"`
	GoogleAPIKey string `json:"googleApiKey"`
	Exec         struct {
		Ffmpeg    string `json:"ffmpeg"`
		Youtubedl string `json:"youtubedl"`
	} `json:"exec"`
	Channels []Channel `json:"channels"`
}

type Channel struct {
	ID              string `json:"id"`
	ChannelID       string `json:"channelId"`
	Retain          int    `json:"retain"`
	RefreshInterval int    `json:"refreshInterval"`
}

func Parse(path string) (c Config, err error) {
	var b []byte
	if b, err = ioutil.ReadFile(path); err != nil {
		return c, err
	}
	if err = json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	return c, nil
}
