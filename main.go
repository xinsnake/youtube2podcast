package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/xinsnake/youtube2podcast/config"
	"google.golang.org/api/googleapi/transport"
	youtube "google.golang.org/api/youtube/v3"
)

const envConfigPath = "Y2P_CONFIG_PATH"

var cfg config.Config
var yService *youtube.Service

func main() {
	readConfigFiles()
	testDirectoryPermission()
	testDependencyApplications()
	prepareYouTubeService()
	go serveStaticContent()
	go fireOffFetchLoop()

	for {
		time.Sleep(3600 * time.Second)
	}
}

func readConfigFiles() {
	log.Printf("Reading config files")

	var err error
	configPath := os.Getenv(envConfigPath)
	if configPath == "" {
		log.Fatalf("Config file path (%s) cannot be empty", envConfigPath)
	}
	cfg, err = config.Parse(configPath)
	if err != nil {
		log.Fatalf("Unable to parse config file: %v", err)
	}
}

func testDirectoryPermission() {
	log.Printf("Test directory permission")

	testFile := filepath.Join(cfg.DataDir, "test")
	f, err := os.OpenFile(testFile, os.O_CREATE|os.O_RDWR, 0644)
	defer f.Close()
	if err != nil {
		log.Fatalf("Folder permission test failed (create): %s: %v", testFile, err)
	}
	err = os.Remove(testFile)
	if err != nil {
		log.Fatalf("Folder permission test failed (delete): %s: %v", testFile, err)
	}
}

func testDependencyApplications() {
	log.Printf("Test dependency applications")

	ffmpegCmd := exec.Command(cfg.Exec.Ffmpeg, "-version")
	output, err := ffmpegCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Unable to find ffmpeg: %s: %v", output, err)
	}

	youtubeDlCmd := exec.Command(cfg.Exec.Youtubedl, "--version")
	output, err = youtubeDlCmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Unable to find youtube-dl: %s: %v", output, err)
	}
}

func serveStaticContent() {
	log.Printf("Start serving static assets")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "index.html" {
			indexFile := filepath.Join(cfg.StaticDir, "index.html")
			tmpl, err := template.ParseFiles(indexFile)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Unable to parse template: %v", err)))
				return
			}
			feeds, err := getCurrentFeeds()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Unable to parse feeds: %v", err)))
				return
			}
			var buf bytes.Buffer
			err = tmpl.Execute(&buf, struct {
				Config config.Config
				Feeds  []rssRoot
			}{
				cfg,
				feeds,
			})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Unable to execute template: %v", err)))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(buf.Bytes())
			return
		}
		if r.URL.Path == "/favicon.ico" || r.URL.Path == "/logo.png" {
			staticPath := filepath.Join(cfg.StaticDir, r.URL.Path)
			http.ServeFile(w, r, staticPath)
			return
		}
		path := filepath.Join(cfg.DataDir, r.URL.Path)
		if f, err := os.Stat(path); err == nil && !f.IsDir() {
			http.ServeFile(w, r, path)
			return
		}
		http.NotFound(w, r)
	})

	err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), nil)
	if err != nil {
		log.Fatalf("Unable to listen to port %s: %v", cfg.Port, err)
	}
}

func prepareYouTubeService() {
	log.Printf("Preparing YouTube service")

	var err error
	yClient := &http.Client{Transport: &transport.APIKey{Key: cfg.GoogleAPIKey}}
	yService, err = youtube.New(yClient)
	if err != nil {
		log.Fatalf("Error creating YouTube service: %v", err)
	}
	if _, err = yService.Activities.List("id").
		ChannelId(cfg.Channels[0].ChannelID).MaxResults(1).Do(); err != nil {
		log.Fatalf("Error getting a working YouTube service: %v", err)
	}
}
