package filegunner

import "time"

type Mailer interface {
	Send(req MailRequest, fileName string, now time.Time) error
}
