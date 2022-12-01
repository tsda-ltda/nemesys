package influxdb

import (
	"strconv"
	"time"
)

func ParseDuration(influxDuration string) (duration time.Duration, err error) {
	if len(influxDuration) < 2 {
		return 0, ErrInvalidDuration
	}

	var magnitude string

	lastIndex := len(influxDuration) - 1
	timeValue, err := strconv.ParseInt(influxDuration[:lastIndex], 0, 64)
	if err != nil {
		timeValue, err = strconv.ParseInt(influxDuration[:lastIndex-1], 0, 64)
		if err != nil {
			return 0, ErrInvalidDuration
		}
		magnitude = influxDuration[lastIndex-1:]
	} else {
		magnitude = influxDuration[lastIndex:]
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
