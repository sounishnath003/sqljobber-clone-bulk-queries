package handlers

import "time"

type ApiResponse struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     error       `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewApiResponse sends an APIResponse with the proper standard format required throughout the project
func NewApiResponse(status int, err error, message string, data interface{}) *ApiResponse {
	defaultResp := &ApiResponse{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
	}

	if err != nil {
		defaultResp.Error = err
	} else {
		defaultResp.Data = data
	}

	return defaultResp
}
