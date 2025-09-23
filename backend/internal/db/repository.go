package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	CheckUserExistsByEmail(ctx context.Context, email string) error
	CheckUserExistsById(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateDocument(ctx context.Context, document *Document) error
	GetDocumentByID(ctx context.Context, documentID string) (*Document, error)
	UpdateDocument(ctx context.Context, document *Document) error
	DeleteDocument(ctx context.Context, documentID string) error
	ListDocumentsByUserID(ctx context.Context, userID string) ([]*Document, error)
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
	panic("unimplemented")
}

func (r *repository) DeleteDocument(ctx context.Context, documentID string) error {
	panic("unimplemented")
}
