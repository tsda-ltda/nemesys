package uuid

import "github.com/segmentio/ksuid"

// New returns a new unique id.
func New() (uuid string, err error) {
	return ksuid.New().String(), nil
}
