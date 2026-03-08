package mail

import (
	"fmt"
	"log"
	"net/smtp"
	"github.com/clarkk/go-util/logs"
)

type SMTP struct {
	auth	smtp.Auth
	addr	string
	logger	*log.Logger
}

func NewSMTP(host string, port int, user, pass, log_path string) *SMTP {
	logger, err := logs.New(log_path, 1024)
	if err != nil {
		log.Fatal(err)
	}
	
	return &SMTP{
		auth:	smtp.PlainAuth("", user, pass, host),
		addr:	fmt.Sprintf("%s:%d", host, port),
		logger:	logger,
	}
}

func (s *SMTP) Send(mail *Mail) error {
	s.logger.Printf("[SEND] %s (Subject: %s)", mail.to_email, mail.subject)
	
	msg, err := mail.Message()
	if err != nil {
		s.logger.Printf("[ERROR] Generate source: %v", err)
		return err
	}
	
	if err = smtp.SendMail(s.addr, s.auth, mail.from_email, []string{mail.to_email}, []byte(msg)); err != nil {
		s.logger.Printf("[ERROR] SMTP for %s: %v", mail.to_email, err)
		return err
	}
	
	s.logger.Printf("[OK] Relay for %s", mail.to_email)
	return nil
}