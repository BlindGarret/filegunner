package filegunner

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type MailRequest struct {
	From        string
	To          string
	Bcc         *string
	Subject     string
	Template    string
	Variables   *string
	ServiceID   string
	Attachments []string
}

func SendMailRequest(mailReq MailRequest, service MailgunService) error {
	var bs bytes.Buffer
	w := multipart.NewWriter(&bs)

	err := w.WriteField("from", mailReq.From)
	if err != nil {
		return err
	}

	err = w.WriteField("to", mailReq.To)
	if err != nil {
		return err
	}

	if mailReq.Bcc != nil {
		err = w.WriteField("bcc", *mailReq.Bcc)
		if err != nil {
			return err
		}
	}

	err = w.WriteField("subject", mailReq.Subject)
	if err != nil {
		return err
	}

	err = w.WriteField("template", mailReq.Template)
	if err != nil {
		return err
	}

	if mailReq.Variables != nil {
		err = w.WriteField("h:X-Mailgun-Variables", *mailReq.Variables)
		if err != nil {
			return err
		}
	}

	if mailReq.Attachments != nil {
		for _, path := range mailReq.Attachments {
			attachmentWriter, err := w.CreateFormFile("attachment", filepath.Base(path))
			if err != nil {
				return err
			}
			err = writeFileToAttachment(path, attachmentWriter)
			if err != nil {
				return err
			}
		}
	}

	err = w.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", service.Url, &bs)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.SetBasicAuth("api", service.ApiKey)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send email with status code: %d", resp.StatusCode)
	}

	return nil
}

func writeFileToAttachment(path string, writer io.Writer) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fileContents, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	_, err = writer.Write(fileContents)
	return err
}
