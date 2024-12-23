package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func Send(ctx context.Context, to, subject, textBody, htmlBody string) (string, error) {
	// Retrieve and trim environment variables
	sendGridAPIKey := strings.TrimSpace(os.Getenv("SENDGRID_API_KEY"))
	if sendGridAPIKey == "" {
		return "", fmt.Errorf("SENDGRID_API_KEY is not set")
	}

	fromEmail := strings.TrimSpace(os.Getenv("OBOT_NO_REPLY_EMAIL"))
	if fromEmail == "" {
		return "", fmt.Errorf("OBOT_NO_REPLY_EMAIL is not set")
	}

	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)
	textBody = strings.TrimSpace(textBody)
	htmlBody = strings.TrimSpace(htmlBody)

	if to == "" {
		return "", fmt.Errorf("recipient email (to) is required")
	}

	// Split comma-delimited recipients and trim each email
	toEmails := strings.Split(to, ",")
	for i := range toEmails {
		toEmails[i] = strings.TrimSpace(toEmails[i])
	}

	// Remove any empty emails after trimming
	var validEmails []string
	for _, email := range toEmails {
		if email != "" {
			validEmails = append(validEmails, email)
		}
	}

	if len(validEmails) < 1 {
		return "", fmt.Errorf("no valid recipient emails provided")
	}

	if subject == "" {
		return "", fmt.Errorf("subject is required")
	}

	if textBody == "" && htmlBody == "" {
		return "", fmt.Errorf("either textBody or htmlBody is required")
	}

	from := mail.NewEmail("Obot", fromEmail)
	personalization := mail.NewPersonalization()

	for _, email := range validEmails {
		personalization.AddTos(mail.NewEmail("", email))
	}

	// Create email message
	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = subject
	message.AddPersonalizations(personalization)
	if textBody != "" {
		message.AddContent(mail.NewContent("text/plain", textBody))
	}
	if htmlBody != "" {
		message.AddContent(mail.NewContent("text/html", htmlBody))
	}

	client := sendgrid.NewSendClient(sendGridAPIKey)
	response, err := client.SendWithContext(ctx, message)
	if err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("failed to send email: status code %d, body: %s", response.StatusCode, response.Body)
	}

	return fmt.Sprintf("email sent successfully with status code: %d", response.StatusCode), nil
}
