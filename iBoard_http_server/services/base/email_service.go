package base_services

import (
	"gopkg.in/gomail.v2"
)

type IEmailService interface {
	SendEmail(to []string, subject string, body string) error
}

type EmailService struct {
	emailClient *gomail.Dialer
}

func NewEmailService(emailClient *gomail.Dialer) IEmailService {
	return &EmailService{
		emailClient: emailClient,
	}
}

func (s *EmailService) SendEmail(to []string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.emailClient.Username)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if err := s.emailClient.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
