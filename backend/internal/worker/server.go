package worker

import (
	"context"
	"encoding/json"
	"log"
	"xpired/internal/config"
	"xpired/internal/db"

	"github.com/hibiken/asynq"
)

const (
	TaskSendReminder = "send_reminder"
)

func NewServer(cfg *config.Config) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)
}

func NewMux(repo db.Repository) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendReminder, func(ctx context.Context, t *asynq.Task) error {
		var payload struct {
			UserID     string `json:"user_id"`
			DocumentID string `json:"document_id"`
			IntervalID int    `json:"interval_id"`
		}

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		userEmail, err := repo.GetUserEmail(ctx, payload.UserID)
		if err != nil {
			return err
		}

		doc, err := repo.GetDocumentByID(ctx, payload.DocumentID)
		if err != nil {
			return err
		}

		email := EmailTemplate(userEmail, doc.Name, doc.ExpirationDate.Format("January 2, 2006"))
		if err := SendEmail(userEmail, "Document Expiration Reminder", email); err != nil {
			log.Printf("Failed to send email to %s: %v", userEmail, err)
		}

		userPhone, _ := repo.GetUserPhoneNumber(ctx, payload.UserID)
		if userPhone != "" {
			sms := SMSMessage(doc.Name, doc.ExpirationDate.Format("January 2, 2006"))
			_ = SendSMS(userPhone, sms)
		}

		log.Printf("Reminder: User %s should be notified about document %s (interval=%d)",
			userEmail, doc.Name, payload.IntervalID)

		return nil
	})
	return mux
}
