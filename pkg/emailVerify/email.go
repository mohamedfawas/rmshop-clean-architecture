package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type Sender struct {
	dialer *gomail.Dialer
}

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
	m.SetBody("text/plain", fmt.Sprintf("Your OTP is: %s", otp))

	return s.dialer.DialAndSend(m)
}
