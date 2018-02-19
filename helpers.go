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
	re := regexp.MustCompile("P(\\d+D)?T(\\d+H)?(\\d+M)?(\\d+S)?")
	matches := re.FindAllStringSubmatch(isoDuration, -1)

	dayStr, hourStr, minuteStr, secondStr :=
		matches[0][1], matches[0][2], matches[0][3], matches[0][4]
	day, hour, minute, second :=
		takeTimePart(dayStr, "D", "Day"),
		takeTimePart(hourStr, "H", "Hour"),
		takeTimePart(minuteStr, "M", "Minute"),
		takeTimePart(secondStr, "S", "Second")

	return fmt.Sprintf("%0.2d:%0.2d:%0.2d", hour+day*24, minute, second)
}

func takeTimePart(input string, tShort string, tLong string) int {
	if input == "" {
		return 0
	}
	output, err := strconv.Atoi(strings.Replace(input, tShort, "", -1))
	if err != nil {
		log.Printf("Unable to convert %s: %v", tLong, err)
		return 0
	}
	return output
}
