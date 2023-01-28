package model

type Response struct {
	// The HTTP status code of the response.
	Status uint16 `json:"status" example:"200"`
	// The error message, if any, associated with the response.
	Message string `json:"message,omitempty" example:"Error message."`
	// Data returned by the request, if any.
	Data interface{} `json:"data,omitempty" example:""`
}
