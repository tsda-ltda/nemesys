package types

import (
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrInvalidParseType = errors.New("invalid parse type")
)

func ParseValue(v any, t MetricType) (any, error) {
	vk := reflect.TypeOf(v).Kind()
	switch vk {
	case reflect.String:
		return ParseString(v.(string), t)
	case reflect.Bool:
		return ParseBool(v.(bool), t)
	case reflect.Uint8:
		return ParseInt(int(v.(uint8)), t)
	case reflect.Uint16:
		return ParseInt(int(v.(uint16)), t)
	case reflect.Uint32:
		return ParseInt(int(v.(uint32)), t)
	case reflect.Uint64:
		return ParseInt(int(v.(uint64)), t)
	case reflect.Uint:
		return ParseInt(int(v.(uint)), t)
	case reflect.Int:
		return ParseInt(v.(int), t)
	case reflect.Int8:
		return ParseInt(int(v.(int8)), t)
	case reflect.Int16:
		return ParseInt(int(v.(int16)), t)
	case reflect.Int32:
		return ParseInt(int(v.(int32)), t)
	case reflect.Int64:
		return ParseInt(int(v.(int64)), t)
	case reflect.Float32:
		return ParseFloat64(float64(v.(float32)), t)
	case reflect.Float64:
		return ParseFloat64(v.(float64), t)
	default:
		return nil, ErrInvalidParseType
	}
}

func ParseBool(b bool, t MetricType) (any, error) {
	n := 0
	if b {
		n = 1
	}
	switch t {
	case MTBool:
		return b, nil
	case MTString:
		if b {
			return "true", nil
		} else {
			return "false", nil
		}
	case MTInt:
		return int64(n), nil
	case MTFloat:
		return float64(n), nil
	default:
		return nil, ErrInvalidParseType
	}
}

func ParseFloat64(f float64, t MetricType) (any, error) {
	switch t {
	case MTBool:
		if f < 1 {
			return false, nil
		} else {
			return true, nil
		}
	case MTString:
		return strconv.FormatFloat(f, 'f', -1, 32), nil
	case MTInt:
		return int64(f), nil
	case MTFloat:
		return f, nil
	default:
		return nil, ErrInvalidParseType
	}
}

func ParseInt(i int, t MetricType) (any, error) {
	switch t {
	case MTBool:
		if i == 0 {
			return false, nil
		} else {
			return true, nil
		}
	case MTString:
		return strconv.FormatInt(int64(i), 10), nil
	case MTInt:
		return int64(i), nil
	case MTFloat:
		return float64(i), nil
	default:
		return nil, ErrInvalidParseType
	}
}

func ParseString(s string, t MetricType) (any, error) {
	s = strings.Replace(s, ",", ".", 1)
	f, err := strconv.ParseFloat(s, 64)
	switch t {
	case MTString:
		return s, nil
	case MTInt:
		return int64(math.Round(f)), err
	case MTFloat:
		return f, err
	case MTBool:
		lower := strings.ToLower(s)
		switch lower {
		case "false":
			return false, nil
		case "true":
			return true, nil
		default:
			if err != nil {
				return nil, ErrInvalidParseType
			}
			if f > 0 {
				return true, nil
			} else {
				return false, nil
			}
		}
	default:
		return nil, ErrInvalidParseType
	}
}
