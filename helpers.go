package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func handleStartupError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

func formatPubDate(isoDate string) string {
	t, err := time.Parse(time.RFC3339Nano, isoDate)
	if err != nil {
		log.Printf("Unable to convert pubDate: %v", err)
		return ""
	}
	return t.Format(time.RFC1123)
}

func formatDuration(isoDuration string) string {
	var err error

	re := regexp.MustCompile("P(\\d+D)?T(\\d+H)?(\\d+M)?(\\d+S)?")
	matches := re.FindAllStringSubmatch(isoDuration, -1)

	dayStr := matches[0][1]
	hourStr := matches[0][2]
	minuteStr := matches[0][3]
	secondStr := matches[0][4]

	var day, hour, minute, second int

	if dayStr != "" {
		day, err = strconv.Atoi(strings.Replace(dayStr, "D", "", -1))
		if err != nil {
			log.Printf("Unable to convert Day: %v", err)
			return ""
		}
	}
	if hourStr != "" {
		hour, err = strconv.Atoi(strings.Replace(hourStr, "H", "", -1))
		if err != nil {
			log.Printf("Unable to convert Hour: %v", err)
			return ""
		}
	}
	if minuteStr != "" {
		minute, err = strconv.Atoi(strings.Replace(minuteStr, "M", "", -1))
		if err != nil {
			log.Printf("Unable to convert Minute: %v", err)
			return ""
		}
	}
	if secondStr != "" {
		second, err = strconv.Atoi(strings.Replace(secondStr, "S", "", -1))
		if err != nil {
			log.Printf("Unable to convert Second: %v", err)
			return ""
		}
	}

	hour = hour + day*24

	return fmt.Sprintf("%0.2d:%0.2d:%0.2d", hour, minute, second)
}
