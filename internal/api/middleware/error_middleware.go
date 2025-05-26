package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"

	"devopstoolkit/youtube-automation/internal/service"
)

// ErrorResponse represents the structure of an error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// ErrorHandler is middleware that handles errors from API endpoints
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the stack trace
				log.Printf("PANIC RECOVERED: %v\n%s", err, debug.Stack())
				
				// Respond with 500 Internal Server Error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{
					Status:  http.StatusInternalServerError,
					Message: "An unexpected error occurred",
				})
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// APIError represents an API error with a status code and message
type APIError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e APIError) Error() string {
	return e.Message
}

// IsAPIError checks if an error is an APIError and returns the APIError and true if it is
func IsAPIError(err error) (APIError, bool) {
	apiErr, ok := err.(APIError)
	return apiErr, ok
}

// HandleAPIError handles API errors and sends appropriate responses
func HandleAPIError(w http.ResponseWriter, err error) {
	// Check if it's one of our known service errors
	switch {
	case err == service.ErrVideoNotFound:
		writeJSONError(w, http.StatusNotFound, "Video not found")
	case err == service.ErrInvalidRequest:
		writeJSONError(w, http.StatusBadRequest, "Invalid request")
	default:
		// Check if it's a custom API error
		if apiErr, ok := IsAPIError(err); ok {
			writeJSONError(w, apiErr.StatusCode, apiErr.Message)
			return
		}
		
		// Default to internal server error
		log.Printf("Internal error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Status:  statusCode,
		Message: message,
	})
}