package types

type MetricType byte

const (
	MTUnknown MetricType = iota
	MTInt8
	MTInt16
	MTInt32
	MTInt64
	MTString
	MTBool
	MTInvalid
)

// Parse Asn1BER type to Metric type.
func ParseAsn1BER(b byte) MetricType {
	switch b {
	case 0x01:
		return MTBool
	case 0x02:
		return MTInt32
	case 0x03:
		return MTString
	case 0x04:
		return MTString
	case 0x06:
		return MTString
	case 0x07:
		return MTString
	case 0x40:
		return MTString
	case 0x41:
		return MTInt32
	case 0x42:
		return MTInt32
	case 0x43:
		return MTInt32
	case 0x44:
		return MTInt32
	case 0x45:
		return MTString
	case 0x46:
		return MTInt64
	case 0x47:
		return MTInt32
	case 0x78:
		return MTInt32
	case 0x79:
		return MTInt64
	default:
		return MTInvalid
	}
}
