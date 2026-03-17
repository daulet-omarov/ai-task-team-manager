package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/daulet-omarov/ai-task-team-manager/internal/logger"
	"go.uber.org/zap"
)

type Mailer struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func New(host, port, username, password, from string) *Mailer {
	return &Mailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (m *Mailer) SendVerificationEmail(toEmail, verificationURL string) error {
	subject := "Verify your email address"
	body := fmt.Sprintf(`Hello,

Please verify your email address by clicking the link below:

%s

This link expires in 24 hours.

If you did not register, please ignore this email.
`, verificationURL)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		m.from, toEmail, subject, body,
	)

	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	addr := fmt.Sprintf("%s:%s", m.host, m.port)

	err := smtp.SendMail(addr, auth, m.from, []string{toEmail}, []byte(msg))

	if err != nil {
		logger.Log.Error("error while send verification email", zap.Error(err))
	}

	return err
}

func (m *Mailer) SendPasswordResetEmail(toEmail, resetURL string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`Hello,

You requested to reset your password. Click the link below:

%s

This link expires in 15 minutes.

If you did not request this, please ignore this email.
Your password will not be changed.
`, resetURL)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		m.from, toEmail, subject, body,
	)

	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	addr := fmt.Sprintf("%s:%s", m.host, m.port)

	err := smtp.SendMail(addr, auth, m.from, []string{toEmail}, []byte(msg))

	if err != nil {
		logger.Log.Error("error while send password reset email", zap.Error(err))
	}

	return err
}
