package services

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail sends an email with the specified subject and body.
func SendEmail(body, receiverEmail, ccEmail, subject string) {
	smtpHost := os.Getenv("SES_HOST")
	smtpPort := 587
	sender := "sender@youremail.com"
	username := os.Getenv("SES_USERNAME")
	password := os.Getenv("SES_PASSWORD")

	// Construct the message
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nCc: %s\r\nSubject: %s\r\n\r\n%s", sender, receiverEmail, ccEmail, subject, body)
	auth := smtp.PlainAuth("", username, password, smtpHost)

	if err := smtp.SendMail(fmt.Sprintf("%s:%d", smtpHost, smtpPort), auth, sender, []string{receiverEmail}, []byte(msg)); err != nil {
		fmt.Printf("Error sending email: %v\n", err)
	} else {
		fmt.Println("Email sent successfully to:", receiverEmail)
	}
}
