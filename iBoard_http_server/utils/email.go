package utils

import (
	"gopkg.in/gomail.v2"
)

var EmailClient *gomail.Dialer

func InitEmail(smtpHost string, smtpPort int, smtpUser string, smtpPass string) *gomail.Dialer {
	EmailClient = gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	return EmailClient
}
