package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	CheckUserExistsByEmail(ctx context.Context, email string) error
	CheckUserExistsById(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserEmail(ctx context.Context, userID string) (string, error)
	GetUserPhoneNumber(ctx context.Context, userID string) (string, error)
	CreateDocument(ctx context.Context, document *Document) error
	GetDocumentByID(ctx context.Context, documentID string) (*Document, error)
	UpdateDocument(ctx context.Context, document *Document) error
	DeleteDocument(ctx context.Context, documentID string) error
	ListDocumentsByUserID(ctx context.Context, userID string) ([]*Document, error)
	GetAllReminderIntervals(ctx context.Context) ([]*ReminderInterval, error)
	GetReminderIntervalsFromIdLabels(ctx context.Context, idLabels []string) ([]*ReminderInterval, error)
	GetReminderIntervalByID(ctx context.Context, id int) (*ReminderInterval, error)
	SetDocumentReminders(ctx context.Context, documentID string, reminder *DocumentReminder) error
	ToggleDocumentReminder(ctx context.Context, documentID string, reminderIntervalID int, enabled bool) error
	GetDocumentRemindersByDocumentID(ctx context.Context, documentID string) ([]*DocumentReminder, error)
}

type repository struct {
	db *DB
}

func NewRepository(db *DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, email, password, phone_number, name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`
	err := r.db.DB.QueryRow(
		query,
		user.ID,
		user.Email,
		user.Password,
		user.PhoneNumber,
		user.Name,
	).Scan(
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *repository) CheckUserExistsByEmail(ctx context.Context, email string) error {
	var userEmail string
	query := `SELECT id FROM users WHERE email = $1`
	err := r.db.DB.QueryRowContext(ctx, query, email).Scan(&userEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user does not exist")
		}
		return fmt.Errorf("failed to check user: %w", err)
	}
	return nil
}

func (r *repository) CheckUserExistsById(ctx context.Context, userID string) error {
	var id string
	query := `SELECT id FROM users WHERE id = $1`
	err := r.db.DB.QueryRowContext(ctx, query, userID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user does not exist")
		}
		return fmt.Errorf("failed to check user: %w", err)
	}
	return nil
}

func (r *repository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	query := `
		SELECT id, email, password, phone_number, name, created_at, updated_at FROM users WHERE id = $1
	`
	row := r.db.DB.QueryRow(query, userID)
	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.PhoneNumber,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password, phone_number, name, created_at, updated_at FROM users WHERE email = $1
	`
	row := r.db.DB.QueryRow(query, email)
	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.PhoneNumber,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

func (r *repository) GetUserEmail(ctx context.Context, userID string) (string, error) {
	var email string
	query := `SELECT email FROM users WHERE id = $1`
	err := r.db.DB.QueryRowContext(ctx, query, userID).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user does not exist")
		}
		return "", fmt.Errorf("failed to get user email: %w", err)
	}
	return email, nil
}

func (r *repository) GetUserPhoneNumber(ctx context.Context, userID string) (string, error) {
	var phoneNumber string
	query := `SELECT phone_number FROM users WHERE id = $1`
	err := r.db.DB.QueryRowContext(ctx, query, userID).Scan(&phoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user does not exist")
		}
		return "", fmt.Errorf("failed to get user phone number: %w", err)
	}
	return phoneNumber, nil
}

