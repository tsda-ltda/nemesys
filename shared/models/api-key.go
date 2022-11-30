package models

type APIKeyInfo struct {
	// Id is the API Key indentifier.
	Id int16 `json:"id" validate:"-"`
	// UserId is the user identifier.
	UserId int32 `json:"-" validate:"-"`
	// Descr is the API Key description.
	Descr string `json:"descr" validate:"required"`
	// TTL is the time-to-live of the API Key in hours.
	TTL int32 `json:"ttl" validate:"min=0"`
	// CreatedAt is the date of creation in UNIX format.
	CreatedAt int64 `json:"created-at" validate:"-"`
}

type APIkey struct {
	APIKey string `json:"api-key"`
}
