package sender

import (
	"crypto/tls"
	"fmt"
	"log/slog"

	"gopkg.in/gomail.v2"
)

// Sender delivers emails via SMTP using gomail.
type Sender struct {
	log    *slog.Logger
	dialer *gomail.Dialer
	from   string
}

// New creates an SMTP sender with the given credentials.
func New(log *slog.Logger, host string, port int, username, password, from string) *Sender {
	d := gomail.NewDialer(host, port, username, password)
	d.TLSConfig = &tls.Config{
		ServerName: host,
	}

	return &Sender{
		log:    log,
		dialer: d,
		from:   from,
	}
}

// Send delivers an email to the given recipient.
func (s *Sender) Send(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("smtp sender: %w", err)
	}

	s.log.Info("email sent",
		slog.String("to", to),
		slog.String("subject", subject),
	)

	return nil
}
