package db

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Email       string    `json:"email" db:"email"`
	Password    string    `json:"-" db:"password"`
	PhoneNumber *string   `json:"phoneNumber,omitempty" db:"phone_number"`
	Name        string    `json:"name" db:"name"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type Document struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"userId" db:"user_id"`
	Name           string    `json:"name" db:"name"`
	Description    *string   `json:"description,omitempty" db:"description"`
	Identifier     *string   `json:"identifier,omitempty" db:"identifier"`
	ExpirationDate time.Time `json:"expirationDate" db:"expiration_date"`
	Timezone       string    `json:"timezone" db:"timezone"`
	AttachmentURL  *string   `json:"attachmentUrl,omitempty" db:"attachment_url"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

type ReminderInterval struct {
	ID         int    `json:"id" db:"id"`
	Label      string `json:"label" db:"label"`
	DaysBefore int    `json:"daysBefore" db:"days_before"`
	IdLabel    string `json:"idLabel" db:"id_label"`
}

type DocumentReminder struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	DocumentID         string     `json:"documentId" db:"document_id"`
	ReminderIntervalID int        `json:"reminderIntervalId" db:"reminder_interval_id"`
	Enabled            bool       `json:"enabled" db:"enabled"`
	SentAt             *time.Time `json:"sentAt,omitempty" db:"sent_at"`
}

type NotificationLog struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	UserID             string    `json:"userId" db:"user_id"`
	DocumentID         string    `json:"documentId" db:"document_id"`
	ReminderIntervalID int       `json:"reminderIntervalId" db:"reminder_interval_id"`
	Channel            string    `json:"channel" db:"channel"`
	Status             string    `json:"status" db:"status"`
	Response           []byte    `json:"response" db:"response"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
}
