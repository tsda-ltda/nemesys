package influxdb

import (
	"strconv"
	"strings"
	"time"
)

func ParseDuration(influxDuration string) (duration time.Duration, err error) {
	// get magnitude
	var magnitude string
	_, err = strconv.Atoi(string(influxDuration[len(influxDuration)-2]))
	if err == nil {
		magnitude = string(influxDuration[len(influxDuration)-1])
	} else {
		magnitude = string(influxDuration[len(influxDuration)-2]) + string(influxDuration[len(influxDuration)-1])
	}

	timeValueString, _, _ := strings.Cut(influxDuration, magnitude)
	timeValue, err := strconv.Atoi(timeValueString)
	if err != nil {
		return duration, ErrInvalidDuration
	}

	switch magnitude {
	case "ns":
		duration = time.Nanosecond * time.Duration(timeValue)
	case "us":
		duration = time.Microsecond * time.Duration(timeValue)
	case "ms":
		duration = time.Millisecond * time.Duration(timeValue)
	case "s":
		duration = time.Second * time.Duration(timeValue)
	case "m":
		duration = time.Minute * time.Duration(timeValue)
	case "h":
		duration = time.Hour * time.Duration(timeValue)
	case "d":
		duration = time.Hour * 24 * time.Duration(timeValue)
	case "y":
		duration = time.Hour * 24 * 365 * time.Duration(timeValue)
	default:
		return duration, ErrInvalidDuration
	}
	return duration, nil
}

func DurationFromSeconds(seconds int64) string {
	return strconv.FormatInt(seconds, 10) + "s"
}
