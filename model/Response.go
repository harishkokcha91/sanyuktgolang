package model

type Response struct {
	// The HTTP status code of the response.
	StatusCode uint16 `json:"code" example:"200"`
	Status     bool   `json:"status" example:"true"`
	// The error message, if any, associated with the response.
	Message string `json:"message,omitempty" example:"Error message."`
	// Data returned by the request, if any.
	Data interface{} `json:"data,omitempty" example:""`
}
