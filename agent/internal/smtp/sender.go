package smtp

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Sender struct {
	from   string
	addr   string
}

func NewSender(host string, port int, from string) *Sender {
	return &Sender{
		from: from,
		addr: fmt.Sprintf("%s:%d", host, port),
	}
}

func (s *Sender) SendErrorAlert(to, projectName, errorMessage, timestamp string) error {
	subject := fmt.Sprintf("[ERROR] Project: %s - %s", projectName, timestamp)

	body := fmt.Sprintf(`У вас ошибка в проекте %s:

Level: ERROR
Timestamp: %s

Message:
%s
`, projectName, timestamp, errorMessage)

	return s.Send(to, subject, body)
}

func (s *Sender) Send(to, subject, body string) error {
	msg := s.buildMessage(to, subject, body)

	err := smtp.SendMail(s.addr, nil, s.from, []string{to}, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}

func (s *Sender) buildMessage(to, subject, body string) string {
	var msg strings.Builder
	msg.WriteString("From: ")
	msg.WriteString(s.from)
	msg.WriteString("\nTo: ")
	msg.WriteString(to)
	msg.WriteString("\nSubject: ")
	msg.WriteString(subject)
	msg.WriteString("\n\n")
	msg.WriteString(body)

	return msg.String()
}