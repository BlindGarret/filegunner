package filegunner

import (
	"fmt"
	"os"
	"strings"
)

func DryRunMail(mailReq MailRequest, fileName string) error {
	logFileName := strings.TrimSuffix(fileName, ".maildata.json") + ".log"

	f, err := os.Create(logFileName)
	if err != nil {
		return err
	}

	_, err = f.WriteString("#################################\n")
	if err != nil {
		return err
	}

	_, err = f.WriteString(fmt.Sprintf("To: %s\n", mailReq.To))
	if err != nil {
		return err
	}

	_, err = f.WriteString(fmt.Sprintf("From: %s\n", mailReq.From))
	if err != nil {
		return err
	}

	if mailReq.Bcc != nil {
		_, err = f.WriteString(fmt.Sprintf("BCC: %s\n", *mailReq.Bcc))
		if err != nil {
			return err
		}
	}
	_, err = f.WriteString(fmt.Sprintf("Subject Line: %s\n", mailReq.Subject))
	if err != nil {
		return err
	}

	_, err = f.WriteString(fmt.Sprintf("Template Name: %s\n", mailReq.Template))
	if err != nil {
		return err
	}

	if mailReq.Variables != nil {
		_, err = f.WriteString(fmt.Sprintf("Variables: %s\n", *mailReq.Variables))
		if err != nil {
			return err
		}
	}

	_, err = f.WriteString(fmt.Sprintf("Mailgun Service: %s\n", mailReq.ServiceID))
	return err
}
