package api

import (
	"time"
)

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
