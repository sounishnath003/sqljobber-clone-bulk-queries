package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ApiResponse struct {
	ID        uuid.UUID   `json:"id"`
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewApiResponse sends an APIResponse with the proper standard format required throughout the project
func NewApiResponse(status int, err error, message string, data interface{}) *ApiResponse {
	defaultResp := &ApiResponse{
		ID:        uuid.New(),
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
	}

	if err != nil {
		defaultResp.Error = err.Error()
	} else {
		defaultResp.Data = data
	}

	return defaultResp
}

func SendApiResponse(w http.ResponseWriter, status int, data *ApiResponse) error {
	w.Header().Add("Content-Type", "application/json, charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
