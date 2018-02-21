package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func formatPubDate(isoDate string) (string, error) {
	t, err := time.Parse(time.RFC3339Nano, isoDate)
	if err != nil {
		return "", err
	}
	return t.Format(time.RFC1123), nil
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
		return 0
	}
	return output
}