func (r *repository) CreateDocument(ctx context.Context, document *Document) error {
	query := `
		INSERT INTO documents (id, user_id, name, description, identifier, expiration_date, timezone, attachment_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`
	err := r.db.DB.QueryRow(
		query,
		document.ID,
		document.UserID,
		document.Name,
		document.Description,
		document.Identifier,
		document.ExpirationDate,
		document.Timezone,
		document.AttachmentURL,
	).Scan(
		&document.CreatedAt, &document.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

func (r *repository) ListDocumentsByUserID(ctx context.Context, userID string) ([]*Document, error) {
	query := `
		SELECT id, user_id, name, description, identifier, expiration_date, timezone, attachment_url, created_at, updated_at
		FROM documents
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var documents []*Document
	for rows.Next() {
		var doc Document
		err := rows.Scan(
			&doc.ID,
			&doc.UserID,
			&doc.Name,
			&doc.Description,
			&doc.Identifier,
			&doc.ExpirationDate,
			&doc.Timezone,
			&doc.AttachmentURL,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, &doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return documents, nil
}

func (r *repository) GetDocumentByID(ctx context.Context, documentID string) (*Document, error) {
	query := `
		SELECT id, user_id, name, description, identifier, expiration_date, timezone, attachment_url, created_at, updated_at
		FROM documents
		WHERE id = $1
	`
	row := r.db.DB.QueryRowContext(ctx, query, documentID)
	var doc Document
	err := row.Scan(
		&doc.ID,
		&doc.UserID,
		&doc.Name,
		&doc.Description,
		&doc.Identifier,
		&doc.ExpirationDate,
		&doc.Timezone,
		&doc.AttachmentURL,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return &doc, nil
}

func (r *repository) UpdateDocument(ctx context.Context, document *Document) error {
	query := `
		UPDATE documents
		SET name = $1, description = $2, identifier = $3, expiration_date = $4, timezone = $5, attachment_url = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`
	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		document.Name,
		document.Description,
		document.Identifier,
		document.ExpirationDate,
		document.Timezone,
		document.AttachmentURL,
		document.ID,
	).Scan(&document.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("document not found")
		}
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

func (r *repository) DeleteDocument(ctx context.Context, documentID string) error {
	query := `
		DELETE FROM documents
		WHERE id = $1
	`
	result, err := r.db.DB.ExecContext(ctx, query, documentID)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

func (r *repository) GetAllReminderIntervals(ctx context.Context) ([]*ReminderInterval, error) {
	query := `
		SELECT id, label, days_before, id_label
		FROM reminder_intervals
	`
	rows, err := r.db.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminder intervals: %w", err)
	}
	defer rows.Close()

	var intervals []*ReminderInterval
	for rows.Next() {
		var interval ReminderInterval
		err := rows.Scan(
			&interval.ID,
			&interval.Label,
			&interval.DaysBefore,
			&interval.IdLabel,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder interval: %w", err)
		}
		intervals = append(intervals, &interval)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return intervals, nil
}

func (r *repository) GetReminderIntervalsFromIdLabels(ctx context.Context, idLabels []string) ([]*ReminderInterval, error) {
	query := `
		SELECT id, label, days_before, id_label
		FROM reminder_intervals
		WHERE id_label = ANY($1)
	`
	rows, err := r.db.DB.QueryContext(ctx, query, pq.Array(idLabels))
	if err != nil {
		return nil, fmt.Errorf("failed to get reminder intervals: %w", err)
	}
	defer rows.Close()

	var intervals []*ReminderInterval
	for rows.Next() {
		var interval ReminderInterval
		err := rows.Scan(
			&interval.ID,
			&interval.Label,
			&interval.DaysBefore,
			&interval.IdLabel,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder interval: %w", err)
		}
		intervals = append(intervals, &interval)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return intervals, nil
}

func (r *repository) GetReminderIntervalByID(ctx context.Context, id int) (*ReminderInterval, error) {
	query := `
		SELECT id, label, days_before, id_label
		FROM reminder_intervals
		WHERE id = $1
	`
	row := r.db.DB.QueryRowContext(ctx, query, id)
	var interval ReminderInterval
	err := row.Scan(
		&interval.ID,
		&interval.Label,
		&interval.DaysBefore,
		&interval.IdLabel,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reminder interval not found")
		}
		return nil, fmt.Errorf("failed to get reminder interval: %w", err)
	}
	return &interval, nil
}

func (r *repository) SetDocumentReminders(ctx context.Context, documentID string, reminder *DocumentReminder) error {
	query := `
		INSERT INTO document_reminders (id, document_id, reminder_interval_id, enabled)
		VALUES ($1, $2, $3, $4)
		RETURNING sent_at
	`
	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		reminder.ID,
		documentID,
		reminder.ReminderIntervalID,
		reminder.Enabled,
	).Scan(&reminder.SentAt)

	if err != nil {
		return fmt.Errorf("failed to create document reminder: %w", err)
	}

	return nil
}

func (r *repository) ToggleDocumentReminder(ctx context.Context, documentID string, reminderIntervalID int, enabled bool) error {
	query := `
		UPDATE document_reminders
		SET enabled = $1, sent_at = NULL
		WHERE document_id = $2 AND reminder_interval_id = $3
	`
	result, err := r.db.DB.ExecContext(ctx, query, enabled, documentID, reminderIntervalID)
	if err != nil {
		return fmt.Errorf("failed to toggle document reminder: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("document reminder not found")
	}

	return nil
}

func (r *repository) GetDocumentRemindersByDocumentID(ctx context.Context, documentID string) ([]*DocumentReminder, error) {
	query := `
		SELECT id, document_id, reminder_interval_id, enabled, sent_at
		FROM document_reminders
		WHERE document_id = $1
	`
	rows, err := r.db.DB.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document reminders: %w", err)
	}
	defer rows.Close()

	var reminders []*DocumentReminder
	for rows.Next() {
		var reminder DocumentReminder
		err := rows.Scan(
			&reminder.ID,
			&reminder.DocumentID,
			&reminder.ReminderIntervalID,
			&reminder.Enabled,
			&reminder.SentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document reminder: %w", err)
		}
		reminders = append(reminders, &reminder)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return reminders, nil
}
