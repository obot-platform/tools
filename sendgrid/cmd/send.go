package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func Send(ctx context.Context, apiKey, from, fromName, to, subject, textBody, htmlBody string) (string, error) {
	apiKey = strings.TrimSpace(apiKey)
	from = strings.TrimSpace(from)
	fromName = strings.TrimSpace(fromName)
	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)
	textBody = strings.TrimSpace(textBody)
	htmlBody = strings.TrimSpace(htmlBody)

	if apiKey == "" {
		return "", fmt.Errorf("api key is required")
	}

	if from == "" {
		return "", fmt.Errorf("from email is required")
	}

	if fromName == "" {
		// Default to "Obot"
		fromName = "Obot"
	}

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

	personalization := mail.NewPersonalization()
	for _, email := range validEmails {
		personalization.AddTos(mail.NewEmail("", email))
	}

	// Create email message
	message := mail.NewV3Mail()
	message.SetFrom(mail.NewEmail(fromName, from))
	message.Subject = subject
	message.AddPersonalizations(personalization)
	if textBody != "" {
		message.AddContent(mail.NewContent("text/plain", textBody))
	}
	if htmlBody != "" {
		message.AddContent(mail.NewContent("text/html", htmlBody))
	}

	client := sendgrid.NewSendClient(apiKey)
	response, err := client.SendWithContext(ctx, message)
	if err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("failed to send email: status code %d, body: %s", response.StatusCode, response.Body)
	}

	return fmt.Sprintf("email sent successfully with status code: %d", response.StatusCode), nil
}
