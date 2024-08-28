package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// EmailSender interface defines the methods that a sender should implement
type EmailSender interface {
	SendOTP(to, otp string) error
	SendPasswordResetToken(to, token string) error
}

type Sender struct {
	dialer *gomail.Dialer
}

// Ensure Sender implements EmailSender
var _ EmailSender = (*Sender)(nil)

func NewSender(host string, port int, username, password string) *Sender {
	return &Sender{
		dialer: gomail.NewDialer(host, port, username, password),
	}
}

func (s *Sender) SendOTP(to, otp string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "noreply@yourdomain.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Your OTP for Real Madrid Shop")
	m.SetBody("text/plain", fmt.Sprintf("Your OTP for sign up is: %s", otp))

	return s.dialer.DialAndSend(m)
}

func (s *Sender) SendPasswordResetToken(to, token string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "noreply@yourdomain.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Password Reset Token for Real Madrid Shop")
	m.SetBody("text/plain", fmt.Sprintf("Your password reset token is: %s", token))
	return s.dialer.DialAndSend(m)
}
