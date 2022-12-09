package types

type ContainerType byte

const (
	CTUnknown ContainerType = iota
	CTBasic
	CTSNMPv2c
	CTFlexLegacy
)

func IsNonFlex(ct ContainerType) bool {
	return ct != CTFlexLegacy
}

func StringfyContainerType(t ContainerType) string {
	switch t {
	case CTBasic:
		return "Basic"
	case CTFlexLegacy:
		return "Flex Legacy"
	case CTSNMPv2c:
		return "SNMPv2c"
	default:
		return "Unknown"
	}
}
