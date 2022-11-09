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
	case MTInt8:
		return int8(n), nil
	case MTInt16:
		return int16(n), nil
	case MTInt32:
		return int32(n), nil
	case MTInt64:
		return int64(n), nil
	case MTFloat64:
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
	case MTInt8:
		return int8(f), nil
	case MTInt16:
		return int16(f), nil
	case MTInt32:
		return int32(f), nil
	case MTInt64:
		return int64(f), nil
	case MTFloat64:
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
	case MTInt8:
		return int8(i), nil
	case MTInt16:
		return int16(i), nil
	case MTInt32:
		return int32(i), nil
	case MTInt64:
		return int64(i), nil
	case MTFloat64:
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
	case MTInt8:
		return int8(math.Round(f)), err
	case MTInt16:
		return int16(math.Round(f)), err
	case MTInt32:
		return int32(math.Round(f)), err
	case MTInt64:
		return int64(math.Round(f)), err
	case MTFloat64:
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
