package filegunner

type Mailer interface {
	Send(req MailRequest, fileName string) error
}
