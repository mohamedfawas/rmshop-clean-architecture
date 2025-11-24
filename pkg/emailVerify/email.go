package email

import (
	"fmt"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailSender is what the usecases depend on.
type EmailSender interface {
	SendOTP(to, otp string) error
	SendPasswordResetToken(to, token string) error
}

// Sender implements EmailSender using SendGrid HTTP API.
type Sender struct {
	client    *sendgrid.Client
	fromEmail string
	fromName  string
}

// NewSender matches the existing main.go signature.
// We ignore host/port/username and treat password as the SendGrid API key.
func NewSender(_ string, _ int, _ string, password, fromEmail, fromName string) *Sender {
	apiKey := strings.TrimSpace(password)

	fromEmail = strings.TrimSpace(fromEmail)
	if fromEmail == "" {
		fromEmail = "no-reply@example.com" // fallback; override via env in prod
	}

	fromName = strings.TrimSpace(fromName)
	if fromName == "" {
		fromName = "Real Madrid Shop"
	}

	client := sendgrid.NewSendClient(apiKey)

	return &Sender{
		client:    client,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

func (s *Sender) send(to, subject, body string) error {
	from := mail.NewEmail(s.fromName, s.fromEmail)
	toEmail := mail.NewEmail("", to)

	message := mail.NewSingleEmail(from, subject, toEmail, body, body)

	_, err := s.client.Send(message)
	if err != nil {
		return fmt.Errorf("sendgrid send failed: %w", err)
	}
	return nil
}

func (s *Sender) SendOTP(to, otp string) error {
	subject := "Your OTP for Real Madrid Shop"
	body := fmt.Sprintf("Your OTP for sign up is: %s", otp)
	return s.send(to, subject, body)
}

func (s *Sender) SendPasswordResetToken(to, token string) error {
	subject := "Password Reset Token for Real Madrid Shop"
	body := fmt.Sprintf("Your password reset token is: %s", token)
	return s.send(to, subject, body)
}
