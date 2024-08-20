package api

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SendResponse(w http.ResponseWriter, statusCode int, message string, data interface{}, err string) {
	response := Response{
		Status:  statusCode,
		Message: message,
		Data:    data,
		Error:   err,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
