package types

import "errors"

type FlexLegacyPortType int16

const (
	FLPTUnknown FlexLegacyPortType = iota
	FLPTMetering
	FLPTStatus
	FLPTCommand
	FLPTVirtual
)

var ErrUnsupportedFlexLegacyPortType = errors.New("unsupported Flex Legacy port type")

func ParseFlexPortType(v string) (t FlexLegacyPortType, err error) {
	switch v {
	case "metering":
		return FLPTMetering, nil
	case "status":
		return FLPTStatus, nil
	case "command":
		return FLPTCommand, nil
	case "virtual":
		return FLPTVirtual, nil
	default:
		return FLPTUnknown, ErrUnsupportedFlexLegacyPortType
	}
}

func FlexPortTypeToString(t FlexLegacyPortType) (s string, err error) {
	switch t {
	case FLPTMetering:
		return "metering", nil
	case FLPTStatus:
		return "status", nil
	case FLPTCommand:
		return "command", nil
	case FLPTVirtual:
		return "virtual", nil
	default:
		return s, ErrUnsupportedFlexLegacyPortType
	}
}
