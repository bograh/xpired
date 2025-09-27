package auth

import "time"

type ErrorResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Status    int       `json:"status"`
}
