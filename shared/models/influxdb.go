package models

type QueryPoint struct {
	// Timestamp is point timestamp in UNIX format
	Timestamp int64 `json:"timestamp"`
	// Value is the point value.
	Value any `json:"value"`
}
