package types

type ContainerType byte

const (
	CTUnknown ContainerType = iota
	CTBasic
	CTSNMPv2c
	CTFlexLegacy
)

func IsNonFlex(ct ContainerType) bool {
	return ct == CTFlexLegacy
}
