package types

import "errors"

type FlexLegacyPortType int16

const (
	FLPTUnknown FlexLegacyPortType = iota
	FLPTSystem
	FLPTPing
	FLPTMetering
	FLPTStatus
	FLPTCommand
	FLPTSNMPOutput
	FLPTSNMPVirtual
	FLPTModbusRTU
	FLPTModbusTCPTable
	FLPTScript
)

var (
	ErrUnsupportedFlexLegacyPortType = errors.New("unsupported Flex Legacy port type")
)

func ParseFlexPortTypeFromOID(oid string) (t FlexLegacyPortType, err error) {
	if len(oid) < 28 {
		return t, ErrUnsupportedFlexLegacyPortType
	}
	if oid[:18] != ".1.3.6.1.4.1.31957" {
		return t, ErrUnsupportedFlexLegacyPortType
	}

	oidType := oid[21:22]
	switch oidType {
	// flexSystem
	case "1":
		return FLPTSystem, nil

	// flexMonitoring
	case "3":
		switch oid[23:24] {
		case "1":
			return FLPTPing, nil
		case "2":
			return FLPTMetering, nil
		case "3":
			return FLPTStatus, nil
		case "4":
			return FLPTCommand, nil
		case "5":
			return FLPTSNMPOutput, nil
		case "6":
			return FLPTSNMPVirtual, nil
		case "7":
			return FLPTModbusRTU, nil
		case "8":
			return FLPTModbusTCPTable, nil
		default:
			return t, ErrUnsupportedFlexLegacyPortType
		}

	// flexSysScript
	case "6":
		return t, ErrUnsupportedFlexLegacyPortType
	default:
		return t, ErrUnsupportedFlexLegacyPortType
	}
}

func ParseFlexPortType(v string) (t FlexLegacyPortType, err error) {
	switch v {
	case "metering":
		return FLPTMetering, nil
	case "status":
		return FLPTStatus, nil
	case "command":
		return FLPTCommand, nil
	case "virtual":
		return FLPTSNMPVirtual, nil
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
	case FLPTSNMPVirtual:
		return "virtual", nil
	default:
		return s, ErrUnsupportedFlexLegacyPortType
	}
}
