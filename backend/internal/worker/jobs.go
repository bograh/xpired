package worker

import (
	"encoding/json"
	"log"
	"time"

	"xpired/internal/config"
	"xpired/internal/db"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

var client *asynq.Client

func InitQueue(cfg *config.Config) {
	client = asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Redis.Addr})
	client.Ping()
	log.Println("Asynq client initialized")
}

func enqueueDelayedTask(taskType string, payload map[string]interface{}, runAt time.Time) error {
	data, _ := json.Marshal(payload)
	task := asynq.NewTask(taskType, data)

	_, err := client.Enqueue(task, asynq.ProcessAt(runAt))
	return err
}

func ScheduleReminders(doc db.Document, userID uuid.UUID, enabledIntervals []db.ReminderInterval) {
	for _, interval := range enabledIntervals {
		reminderTime := doc.ExpirationDate.AddDate(0, 0, -interval.DaysBefore)

		if reminderTime.Before(time.Now()) {
			log.Printf("Skipping past reminder for doc %s (interval %d)", doc.ID.String(), interval.ID)
			continue
		}

		reminderTimeUTC := reminderTime.UTC()
		payload := map[string]interface{}{
			"user_id":     userID.String(),
			"document_id": doc.ID.String(),
			"interval_id": interval.ID,
		}

		if err := enqueueDelayedTask("send_reminder", payload, reminderTimeUTC); err != nil {
			log.Printf("Failed to enqueue reminder for doc %s: %v", doc.ID.String(), err)
		}
	}
}
