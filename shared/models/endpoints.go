package models

type AlarmEndpoint struct {
	Id int32 `json:"id" validate:"-"`
	// Name is the endpoint name.
	Name string `json:"name" validate:"required,max=50"`
	// URL is the endpoint URL.
	URL string `json:"url" validate:"required,max=255"`
	// Headers is the request headers.
	Headers []EndpointHeader `json:"headers" validate:"max=20"`
}
type EndpointHeader struct {
	Header string `json:"header"`
	Value  string `json:"value"`
}
