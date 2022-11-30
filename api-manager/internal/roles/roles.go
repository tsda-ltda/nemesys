package roles

type Role = uint8

const (
	Unknown Role = iota
	Viewer
	TeamsManager
	Admin
	Master
	Invalid
)

func ValidateRole(role Role) bool {
	return role > Unknown && role < Invalid
}
