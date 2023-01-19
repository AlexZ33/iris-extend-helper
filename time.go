package iris_extend_helper

import (
	"log"
	"math"
	"strings"
	"time"
)

func Timestamp(milliseconds float64) time.Time {
	seconds := math.Floor(milliseconds / 1000)
	nanoseconds := (milliseconds - seconds*1000) * 1000000
	return time.Unix(int64(seconds), int64(nanoseconds))
}

func ParseTimestamp(value string) (time.Time, bool) {
	layout := ""
	length := len(value)
	switch {
	case strings.Count(value, "/") == 2:
		if length == 10 {
			layout = "2006/01/02"
		} else if length == 19 {
			layout = "2006/01/02 15:04:05"
		} else if length < 10 {
			layout = "2006/1/2"
		} else {
			layout = "2006/1/2 15:04:05"
		}
	case length > 0 && length < 20:
		str := "2006-01-02 15:04:05"
		layout = string(str[0:length])
		value = strings.Replace(value, "T", " ", 1)
	case length == 24:
		layout = "2006-01-02T15:04:05.999Z"
	case length == 25:
		layout = "2006-01-02T15:04:05-07:00"
	case length >= 26 && length <= 35:
		layout = "2006-01-02T15:04:05." + strings.Repeat("9", length-26) + "-07:00"
		value = strings.Replace(value, " ", "+", 1)
	}
	if layout != "" {
		if timestamp, err := time.Parse(layout, value); err != nil {
			log.Println(err)
		} else {
			return timestamp, true
		}
	}
	return time.Unix(0, 0), false
}

func StringifyTime(t time.Time) string {
	layout := "2006-01-02T15:04:05.999999-07:00"
	return t.Format(layout)
}
