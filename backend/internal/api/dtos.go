package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type UserRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	Name        string  `json:"name"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
}

type ErrorResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Status    int       `json:"status"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	Name        string  `json:"name"`
}

type DocumentRequest struct {
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	Identifier     *string   `json:"identifier,omitempty"`
	ExpirationDate time.Time `json:"expirationDate"`
	Timezone       string    `json:"timezone"`
	AttachmentURL  *string   `json:"attachmentUrl,omitempty"`
	Reminders      []string  `json:"reminders"`
}

type DocumentResponse struct {
	ID             string                     `json:"id"`
	UserID         string                     `json:"userId"`
	Name           string                     `json:"name"`
	Description    *string                    `json:"description,omitempty"`
	Identifier     *string                    `json:"identifier,omitempty"`
	ExpirationDate string                     `json:"expirationDate"`
	Timezone       string                     `json:"timezone"`
	AttachmentURL  *string                    `json:"attachmentUrl,omitempty"`
	Reminders      []ReminderIntervalResponse `json:"reminders"`
	CreatedAt      time.Time                  `json:"createdAt"`
	UpdatedAt      time.Time                  `json:"updatedAt"`
}

type ReminderIntervalResponse struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type ToggleDocumentReminderRequest struct {
	ReminderIntervalID string `json:"interval_id"`
	Enabled            bool   `json:"enabled"`
}

type DocumentReminderIntervalResponse struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
}

func NotFoundError(message string) ErrorResponse {
	var errResp ErrorResponse
	errResp.Message = message
	errResp.Status = http.StatusNotFound
	errResp.Timestamp = time.Now()
	return errResp
}

func BadRequestError(message string) ErrorResponse {
	var errResp ErrorResponse
	errResp.Message = message
	errResp.Status = http.StatusBadRequest
	errResp.Timestamp = time.Now()
	return errResp
}

func ConflictError(message string) ErrorResponse {
	var errResp ErrorResponse
	errResp.Message = message
	errResp.Status = http.StatusConflict
	errResp.Timestamp = time.Now()
	return errResp
}

func InternalServerError(message string) ErrorResponse {
	var errResp ErrorResponse
	errResp.Message = message
	errResp.Status = http.StatusInternalServerError
	errResp.Timestamp = time.Now()
	return errResp
}

func UnauthorizedError(message string) ErrorResponse {
	var errResp ErrorResponse
	errResp.Message = message
	errResp.Status = http.StatusUnauthorized
	errResp.Timestamp = time.Now()
	return errResp
}

func ForbiddenError(message string) ErrorResponse {
	var errResp ErrorResponse
	errResp.Message = message
	errResp.Status = http.StatusForbidden
	errResp.Timestamp = time.Now()
	return errResp
}

func WriteErrorResponse(w http.ResponseWriter, errResp ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errResp.Status)
	json.NewEncoder(w).Encode(errResp)
}
