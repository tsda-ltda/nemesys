package roles

type Role = uint8

const (
	Viewer       Role = 1
	TeamsManager Role = 2
	Admin        Role = 3
	Master       Role = 4
)
