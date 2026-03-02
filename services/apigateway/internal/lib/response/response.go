package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, SuccessResponse{Data: data})
}

func Error(w http.ResponseWriter, status int, errType, message string) {
	JSON(w, status, ErrorResponse{
		Error:   errType,
		Message: message,
	})
}
