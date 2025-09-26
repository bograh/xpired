package worker

import "log"

func SendEmail(to, subject, body string) error {
	// Simulate sending email
	log.Printf("Sending email to: %s, Subject: %s", to, subject)
	return nil
}

func SendSMS(to, message string) error {
	// Simulate sending SMS
	log.Printf("Sending SMS to: %s, Message: %s", to, message)
	return nil
}
