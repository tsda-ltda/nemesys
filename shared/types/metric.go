package types

type MetricType byte

const (
	MTUnknown MetricType = iota
	MTInt
	MTFloat
	MTString
	MTBool
	MTInvalid
)

// ValidateMetricType validates the metric type.
func ValidateMetricType(t MetricType) bool {
	return t > MTUnknown && t < MTInvalid
}

// Parse Asn1BER type to Metric type.
func ParseAsn1BER(b byte) MetricType {
	switch b {
	case 0x01:
		return MTBool
	case 0x02:
		return MTInt
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
		return MTInt
	case 0x42:
		return MTInt
	case 0x43:
		return MTInt
	case 0x44:
		return MTInt
	case 0x45:
		return MTString
	case 0x46:
		return MTInt
	case 0x47:
		return MTInt
	case 0x78:
		return MTInt
	case 0x79:
		return MTInt
	default:
		return MTInvalid
	}
}
