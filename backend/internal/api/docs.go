package api

import (
	"time"
	"xpired/internal/db"
)

type DocumentRequest struct {
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	Identifier     *string   `json:"identifier,omitempty"`
	ExpirationDate time.Time `json:"expirationDate"`
	Timezone       string    `json:"timezone"`
	AttachmentURL  *string   `json:"attachmentUrl,omitempty"`
	Reminders      []int     `json:"reminders"`
}

type DocumentResponse struct {
	ID             string                `json:"id"`
	UserID         string                `json:"userId"`
	Name           string                `json:"name"`
	Description    *string               `json:"description,omitempty"`
	Identifier     *string               `json:"identifier,omitempty"`
	ExpirationDate string                `json:"expirationDate"`
	Timezone       string                `json:"timezone"`
	AttachmentURL  *string               `json:"attachmentUrl,omitempty"`
	Reminders      []db.DocumentReminder `json:"reminders"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
}
