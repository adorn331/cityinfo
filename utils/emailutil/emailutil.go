package emailutil

import (
	"cityinfo/configs"
	"log"
	"net/smtp"
	"strings"
	"sync"
)

var (
	// global err email sender
	ErrNotifier EmailHandler
	// onceInit guarantee initialize logger only once
	onceInit sync.Once
)

type EmailHandler struct {
	auth smtp.Auth
	to []string
	from string
	senderNickname string
	subject string
	contentType string
}

func init() {
	onceInit.Do(func() {
		// load the config and construct sender
		ErrNotifier.auth = smtp.PlainAuth("", configs.SMTP_USER, configs.SMTP_PWD, configs.SMTP_HOST)
		ErrNotifier.to = configs.GetErrEmailReciver()
		ErrNotifier.from = configs.EMAIL_FROM
		ErrNotifier.senderNickname = configs.EMAIL_FROM_NICKNAME
		ErrNotifier.subject = configs.EMAIL_SUBJECT
		ErrNotifier.contentType = configs.EMAIL_CONTENT_TYPE
	})
}

// implement io.Writer, so that logger can register ErrNotifier
// and send mail when highPriority err occur
func (s EmailHandler) Write(msg []byte) (int, error) {
	body := []byte("To: " + strings.Join(s.to, ",") +
		"\r\nFrom: " + s.senderNickname +
		"<" + s.from + ">" +
		"\r\nSubject: " + s.subject + "\r\n" +
		s.contentType + "\r\n\r\n" + string(msg))

	err := smtp.SendMail(configs.SMTP_HOST+":"+configs.SMTP_PORT, s.auth, s.from, s.to, body)
	if err != nil {
		log.Print("Send mail error.")
		return 0, err
	}
	return len(msg), nil
}