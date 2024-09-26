package filegunner

import (
	"fmt"
	"strings"
)

type DryRunMailer struct {
	fileCreateFunc CreateFileFunc
}

func NewDryRunMailer(fileCreateFunc CreateFileFunc) *DryRunMailer {
	return &DryRunMailer{
		fileCreateFunc: fileCreateFunc,
	}
}

func (m *DryRunMailer) Send(req MailRequest, fileName string) error {
	logFileName := strings.TrimSuffix(fileName, ".maildata.json") + ".log"
	f, err := m.fileCreateFunc(logFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte("#################################\n"))
	if err != nil {
		return err
	}

	_, err = f.Write([]byte(fmt.Sprintf("To: %s\n", req.To)))
	if err != nil {
		return err
	}

	_, err = f.Write([]byte(fmt.Sprintf("From: %s\n", req.From)))
	if err != nil {
		return err
	}

	if req.Bcc != nil {
		_, err = f.Write([]byte(fmt.Sprintf("BCC: %s\n", *req.Bcc)))
		if err != nil {
			return err
		}
	}
	_, err = f.Write([]byte(fmt.Sprintf("Subject Line: %s\n", req.Subject)))
	if err != nil {
		return err
	}

	_, err = f.Write([]byte(fmt.Sprintf("Template Name: %s\n", req.Template)))
	if err != nil {
		return err
	}

	if req.Variables != nil {
		_, err = f.Write([]byte(fmt.Sprintf("Variables: %s\n", *req.Variables)))
		if err != nil {
			return err
		}
	}

	if req.Attachments != nil {
		for _, attachment := range req.Attachments {
			_, err = f.Write([]byte(fmt.Sprintf("Attachment: %s as %s\n", attachment.FilePath, attachment.AttachmentName)))
			if err != nil {
				return err
			}
		}
	}

	_, err = f.Write([]byte(fmt.Sprintf("Mailgun Service: %s\n", req.ServiceID)))
	return err
}
