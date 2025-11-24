package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailSender is what usecases depend on.
type EmailSender interface {
	SendOTP(to, otp string) error
	SendPasswordResetToken(to, token string) error
}

// Sender implements EmailSender using net/smtp.
type Sender struct {
	host string
	port int

	username string
	password string

	fromEmail string
	fromName  string
	addr      string
	auth      smtp.Auth
}

// NewSender matches how main.go calls it.
func NewSender(host string, port int, username, password, fromEmail, fromName string) *Sender {
	// Trim to avoid hidden spaces/newlines from env/.env
	host = strings.TrimSpace(host)
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	fromEmail = strings.TrimSpace(fromEmail)
	fromName = strings.TrimSpace(fromName)

	if fromEmail == "" {
		fromEmail = username
	}
	if fromName == "" {
		fromName = "Real Madrid Shop"
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	// For dev with MailHog/localhost (no auth)
	var auth smtp.Auth
	if host != "mailhog" && host != "localhost" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	return &Sender{
		host:      host,
		port:      port,
		username:  username,
		password:  password,
		fromEmail: fromEmail,
		fromName:  fromName,
		addr:      addr,
		auth:      auth,
	}
}

func (s *Sender) buildMessage(to, subject, body string) []byte {
	from := s.fromEmail
	if s.fromName != "" {
		from = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(body)
	return buf.Bytes()
}

func (s *Sender) send(to, subject, body string) error {
	msg := s.buildMessage(to, subject, body)
	return smtp.SendMail(
		s.addr,
		s.auth,
		s.fromEmail,  // envelope from
		[]string{to}, // envelope to
		msg,
	)
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
